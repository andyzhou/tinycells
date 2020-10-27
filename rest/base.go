package rest

import (
	"errors"
	"github.com/emicklei/go-restful"
	"github.com/gorilla/schema"
	"log"
	"sync"
)

/*
 * base interface for Rest api
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

//base face
type BaseRest struct {
	ws *restful.WebService `web service instance`
	decoder *schema.Decoder
	sync.RWMutex `share data locker`
}

//construct
func NewBaseRest(ws *restful.WebService) *BaseRest {
	this := &BaseRest{
		ws:ws,
		decoder:schema.NewDecoder(),
	}
	return this
}

//get schema decoder
func (f *BaseRest) GetSchemaDecoder() *schema.Decoder {
	return f.decoder
}

//get web service
func (f *BaseRest) GetWebService() *restful.WebService {
	return f.ws
}

//parse request form
func (f *BaseRest) ParseReqForm(
					formFace interface{},
					req *restful.Request) error {
	//basic check
	if formFace == nil || req == nil {
		return errors.New("invalid parameters")
	}

	//parse post form
	err := req.Request.ParseForm()
	if err != nil {
		return err
	}

	//decode form data
	err = f.decoder.Decode(formFace, req.Request.PostForm)
	if err != nil {
		return err
	}
	return nil
}

//register dynamic sub route
//dynamicRootUrl like /test/{para1}/{para2}
//should use request.PathParameter("para1") to get real path parameter value
func (f *BaseRest) RegisterDynamicSubRoute(
					method string,
					consumes string,
					dynamicRootUrl string,
					dynamicPathSlice []string,
					routeFunc restful.RouteFunction,
					ws *restful.WebService,
				) bool {

	//basic check
	if dynamicRootUrl == "" || routeFunc == nil || ws == nil {
		return false
	}

	//init new route builder
	rb := new(restful.RouteBuilder)
	rb.Method(method).Path(dynamicRootUrl).To(routeFunc)

	if consumes != "" {
		rb.Consumes(consumes)
	}

	//init path parameter
	if dynamicPathSlice != nil && len(dynamicPathSlice) > 0 {
		for _, key := range dynamicPathSlice {
			pp := f.CreatePathParameter(key, "string", ws)
			rb.Param(pp)
		}
	}

	//add sub route
	ws.Route(rb)

	return true
}

//register static sub route
func (f *BaseRest) RegisterSubRoute(
				method, routeUrl, consumes string,
				parameters [] *restful.Parameter,
				routeFunc restful.RouteFunction,
				ws *restful.WebService) bool {

	//basic check
	if method == "" || routeUrl == "" {
		return false
	}
	if routeFunc == nil || ws == nil {
		return false
	}

	//init new route builder
	rb := new(restful.RouteBuilder)

	//set method, request url and route func
	rb.Method(method).Path(routeUrl).To(routeFunc)

	if consumes != "" {
		rb.Consumes(consumes)
	}

	//set parameter
	if parameters != nil && len(parameters) > 0 {
		for _, parameter := range parameters {
			//set sub parameter
			rb.Param(parameter)
		}
	}

	//add sub route
	ws.Route(rb)

	//ws.PathParameter("key", "parameter for key").DataType("string")

	return true
}

//create ws form parameter
func (f *BaseRest) CreateParameter(
					name, kind, defaultVal string,
					ws *restful.WebService) *restful.Parameter {
	//basic check
	if name == "" || kind == "" {
		return nil
	}
	//init new
	param := ws.FormParameter(name, "").DataType(kind).DefaultValue(defaultVal)
	return param
}

//create ws path parameter
func (f *BaseRest) CreatePathParameter(
				name, kind string,
				ws *restful.WebService) *restful.Parameter {
	//basic check
	if name == "" || kind == "" {
		return nil
	}
	//init new
	param := ws.PathParameter(name, "").DataType(kind)
	return param
}

//create empty parameter slice
func (f *BaseRest) CreateParameters() []*restful.Parameter {
	parameters := make([]*restful.Parameter, 0)
	return parameters
}

//generate response json data
func (f *BaseRest) GenJsonResp(jsonObj interface{}, resp *restful.Response) {
	if jsonObj == nil || resp == nil {
		return
	}
	//send to client side
	err := resp.WriteAsJson(jsonObj)
	if err != nil {
		log.Println("BaseRest::GenJsonResp failed, err:", err.Error())
	}
}

