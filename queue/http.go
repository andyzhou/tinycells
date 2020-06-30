package queue

import (
	"io/ioutil"
	"net/http"
	"strings"
	"bytes"
	"sync"
	"time"
	"log"
	"fmt"
	"net"
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

//http request
type HttpReq struct {
	Kind int `GET or POST`
	Url string
	Headers map[string]string
	Params map[string]interface{}
	FilePara string
	FilePath string
	Body []byte
	receiverChan chan []byte `http request receiver chan`
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
		closeChan:make(chan bool),
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
		receiverChan:make(chan []byte, HttpReqChanSize),
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

	//wait for response
	resp, _ = <- req.receiverChan

	return
}

//////////////////
//private func
//////////////////

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

	//general request
	req, err = q.generalReq(reqObj)
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
					req.receiverChan <- []byte{}
				}else{
					req.receiverChan <- resp
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