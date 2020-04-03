package service

import (
	"github.com/emicklei/go-restful"
	"log"
	"fmt"
	"net/http"
)

/**
 * REST web api service face
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 * base on `github.com/emicklei/go-restful` lib
 */

//web rest api service
type RestService struct {
	httpPort int
	ws *restful.WebService `ws instance`
	//route *rest.Route
}

//construct
func NewRestService(httpPort int) *RestService {
	//self init
	this := &RestService{
		httpPort:httpPort,
		ws:new(restful.WebService),
	}
	//inter init
	this.interInit()
	return this
}

//start service
func (s *RestService) Start() {
	//add default container
	restful.DefaultContainer.Add(s.ws)

	//start http service
	log.Println("start web service on port", s.httpPort)
	portStr := fmt.Sprintf(":%d", s.httpPort)
	http.ListenAndServe(portStr, nil)
}

//get web service
func (s *RestService) GetWebService() *restful.WebService {
	return s.ws
}

//inter init
func (s *RestService) interInit() {
	//if http port <= 0, http service can't be started.
	if s.httpPort <= 0 {
		return
	}

	//set mime kind, use json format
	s.ws.Consumes(restful.MIME_JSON).Produces(restful.MIME_JSON)

	//inter api init
	//s.route = rest.NewRoute(s.ws)
}
