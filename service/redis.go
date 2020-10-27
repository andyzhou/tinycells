package service

import (
	"errors"
	"github.com/andyzhou/tinycells/queue"
	"github.com/andyzhou/tinycells/tc"
	"strconv"
	"sync"
)

/*
 * service for redis data opt
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 * this is internal service
 */

//single redis config
type RedisConf struct {
	Switcher bool
	Server string
	Password string
	Db string `db name`
	Pools int `connect pools`
	Queues int `sub queues`
}

//redis service
type RedisService struct {
	sonWorkers map[string]*RedisSonWorker `tag:object`
	sync.RWMutex
}

//son worker
type RedisSonWorker struct {
	tag string
	redisConf *RedisConf
	redisQueueMap map[int]*queue.RedisQueue
	tc.Utils
	sync.RWMutex
}

//construct
func NewRedisService() *RedisService {
	//self init
	this := &RedisService{
		sonWorkers:make(map[string]*RedisSonWorker),
	}
	return this
}

func NewRedisSonWorker(tag string, redisConf *RedisConf) *RedisSonWorker {
	//self init
	this := &RedisSonWorker{
		tag:tag,
		redisConf:redisConf,
		redisQueueMap:make(map[int]*queue.RedisQueue),
	}
	//inter init
	go this.interInit()
	return this
}


//////////
//api
//////////

//quit
func (w *RedisService) Quit() bool {
	if w.sonWorkers == nil || len(w.sonWorkers) <= 0 {
		return false
	}
	for _, sonWorker := range w.sonWorkers {
		sonWorker.quit()
	}
	return true
}

//send redis locker request opt
func (w *RedisService) SendLockerReq(
						tag, key string,
						expireSeconds int,
					) *queue.RedisResp {
	//basic check
	resp := &queue.RedisResp{}
	if tag == "" || tag == "" || expireSeconds <= 0 {
		resp.Err = errors.New("lost parameters")
		return resp
	}

	//get son worker
	sonWorker := w.getSonWorker(tag)
	if sonWorker == nil {
		resp.Err = errors.New("get son worker by tag")
		return resp
	}

	//init inter request
	req := queue.RedisLockerReq{
		Key:key,
		ExpireSeconds:expireSeconds,
		RespChan:make(chan queue.RedisResp, 1),
	}

	//send to son worker
	resp = sonWorker.sendLockerReq(&req)

	return resp
}

//send single redis lua request opt
func (w *RedisService) SendLuaReq(
					tag, script string,
					keys []string,
					args []interface{},
				) *queue.RedisResp {
	//basic check
	resp := &queue.RedisResp{}
	if tag == "" || script == "" {
		resp.Err = errors.New("lost parameters")
		return resp
	}

	//get son worker
	sonWorker := w.getSonWorker(tag)
	if sonWorker == nil {
		resp.Err = errors.New("get son worker by tag")
		return resp
	}

	//init inter request
	req := queue.RedisLuaReq{
		Script:script,
		Keys:keys,
		Args:args,
		RespChan:make(chan queue.RedisResp, 1),
	}

	//send to son worker
	resp = sonWorker.sendLuaReq(&req)

	return resp
}

//send single redis request opt
func (w *RedisService) SendReq(
					tag, command string,
					paraSlice []interface{},
					) *queue.RedisResp {
	//basic check
	resp := &queue.RedisResp{}
	if tag == "" || command == "" {
		resp.Err = errors.New("lost parameters")
		return resp
	}

	//get son worker
	sonWorker := w.getSonWorker(tag)
	if sonWorker == nil {
		resp.Err = errors.New("get son worker by tag")
		return resp
	}

	//init inter request
	req := queue.RedisReq{
		Command:command,
		ParaSlice:make([]interface{}, 0),
		RespChan:make(chan queue.RedisResp, 1),
	}

	//copy para slice
	req.ParaSlice = append(req.ParaSlice, paraSlice...)

	//send to son worker
	resp = sonWorker.sendReq(&req)

	return resp
}

//send single redis request opt
func (w *RedisService) SendReqWithExpire(
		tag, command string,
		paraSlice []interface{},
		expireSeconds int64,
	) *queue.RedisResp {
	//basic check
	resp := &queue.RedisResp{}
	if tag == "" || command == "" {
		resp.Err = errors.New("lost parameters")
		return resp
	}

	//get son worker
	sonWorker := w.getSonWorker(tag)
	if sonWorker == nil {
		resp.Err = errors.New("get son worker by tag")
		return resp
	}

	//init inter request
	req := queue.RedisReq{
		Command:command,
		ParaSlice:make([]interface{}, 0),
		ExpireSeconds:expireSeconds,
		RespChan:make(chan queue.RedisResp, 1),
	}

	//copy para slice
	req.ParaSlice = append(req.ParaSlice, paraSlice...)

	//send to son worker
	resp = sonWorker.sendReq(&req)

	return resp
}

