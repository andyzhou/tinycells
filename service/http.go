package service

import (
	"github.com/andyzhou/tinycells/tc"
	"github.com/andyzhou/tinycells/queue"
	"sync"
)

/*
 * Http client service
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

//http client service
type HttpService struct {
	queues int `total queues`
	httpQueueMap map[int]*queue.HttpQueue
	tc.Utils
	sync.RWMutex `data locker`
}

//construct
func NewHttpService(queues int) *HttpService {
	//self init
	this := &HttpService{
		queues:queues,
		httpQueueMap:make(map[int]*queue.HttpQueue),
	}
	//inter init
	go this.interInit()
	return this
}


//service process quit
func (s *HttpService) Quit() {
	for _, queue := range s.httpQueueMap {
		queue.Quit()
	}
}

//send request
func (s *HttpService) SendRequest(reqObj *queue.HttpReq) []byte {
	//basic check
	if reqObj == nil {
		return nil
	}

	//get http queue
	httpQueue := s.getRandomQueue()
	if httpQueue == nil {
		return nil
	}

	//send to son queue
	resp := httpQueue.SendReq(reqObj)

	return resp
}

///////////////////////////
//private func for service
//////////////////////////

//get random queue
func (s *HttpService) getRandomQueue() *queue.HttpQueue {
	//basic check
	if s.httpQueueMap == nil {
		return nil
	}
	queueSize := len(s.httpQueueMap)
	if queueSize <= 0 {
		return nil
	}

	//get random index
	idx := s.GetRandomVal(queueSize) + 1
	v, ok := s.httpQueueMap[idx]
	if !ok {
		return nil
	}
	return v
}

//inter init
func (s *HttpService) interInit() bool {
	//init batch http queue with locker
	s.Lock()
	defer s.Lock()
	for i := 1; i <= s.queues; i++ {
		//init son queue
		httpQueue := queue.NewHttpQueue()
		s.httpQueueMap[i] = httpQueue
	}
	return true
}

