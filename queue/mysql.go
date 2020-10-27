package queue

import (
	"errors"
	"fmt"
	"github.com/andyzhou/tinycells/tc"
	"log"
	"strings"
	"sync"
)

/*
 * mysql queue face
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

 //inter macro define
 const (
 	MysqlReqChanSize = 128
 )

 //inter request and response
 type MysqlResp struct {
	Err error
	Resp interface{}
	LastInsertId int64
	EffectRows int64
 }

 type MysqlReq struct {
	Sql string
	Values []interface{}
	RespChan chan MysqlResp
	IsLazy bool
 }

 //mysql queue
 type MysqlQueue struct {
	 tag string
	 db *tc.DBService `db instance map`
	 reqChan chan MysqlReq `request lazy chan`
	 closeChan chan bool
	 sync.RWMutex
 }
 
 //construct
func NewMysqlQueue(db *tc.DBService) *MysqlQueue {
	//self init
	this := &MysqlQueue{
		db:db,
		reqChan:make(chan MysqlReq, MysqlReqChanSize),
		closeChan:make(chan bool),
	}

	//spawn main process
	go this.runMainProcess()

	return this
}

//////////
//api
//////////

//queue quit
func (q *MysqlQueue) Quit() {
	q.closeChan <- true
}

//send request
func (q *MysqlQueue) SendReq(req *MysqlReq) (resp *MysqlResp) {
	//init return result
	resp = &MysqlResp{}

	//try catch panic
	defer func(resp *MysqlResp) {
		if err := recover(); err != nil {
			tips := fmt.Sprintf("MysqlQueue::SendReq panic happened, err:%v", err)
			log.Println(tips)
			resp.Err = errors.New(tips)
			return
		}
	}(resp)

	//send to chan
	q.reqChan <- *req

	if !req.IsLazy {
		//wait for sync request response
		*resp, _ = <- req.RespChan
	}

	return
}

//////////////////
//private func
//////////////////

//process real redis request
func (q *MysqlQueue) processRequest(req *MysqlReq) *MysqlResp {
	//basic check
	if req == nil || q.db == nil {
		return nil
	}

	//analyze sql statement
	sqlSlice := strings.Split(req.Sql, " ")
	if sqlSlice == nil || len(sqlSlice) <= 0 {
		return nil
	}

	//get main command, like select, insert, update, etc.
	mainCommand := strings.ToUpper(sqlSlice[0])
	resp := MysqlResp{}
	if mainCommand == "SELECT" {
		//query original data
		result, err := q.db.GetArray(req.Sql, req.Values...)
		if err != nil {
			log.Println("MysqlQueue::processRequest failed, err:", err.Error())
			resp.Err = err
			return &resp
		}
		if !req.IsLazy {
			//analyze result
			q.analyzeRecordList(result)
			resp.Resp = result
		}
	}else{
		lastInsertId, effectRows, err := q.db.Execute(req.Sql, req.Values...)
		if err != nil {
			log.Println("MysqlQueue::processRequest failed, err:", err.Error())
			resp.Err = err
			return &resp
		}
		resp.LastInsertId = lastInsertId
		resp.EffectRows = effectRows
	}
	return &resp
}

//analyze record list
func (q *MysqlQueue) analyzeRecordList(recordSlice []map[string]interface{}) {
	for _, record := range recordSlice {
		for k, v := range record {
			switch v.(type) {
			case []uint8:
				{
					if v1, ok := v.([]uint8); ok {
						v = string(v1)
						record[k] = v
					}
				}
			}
		}
	}
}

//run main process
func (q *MysqlQueue) runMainProcess() {
	var (
		req MysqlReq
		resp = &MysqlResp{}
		isOk, needQuit bool
	)
	for {
		if needQuit && len(q.reqChan) == 0 {
			break
		}
		select {
		case req, isOk = <- q.reqChan://request
			{
				if isOk {
					//process request
					resp = q.processRequest(&req)
					if !req.IsLazy {
						req.RespChan <- *resp
					}
				}
			}
		case <- q.closeChan:
			needQuit = true
		}
	}
	//close relate chan
	close(q.reqChan)
	close(q.closeChan)
}
