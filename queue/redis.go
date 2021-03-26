package queue

import (
	"context"
	"errors"
	"fmt"
	"github.com/andyzhou/tinycells/tc"
	"github.com/go-redis/redis"
	"log"
	"strconv"
	"sync"
	"time"
)

/*
 * redis queue face
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

 //inter macro define
 const (
	RedisReqChanSize = 256
 )

 //inter request and response
 type RedisResp struct {
	Err error
	Resp interface{}
 }

 type RedisReq struct {
	Command string
	ParaSlice []interface{}
	ExpireSeconds int64
	RespChan chan RedisResp
 }

 type RedisLuaReq struct {
	Script string
	Keys []string
	Args []interface{}
	RespChan chan RedisResp
 }

 //atom locker request
 type RedisLockerReq struct {
 	Key string
 	ExpireSeconds int
 	RespChan chan RedisResp
 }

 //redis queue
 type RedisQueue struct {
	tag string
	gRedis *tc.GRedis `redis instance`
	reqChan chan RedisReq `request lazy chan`
	luaReqChan chan RedisLuaReq
	lockerReqChan chan RedisLockerReq
	closeChan chan bool
	sync.RWMutex
 }

//construct
func NewRedisQueue(gRedis *tc.GRedis) *RedisQueue {
	return NewRedisQueueWithChanSize(gRedis, 0)
}

func NewRedisQueueWithChanSize(gRedis *tc.GRedis, reqChanSize int) *RedisQueue {
	//check or init request chan size
	realReqChanSize := reqChanSize
	if realReqChanSize <= 0 {
		realReqChanSize = RedisReqChanSize
	}

	//self init
	this := &RedisQueue{
		gRedis:gRedis,
		reqChan:make(chan RedisReq, realReqChanSize),
		luaReqChan:make(chan RedisLuaReq, realReqChanSize),
		lockerReqChan:make(chan RedisLockerReq, realReqChanSize),
		closeChan:make(chan bool, 1),
	}

	//spawn main process
	go this.runMainProcess()

	return this
}

//////////
//api
//////////

//queue quit
func (q *RedisQueue) Quit() {
	q.closeChan <- true
}

//send locker request
func (q *RedisQueue) SendLockerReq(req *RedisLockerReq) (resp *RedisResp) {
	//init response
	resp = &RedisResp{}

	//try catch panic
	defer func(resp *RedisResp) {
		if err := recover(); err != nil {
			tips := fmt.Sprintf("RedisQueue::SendLockerReq panic happened, err:%v", err)
			log.Println(tips)
			resp.Resp = nil
			resp.Err = errors.New(tips)
		}
	}(resp)

	//send to chan
	q.lockerReqChan <- *req

	//wait for response
	*resp, _ = <- req.RespChan
	return
}

//send lua opt request
func (q *RedisQueue) SendLuaReq(req *RedisLuaReq) (resp *RedisResp) {
	//init response
	resp = &RedisResp{}

	//try catch panic
	defer func(resp *RedisResp) {
		if err := recover(); err != nil {
			tips := fmt.Sprintf("RedisQueue::SendLuaReq panic happened, err:%v", err)
			log.Println(tips)
			resp.Resp = nil
			resp.Err = errors.New(tips)
		}
	}(resp)

	//send to chan
	q.luaReqChan <- *req

	//wait for response
	*resp, _ = <- req.RespChan
	return
}

//send general opt request
func (q *RedisQueue) SendReq(req *RedisReq) (resp *RedisResp) {
	//init response
	resp = &RedisResp{}

	//try catch panic
	defer func(resp *RedisResp) {
		if err := recover(); err != nil {
			tips := fmt.Sprintf("RedisQueue::SendReq panic happened, err:%v", err)
			log.Println(tips)
			resp.Resp = nil
			resp.Err = errors.New(tips)
		}
	}(resp)

	//send to chan
	q.reqChan <- *req

	//wait for response
	*resp, _ = <- req.RespChan
	return
}

//////////////////
//private func
//////////////////

//process single request
func (q *RedisQueue) processRequest(req *RedisReq) *RedisResp {
	//init response
	resp := &RedisResp{}

	//basic check
	if q.gRedis == nil {
		resp.Err = errors.New("redis instance is nil")
		resp.Resp = nil
		return resp
	}

	//get client
	client := q.gRedis.GetClient()
	if client == nil {
		resp.Err = errors.New("redis instance is nil")
		resp.Resp = nil
		return resp
	}

	//init real args slice
	args := make([]interface{}, 0)
	args = append(args, req.Command)
	args = append(args, req.ParaSlice...)

	//execute command
	genVal, err := client.Do(context.Background(), args...).Result()

	//set key expire seconds
	if err != nil && req.ExpireSeconds > 0 {
		if len(req.ParaSlice) > 0 {
			redisKey, isOk := req.ParaSlice[0].(string)
			if isOk {
				timeEnd := time.Now().Unix() + req.ExpireSeconds
				timeDuration := time.Duration(timeEnd) * time.Second
				client.Expire(context.Background(), redisKey, timeDuration).Result()
			}
		}
	}

	//copy into response
	resp.Resp = genVal
	resp.Err = err

	return resp
}

//process lua script request
func (q *RedisQueue) processLuaRequest(req *RedisLuaReq) *RedisResp {
	resp := &RedisResp{}

	//basic check
	if q.gRedis == nil {
		resp.Err = errors.New("redis instance is nil")
		resp.Resp = nil
		return resp
	}

	//init lua script
	script := redis.NewScript(req.Script)
	respVal, err := script.Run(
						context.Background(),
						q.gRedis.GetClient(),
						req.Keys,
						req.Args...,
					).Result()

	if err != nil && err == redis.Nil {
		log.Println("key dose not exists")
	}
	//set return value
	resp.Resp = respVal
	resp.Err = err

	return resp
}

//process locker request
//setnx+getset
func (q *RedisQueue) processLockerRequest(req *RedisLockerReq) *RedisResp {
	resp := &RedisResp{
		Resp:false,
	}

	//basic check
	if q.gRedis == nil {
		resp.Err = errors.New("redis instance is nil")
		resp.Resp = nil
		return resp
	}

	//try set locker
	client := q.gRedis.GetClient()
	if client == nil {
		resp.Err = errors.New("invalid client of redis instance")
		resp.Resp = nil
		return resp
	}

	//set locker value
	now := time.Now().Unix()
	value := now + int64(req.ExpireSeconds)

	////////////
	//lua mode
	////////////
	luaScript := `
		local key = KEYS[1]
		local val = tonumber(ARGV[1])
		local now = tonumber(ARGV[2])
		local oldLocker
		local locker
		local result
		locker = redis.call('SETNX', key, val)
		if locker > 0 then
			result = true
		else
			oldLocker = tonumber(redis.call('GET', key))
			if oldLocker < now then
				redis.call('DEL', key)
				locker = redis.call('SETNX', key, val)
				if locker > 0 then
					result = true
				else
					result = false
				end
			else
				result = false
			end
		end
		return result
	`

	//init lua script
	keys := []string{
		req.Key,
	}
	script := redis.NewScript(luaScript)
	respVal, err := script.Run(context.Background(), client, keys, value, now).Result()

	//set return value
	resp.Resp = respVal
	resp.Err = err
	return resp


	////////////
	//general mode
	////////////
	bRet, err := client.SetNX(context.Background(), req.Key, value, 0).Result()
	if err != nil {
		resp.Err = err
		return resp
	}

	//get locker
	if bRet {
		resp.Resp = true
		return resp
	}

	//try get locker
	lockerTimeStr, err := client.Get(context.Background(), req.Key).Result()
	if err != nil {
		resp.Err = err
		return resp
	}

	//check locker
	lockerTimeInt, _ := strconv.ParseInt(lockerTimeStr, 10, 64)
	if lockerTimeInt < now {
		//try reset locker
		client.Del(context.Background(), req.Key)
		bRet, err = client.SetNX(context.Background(), req.Key, value, 0).Result()
		if err != nil {
			resp.Err = err
			return resp
		}

		//set response value
		resp.Resp = bRet
	}

	return resp
}

//run main process
func (q *RedisQueue) runMainProcess() {
	var (
		req RedisReq
		luaReq RedisLuaReq
		lockerReq RedisLockerReq
		resp = &RedisResp{}
		isOk, needQuit bool
	)

	//defer
	defer func() {
		//close chan
		close(q.reqChan)
		close(q.luaReqChan)
		close(q.lockerReqChan)
		close(q.closeChan)
	}()

	//loop receive
	for {
		if needQuit && len(q.reqChan) <= 0 && len(q.luaReqChan) <= 0 {
			break
		}
		select {
		case req, isOk = <- q.reqChan://general request
			{
				if isOk {
					resp = q.processRequest(&req)
					req.RespChan <- *resp
				}
			}
		case luaReq, isOk = <- q.luaReqChan://lua script request
			{
				if isOk {
					resp = q.processLuaRequest(&luaReq)
					luaReq.RespChan <- *resp
				}
			}
		case lockerReq, isOk = <- q.lockerReqChan://locker request
			{
				if isOk {
					resp = q.processLockerRequest(&lockerReq)
					lockerReq.RespChan <- *resp
				}
			}
		case <- q.closeChan:
			needQuit = true
		}
	}
}
