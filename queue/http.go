package queue

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

/*
 * http queue face
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

//http request method
const (
	HttpReqGet = iota
	HttpReqPost
)

//inter macro define
const (
	HttpClientTimeOut = 5
	HttpReqChanSize = 128
)

//http file para
type HttpFilePara struct {
	FilePath string
	FilePara string
}

//http request
type HttpReq struct {
	Kind int `GET or POST`
	Url string
	Headers map[string]string
	Params map[string]interface{}
	FilePara HttpFilePara
	Body []byte
	ReceiverChan chan []byte `http request receiver chan`
	IsAsync bool
}

//http queue
type HttpQueue struct {
	client *http.Client `http client instance`
	reqChan chan HttpReq `request lazy chan`
	closeChan chan bool
	sync.RWMutex
}

//construct
func NewHttpQueue() *HttpQueue {
	//self init
	this := &HttpQueue{
		reqChan:make(chan HttpReq, HttpReqChanSize),
		closeChan:make(chan bool, 1),
	}

	//inter init
	this.interInit()

	//spawn main process
	go this.runMainProcess()

	return this
}

func NewHttpReq() *HttpReq {
	this := &HttpReq{
		Headers:make(map[string]string),
		Params:make(map[string]interface{}),
		Body:make([]byte, 0),
		ReceiverChan:make(chan []byte),
	}
	return this
}


//////////
//api
//////////

//queue quit
func (q *HttpQueue) Quit() {
	q.closeChan <- true
}

//send request
func (q *HttpQueue) SendReq(req *HttpReq) (resp []byte) {
	//init return result
	resp = make([]byte, 0)

	//try catch panic
	defer func(resp []byte) {
		if err := recover(); err != nil {
			log.Println("HttpQueue::SendReq, panic happened, err:", err)
			return
		}
	}(resp)

	//send to chan
	q.reqChan <- *req

	//if request is async, just return
	if req.IsAsync {
		return
	}

	//sync mode, need wait for response
	resp, _ = <- req.ReceiverChan
	return
}

//////////////////
//private func
//////////////////

//run main process
func (q *HttpQueue) runMainProcess() {
	var (
		req HttpReq
		resp = make([]byte, 0)
		needQuit, isOk bool
	)
	for {
		if needQuit && len(q.reqChan) <= 0 {
			break
		}
		select {
		case req, isOk = <- q.reqChan:
			//process single request
			if isOk {
				resp = q.sendHttpReq(&req)
				if resp == nil {
					req.ReceiverChan <- []byte{}
				}else{
					req.ReceiverChan <- resp
				}
			}
		case <- q.closeChan:
			needQuit = true
		}
	}
	//close chan
	close(q.reqChan)
	close(q.closeChan)
}

//upload file request
func (q *HttpQueue) fileUploadReq(reqObj *HttpReq) (*http.Request, error) {
	//try open file
	file, err := os.Open(reqObj.FilePara.FilePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	//init multi part
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(
							reqObj.FilePara.FilePara,
							filepath.Base(reqObj.FilePara.FilePath),
						)
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, file)

	//add extend parameters
	for key, val := range reqObj.Params {
		v2, ok := val.(string)
		if !ok {
			continue
		}
		_ = writer.WriteField(key, v2)
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", reqObj.Url, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req, err
}

//general request
func (q *HttpQueue) generalReq(reqObj *HttpReq) (*http.Request, error) {
	var (
		tempStr string
		buffer = bytes.NewBuffer(nil)
		req *http.Request
		err error
	)

	//check parameters
	if len(reqObj.Params) > 0 {
		i := 0
		for k, v := range reqObj.Params {
			if i > 0 {
				buffer.WriteString("&")
			}
			tempStr = fmt.Sprintf("%s=%v", k, v)
			buffer.WriteString(tempStr)
			i++
		}
	}

	//get request method
	switch reqObj.Kind {
	case HttpReqPost:
		{
			if reqObj.Body != nil {
				buffer.Write(reqObj.Body)
			}
			//int post req
			req, err = http.NewRequest("POST", reqObj.Url, strings.NewReader(buffer.String()))
			if req.Header == nil || len(req.Header) <= 0 {
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			}
		}
	default:
		//init get req
		req, err = http.NewRequest("GET", reqObj.Url, nil)
	}

	//format post form
	if reqObj.Params != nil {
		for k, v := range reqObj.Params {
			keyVal := fmt.Sprintf("%v", v)
			req.Form.Add(k, keyVal)
		}
	}

	return req, err
}

//send original http request and get response
func (q *HttpQueue) sendHttpReq(reqObj *HttpReq) []byte {
	var (
		req *http.Request
		err error
	)

	//basic check
	if q.client == nil || reqObj == nil {
		return nil
	}
	if reqObj.Kind < HttpReqGet || reqObj.Url == "" {
		return nil
	}

	if reqObj.FilePara.FilePath != "" &&
	   reqObj.FilePara.FilePara != "" {
		//file upload request
		req, err = q.fileUploadReq(reqObj)
	}else{
		//general request
		req, err = q.generalReq(reqObj)
	}
	if err != nil {
		log.Println("HttpQueue::sendHttpReq, create request failed, err:", err.Error())
		return nil
	}

	//set headers
	if reqObj.Headers != nil {
		for k, v := range reqObj.Headers {
			req.Header.Set(k, v)
		}
	}

	//set http connect close
	req.Header.Set("Connection", "close")
	req.Close = true

	//begin send request
	//c.client.Timeout = time.Second
	resp, err := q.client.Do(req)
	if err != nil {
		log.Println("HttpQueue::sendHttpReq, send http request failed, err:", err.Error())
		return nil
	}

	//close resp before return
	defer resp.Body.Close()

	//read response
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("HttpQueue::sendHttpReq, read response body failed, err:", err.Error())
		return nil
	}

	//return response
	return respBody
}

//inter init
func (q *HttpQueue) interInit() {
	//init http trans
	netTransport := &http.Transport{
		Dial: (&net.Dialer{
			Timeout: HttpClientTimeOut * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: HttpClientTimeOut * time.Second,
		ResponseHeaderTimeout: HttpClientTimeOut * time.Second,
		ExpectContinueTimeout: HttpClientTimeOut* time.Second,
	}

	//init native http client
	q.client = &http.Client{
		Timeout:time.Second * HttpClientTimeOut,
		Transport:netTransport,
	}
}