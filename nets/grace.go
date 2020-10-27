package nets

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

/*
 * Grace restart app interface
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 * base on the code of https://gravitational.com/blog/golang-ssh-bastion-graceful-restarts/
 * - support http
 * - support general tcp
 *
 * use `kill -SIGUSR2 pid` command to kill old version process
 */

 //internal macro defines
 const (
 	GraceListenerEnv = "LISTENER"
 	GraceSignalSize = 32
 	GraceHttpTimeOut = 10 //xx seconds
 )

 //tcp kind
 const (
 	TcpKindGen = iota
 	TcpKindSocket
 )

 //listener info
 type listener struct {
	Addr     string `json:"addr"`
	FD       int    `json:"fd"`
	Filename string `json:"filename"`
 }

 //grace info
 type Grace struct {
 	port int `tcp port`
 	address string `tcp address`
 	tcpKind int `tcp kind`
 	listener net.Listener `net listener`
 	httpServer *http.Server `http service instance`
 	socketCallback func(net.Listener) `callback for socket`
 	httpHandlers map[string]bool `http handlers map`
 	isAttach bool `is attached listener`
 	hasForked bool `current process has forked or not`
 	sync.Mutex
 }

 //construct
func NewGrace(port, tcpKind int) *Grace {
	//init address
	address := fmt.Sprintf(":%d", port)

	//self init
	this := &Grace{
		port:port,
		address:address,
		tcpKind:tcpKind,
		httpHandlers:make(map[string]bool),
	}

	//init tcp listener
	// Create (or import) a net.Listener and start a goroutine that runs
	// a TCP/HTTP server on that net.Listener.
	listener, err := this.createOrImportListener(address)
	if err != nil {
		tips := fmt.Sprintf("Create or import listener failed, err:%v", err.Error())
		log.Println(tips)
		panic(tips)
	}

	//sync listener
	this.listener = listener

	return this
}

//////////
//api
/////////

//get tcp listener
func (g *Grace) GetListener() net.Listener {
	return g.listener
}

//check current process has forked or not
func (g *Grace) HasForked() bool {
	return g.hasForked
}

//set is attached
func (g *Grace) SetAttached(val bool) {
	g.isAttach = val
}

//bind socket interface
func (g *Grace) BindSocket(cb func(net.Listener)) bool {
	if cb == nil || g.socketCallback != nil {
		return false
	}

	//set socket call back
	g.socketCallback = cb

	//spawn process for socket service
	go g.socketCallback(g.listener)

	return true
}

//bind http interface
func (g *Grace) BindHttp() bool {
	if g.httpServer != nil {
		return false
	}

	//init http server
	g.httpServer = &http.Server {
		Addr:g.address,//listen address and port
		ReadTimeout:time.Second * GraceHttpTimeOut,
		WriteTimeout:time.Second * GraceHttpTimeOut,
	}

	return true
}

//register http handler
//handler can't be pointer type!!!
func (g *Grace) RegisterHandler(handler http.Handler) bool {
	if handler == nil || g.httpServer == nil {
		return false
	}
	//set http handler
	g.httpServer.Handler = handler
	return true
}

//register http handler func
func (g *Grace) RegisterHttpFunc(reqUrl string, handler func(w http.ResponseWriter, r *http.Request)) bool {
	if reqUrl == "" {
		return false
	}

	//check is exists or not
	_, ok := g.httpHandlers[reqUrl]
	if ok {
		return true
	}

	//register new handler
	http.HandleFunc(reqUrl, handler)

	//add into cache
	g.Lock()
	g.httpHandlers[reqUrl] = true
	g.Unlock()

	return true
}

//quit
func (g *Grace) Quit() {
	if !g.isAttach {
		g.httpServer.Close()
		g.httpServer = nil
	}
}

//start
func (g *Grace) Start() error {
	if !g.isAttach {
		//check and start http service
		if g.httpServer != nil {
			go g.httpServer.Serve(g.listener)
		}
	}

	//try watch signals
	return g.waitForSignals()
}

/////////////////
//private func
////////////////

//close http service
func (g *Grace) closeHttpService() bool {
	// Create a context that will expire in 5 seconds and use this as a
	// timeout to Shutdown.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	//
	// Return any errors during shutdown.
	g.httpServer.Shutdown(ctx)
	return true
}

