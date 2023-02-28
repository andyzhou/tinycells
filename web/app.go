package web

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"html/template"
	"os"
	"sync"
)

/*
 * face for gin app
 */

//request path
const (
	AnyPath = "any"
)

//sub web app interface
type IWebSubApp interface {
	Entry(c *gin.Context)
}

//face info
type App struct {
	port int //web port
	server *gin.Engine //gin server
	tplPath string //tpl root path
	//runner *iris.Runner //iris runner
	wg sync.WaitGroup
}

//construct
func NewApp(g ...*gin.Engine) *App {
	var (
		s *gin.Engine
	)
	//check
	if g != nil && len(g) > 0 {
		s = g[0]
	}else{
		s = gin.Default()
	}
	//self init
	this := &App{
		server: s,
	}
	return this
}

//stop app
func (f *App) Stop() {
	f.wg.Done()
}

//start app
func (f *App) Start(port int) bool {
	if port <= 0 {
		return false
	}

	//set port and address
	f.port = port
	addr := fmt.Sprintf(":%v", port)

	//start app
	f.wg.Add(1)
	f.server.Run(addr)
	f.wg.Wait()
	return true
}

//register root app entry
//url like: /xxx or /xxx/{ParaName:string|integer}
func (f *App) RegisterSubApp(
			reqUrlPara string,
			face IWebSubApp,
		) bool {
	//check
	if reqUrlPara == "" || face == nil {
		return false
	}

	//root request route
	requestAnyPath := fmt.Sprintf("/%v/*%v", reqUrlPara, AnyPath)

	//set get、post request
	f.server.Any(requestAnyPath, face.Entry)
	return true
}

//get gin engine
func (f *App) GetGin() *gin.Engine {
	return f.server
}

//set map func
func (f *App) SetMapFunc(mf template.FuncMap) {
	if mf == nil {
		return
	}
	f.server.SetFuncMap(mf)
}

//set static file path
func (f *App) SetStaticPath(url, path string) bool {
	if url == "" || path == "" {
		return false
	}
	f.server.Static(url, path)
	return true
}

//set tpl root path
func (f *App) SetTplPath(path string) bool {
	if path == "" {
		return false
	}
	//set tpl path
	f.tplPath = path

	//check path
	info, err := os.Stat(path)
	if err != nil || info == nil {
		return false
	}
	if !info.IsDir() {
		return false
	}

	//init templates
	f.server.LoadHTMLGlob(fmt.Sprintf("%v", f.tplPath))
	return true
}

//set gin reference
func (f *App) SetGin(gin *gin.Engine) {
	if gin == nil {
		return
	}
	f.server = gin
}