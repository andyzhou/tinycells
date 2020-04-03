package tc

import (
	"reflect"
	"sync"
	"errors"
	"fmt"
)

/**
 * simple dynamic call instance method
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 *
 * for example:
 	face := base.NewFaceMap()
 	players := player.NewPlayers()
	face.Bind("players", players)
	result, err := face.Call("players", "Test", "this is string test")
	fmt.Println(result[0].String()) //print string type result
 */

 //internal macro variables
 const (
 	MaxInParams = 10
 )

//face map info
type FaceMap struct {
	faceMap map[string]reflect.Value `face:Value`
	callResult []reflect.Value `call reflect value`
	face reflect.Value `face reflect value`
	isOk bool
	tips string
	inParam []reflect.Value
	params int
	callErr error
	i int
	para interface{}
	sync.Mutex `data locker`
}

//construct
func NewFaceMap() *FaceMap {
	this := &FaceMap{
		faceMap:make(map[string]reflect.Value),
		callResult:make([]reflect.Value, 0),
		inParam:make([]reflect.Value, MaxInParams),
	}
	return this
}

///////
//API
///////

//get face instance
func (f *FaceMap) GetFace(name string) interface{} {
	face, ok := f.faceMap[name]
	if !ok {
		return nil
	}
	return face
}

//bind face with name
func (f *FaceMap) Bind(name string, face interface{}) bool {
	if name == "" || face == nil {
		return false
	}
	//check is exists or not
	if _, ok := f.faceMap[name]; ok {
		//already exists
		return true
	}
	//add face with locker
	f.Lock()
	f.faceMap[name] = reflect.ValueOf(face)
	f.Unlock()
	return true
}

//call method on all faces
func (f *FaceMap) Cast(method string, params ...interface{}) bool {
	var (
		para interface{}
		inParam = make([]reflect.Value, MaxInParams)
		paramNum = 0
		i = 0
	)
	if method == "" {
		inParam = []reflect.Value{}
		return false
	}
	//init parameters
	paramNum = len(params)
	for i, para = range params {
		if i >= MaxInParams {
			break
		}
		inParam[i] = reflect.ValueOf(para)
	}
	//call method on each face
	for _, face := range f.faceMap {
		face.MethodByName(method).Call(inParam[0:paramNum])
	}
	//reset in param slice
	inParam = []reflect.Value{}
	return true
}

//dynamic call method with parameters support
func (f *FaceMap) Call(name string, method string, params ...interface{}) ([]reflect.Value, error) {
	//result := make([]reflect.Value, 0)
	//check instance
	f.face, f.isOk = f.faceMap[name]
	if !f.isOk {
		f.tips = fmt.Sprintf("No face instance for name %s", name)
		return nil, errors.New(f.tips)
	}

	//init parameters
	f.params = len(params)
	for f.i, f.para = range params {
		if f.i >= MaxInParams {
			break
		}
		f.inParam[f.i] = reflect.ValueOf(f.para)
	}

	//check method is nil or not
	f.callResult = f.face.MethodByName(method).Call(f.inParam[0:f.params])

	return f.callResult, nil
}