//wait for signals which be catch
func (g *Grace) waitForSignals() error {
	var (
		signalCh = make(chan os.Signal, GraceSignalSize)
	)
	//register signal
	signal.Notify(signalCh,
		syscall.SIGHUP,
		syscall.SIGUSR2,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
		os.Interrupt,
		os.Kill,
	)

	for {
		select {
		case s := <-signalCh:
			log.Printf("%v signal received.\n", s)
			switch s {
			case os.Interrupt:
				fallthrough
			case os.Kill:
				break
			case syscall.SIGHUP:
				// Fork a child process.
				p, err := g.forkChild(g.address, g.listener)
				if err != nil {
					log.Printf("Unable to fork child: %v.\n", err)
					continue
				}

				//mark forked
				g.hasForked = true
				log.Printf("Forked child %v.\n", p.Pid)

				//check http service
				if g.httpServer != nil {
					go g.closeHttpService()
				}

			case syscall.SIGUSR2:
				// Fork a child process.
				p, err := g.forkChild(g.address, g.listener)
				if err != nil {
					log.Printf("Unable to fork child: %v.\n", err)
					continue
				}

				//mark forked
				g.hasForked = true

				// Print the PID of the forked process and keep waiting for more signals.
				fmt.Printf("Forked child %v.\n", p.Pid)
			case syscall.SIGINT, syscall.SIGQUIT:
				//check http service
				if g.httpServer != nil {
					go g.closeHttpService()
				}
			}
		}
	}
	return nil
}

//fork new child process
func (g *Grace) forkChild(addr string, ln net.Listener) (*os.Process, error) {
	// Get the file descriptor for the listener and marshal the metadata to pass
	// to the child in the environment.
	lnFile, err := g.getListenerFile(ln)
	log.Println("lnFile:", lnFile)
	if err != nil {
		return nil, err
	}
	defer lnFile.Close()
	l := listener{
		Addr:     addr,
		FD:       3,
		Filename: lnFile.Name(),
	}
	listenerEnv, err := json.Marshal(l)
	if err != nil {
		return nil, err
	}

	// Pass stdin, stdout, and stderr along with the listener to the child.
	files := []*os.File{
		os.Stdin,
		os.Stdout,
		os.Stderr,
		lnFile,
	}

	// Get current environment and add in the listener to it.
	tempListenerEnv := fmt.Sprintf("%s=%s", GraceListenerEnv, string(listenerEnv))
	environment := append(os.Environ(), tempListenerEnv)

	// Get current process name and directory.
	execName, err := os.Executable()
	if err != nil {
		return nil, err
	}
	execDir := filepath.Dir(execName)
	log.Println("execName:", execName, ", execDir:", execDir)

	// Spawn child process.
	p, err := os.StartProcess(execName, []string{execName}, &os.ProcAttr{
		Dir:   execDir,
		Env:   environment,
		Files: files,
		Sys:   &syscall.SysProcAttr{},
	})
	if err != nil {
		return nil, err
	}

	return p, nil
}

//get tcp listener file
func (g *Grace) getListenerFile(ln net.Listener) (*os.File, error) {
	switch t := ln.(type) {
	case *net.TCPListener:
		return t.File()
	case *net.UnixListener:
		return t.File()
	}
	return nil, fmt.Errorf("unsupported listener: %T", ln)
}

//create and import tcp listener
func (g *Grace) createOrImportListener(addr string) (net.Listener, error) {
	// Try and import a listener for addr. If it's found, use it.
	ln, err := g.importListener(addr)
	if err == nil {
		log.Printf("Imported listener file descriptor for %v.\n", addr)
		return ln, nil
	}

	// No listener was imported, that means this process has to create one.
	ln, err = g.createListener(addr)
	if err != nil {
		return nil, err
	}
	//log.Printf("Created listener file descriptor for %v.\n", addr)
	return ln, nil
}

//create tcp listener
func (g *Grace) createListener(addr string) (net.Listener, error) {
	var (
		listener net.Listener
		err error
	)

	switch g.tcpKind {
	case TcpKindSocket:
		//socket kind tcp
		tcpAddress, err := net.ResolveTCPAddr("tcp", addr)
		if err != nil {
			log.Fatalln("net.ResolveTCPddr failed.", err)
			return nil, err
		}

		//begin listen tcp
		listener, err = net.ListenTCP("tcp", tcpAddress)
	case TcpKindGen:
		//general
		listener, err = net.Listen("tcp", addr)
	default:
		err = errors.New(fmt.Sprintf("unsupported tcp kind %d", g.tcpKind))
	}

	//ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	return listener, nil
}

//import tcp listener
func (g *Grace) importListener(addr string) (net.Listener, error) {
	// Extract the encoded listener metadata from the environment.
	listenerEnv := os.Getenv(GraceListenerEnv)
	if listenerEnv == "" {
		return nil, fmt.Errorf("unable to find LISTENER environment variable")
	}

	// Unmarshal the listener metadata.
	var l listener
	err := json.Unmarshal([]byte(listenerEnv), &l)
	if err != nil {
		return nil, err
	}
	if l.Addr != addr {
		return nil, fmt.Errorf("unable to find listener for %v", addr)
	}

	// The file has already been passed to this process, extract the file
	// descriptor and name from the metadata to rebuild/find the *os.File for
	// the listener.
	listenerFile := os.NewFile(uintptr(l.FD), l.Filename)
	if listenerFile == nil {
		return nil, fmt.Errorf("unable to create listener file: %v", err)
	}
	defer listenerFile.Close()

	// Create a net.Listener from the *os.File.
	ln, err := net.FileListener(listenerFile)
	if err != nil {
		return nil, err
	}

	return ln, nil
}