//add son worker
func (w *RedisService) AddSonWorker(tag string, redisConf *RedisConf) bool {
	//basic check
	if tag == "" || redisConf == nil {
		return false
	}
	w.createSonWorker(tag, redisConf)
	return true
}

////////////////////////////////////
//private func for RedisSonWorker
////////////////////////////////////

//quit
func (ws *RedisSonWorker) quit() {
	//ws.closeChan <- true
	for _, redisQueue := range ws.redisQueueMap {
		redisQueue.Quit()
	}
}
//send locker opt request
func (ws *RedisSonWorker) sendLockerReq(req *queue.RedisLockerReq) (resp *queue.RedisResp) {
	//init response
	resp = &queue.RedisResp{}

	//get random queue
	sonQueue := ws.getRandomQueue()
	if sonQueue == nil {
		resp.Err = errors.New("can't get son redis queue")
		return
	}

	//send request to son queue
	resp = sonQueue.SendLockerReq(req)

	return
}

//send lua opt request
func (ws *RedisSonWorker) sendLuaReq(req *queue.RedisLuaReq) (resp *queue.RedisResp) {
	//init response
	resp = &queue.RedisResp{}

	//get random queue
	sonQueue := ws.getRandomQueue()
	if sonQueue == nil {
		resp.Err = errors.New("can't get son redis queue")
		return
	}

	//send request to son queue
	resp = sonQueue.SendLuaReq(req)

	return
}

//send general opt request
func (ws *RedisSonWorker) sendReq(req *queue.RedisReq) (resp *queue.RedisResp) {
	//init response
	resp = &queue.RedisResp{}

	//get random queue
	sonQueue := ws.getRandomQueue()
	if sonQueue == nil {
		resp.Err = errors.New("can't get son redis queue")
		return
	}

	//send request to son queue
	resp = sonQueue.SendReq(req)

	return
}

//get random queue
func (ws *RedisSonWorker) getRandomQueue() *queue.RedisQueue {
	//basic check
	if ws.redisQueueMap == nil {
		return nil
	}
	queueSize := len(ws.redisQueueMap)
	if queueSize <= 0 {
		return nil
	}

	//get random index
	idx := ws.GetRandomVal(queueSize) + 1
	v, ok := ws.redisQueueMap[idx]
	if !ok {
		return nil
	}
	return v
}

//inter init
func (ws *RedisSonWorker) interInit() bool {
	//init batch redis queue with locker
	ws.Lock()
	defer ws.Unlock()
	for i := 1; i <= ws.redisConf.Queues; i++ {
		//init son queue
		gRedis := ws.createRedis()
		redisQueue := queue.NewRedisQueue(gRedis)
		ws.redisQueueMap[i] = redisQueue
	}

	return true
}


//init redis instance
func (ws *RedisSonWorker) createRedis() *tc.GRedis {
	//basic check
	if ws.redisConf == nil || !ws.redisConf.Switcher {
		return nil
	}
	//init redis service
	dbInt, _ := strconv.Atoi(ws.redisConf.Db)
	gRedis := tc.NewGRedisWithPoolSize(
		ws.redisConf.Server,
		ws.redisConf.Password,
		dbInt,
		ws.redisConf.Pools,
	)
	return gRedis
}

/////////////////////////////////
//private func for RedisService
/////////////////////////////////

//get son worker by tag
func (w *RedisService) getSonWorker(tag string) *RedisSonWorker {
	//get worker
	worker, ok := w.sonWorkers[tag]
	if !ok {
		return nil
	}
	return worker
}

//create son worker
func (w *RedisService) createSonWorker(tag string, redisConf *RedisConf) bool {
	if tag == "" || redisConf == nil {
		return false
	}
	w.Lock()
	defer w.Unlock()

	//check
	_, ok := w.sonWorkers[tag]
	if ok {
		return false
	}

	//init son worker with locker
	sonWorker := NewRedisSonWorker(tag, redisConf)
	w.sonWorkers[tag] = sonWorker

	return true
}
