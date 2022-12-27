package net

import (
	"errors"
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path"
	"sync"
	"time"
)

/*
 * ssh client, not completed!!!
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

//internal macro define
const (
	ModOfCertPassword = iota + 1
	ModOfCertPubKeyFile
)

const (
	ConnectTimeOut = 10 //seconds
	ConnectCheckRate = 30 //seconds
	ScpFileChanSize = 1024
	ScpFilePermission = "0666"
)

//ssh config
type SSHConfig struct {
	Mode int //1,2
	Address string
	User string
	CertPath string
	RemoteRootPath string
}

//ssh face info
type SSHClient struct {
	switcher bool
	conf *SSHConfig
	mode int `auth mode`
	address string `remote host:ip`
	user string `ssh account`
	cert string `ssh password or key file path`
	remoteRoot string `remote server root path`
	session *ssh.Session
	client *ssh.Client
	sync.RWMutex
}

//construct
func NewSSHClient() *SSHClient {
	this := &SSHClient{}
	return this
}

//quit
func (f *SSHClient) Quit() {

}

//set config, step-1
func (f *SSHClient) SetConfig(cfg *SSHConfig) error {
	//check
	if cfg == nil {
		return errors.New("invalid parameter")
	}
	f.conf = cfg
	return nil
}

//start, step-2
func (f *SSHClient) Start() error {
	//check
	if f.conf == nil {
		return errors.New("config not setup")
	}

	//set key data
	f.mode = f.conf.Mode
	f.address = f.conf.Address
	f.user = f.conf.User
	f.cert = f.conf.CertPath
	f.remoteRoot = f.conf.RemoteRootPath

	//try connect server
	err := f.connectServer(f.mode)
	return err
}

//read file
//todo...
func (f *SSHClient) ReadFile(filePath string) ([]byte, error) {
	//check
	if filePath == "" || f.client == nil {
		return nil, errors.New("invalid parameter or ssh client is nil")
	}
	//init new session
	session, err := f.client.NewSession()
	if err != nil {
		return nil, err
	}
	defer session.Close()
	return nil, err
}

//copy local file to server side
func (f *SSHClient) CopyFile(filePath string) error {
	//check
	if filePath == "" || f.client == nil {
		return errors.New("invalid parameter or ssh client is nil")
	}

	//try open source file
	h, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer h.Close()

	//get file status
	stat, err := h.Stat()
	if err != nil {
		return err
	}
	//init remote path
	fileName := path.Base(filePath)
	remoteFullPath := fmt.Sprintf("%s/%s", f.remoteRoot, fileName)

	//begin copy
	err = f.scpFile(h, remoteFullPath, ScpFilePermission, stat.Size())
	if err != nil {
		return err
	}

	//remove origin file
	err = os.Remove(filePath)
	return err
}

////////////////
//private func
////////////////

//scp file to remote server
func (f *SSHClient) scpFile(r io.Reader, remotePath, permissions string, size int64) error {
	//check
	if f.client == nil || r == nil || remotePath == "" || size <= 0 {
		return errors.New("invalid parameter")
	}
	//get file info
	filename := path.Base(remotePath)
	directory := path.Dir(remotePath)

	//init new session
	session, err := f.client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()
	w, err := session.StdinPipe()
	if err != nil {
		return err
	}

	//init and start command
	cmd := "/usr/bin/scp -t " + directory
	if err := session.Start(cmd); err != nil {
		w.Close()
		return err
	}
	errors := make(chan error)
	go func() {
		errors <- session.Wait()
	}()

	//begin copy original data
	fmt.Fprintln(w, "C"+permissions, size, filename)
	io.Copy(w, r)
	fmt.Fprint(w, "\x00")
	w.Close()
	err = <- errors
	return err
}

//connect server
func (f *SSHClient) connectServer(mode int) error {
	var (
		sshConfig *ssh.ClientConfig
		auth  []ssh.AuthMethod
		err error
	)

	//init auth by mode
	switch mode {
	case ModOfCertPassword:
		auth = []ssh.AuthMethod{ssh.Password(f.cert)}
	case ModOfCertPubKeyFile:
		auth = []ssh.AuthMethod{f.readPublicKeyFile(f.cert, err)}
		if err != nil {
			return err
		}
	default:
		return errors.New("un-support mode")
	}

	//init host key call back
	cb := func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		return nil
	}

	//begin init ssh client config
	sshConfig = &ssh.ClientConfig{
		User: f.user,
		Auth: auth,
		//if not need auth at server side, cb should be nil.
		HostKeyCallback:cb,
		Timeout:time.Second * ConnectTimeOut,
	}

	//try connect server
	client, err := ssh.Dial("tcp", f.address, sshConfig)
	if err != nil {
		return err
	}

	//connect success, sync client with locker
	f.Lock()
	defer f.Unlock()
	f.client = client
	return nil
}

//read pub key file
func (f *SSHClient) readPublicKeyFile(file string, err error) ssh.AuthMethod {
	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		return nil
	}
	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		return nil
	}
	return ssh.PublicKeys(key)
}
