package service

import (
	"errors"
	"fmt"
	"github.com/andyzhou/tinycells/queue"
	"github.com/andyzhou/tinycells/tc"
	"sync"
)

/*
 * service for mysql
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

//single db config
type MysqlDBConf struct {
	Switcher bool
	Host string
	Port int
	User string
	Password string
	Name string `db name`
	Pools int `connect pools`
	Queues int `sub queues`
}

//mysql service
type MysqlService struct {
	sonWorkers map[string]*MysqlSonWorker `worker map`
	sync.Mutex
}

//son worker
type MysqlSonWorker struct {
	tag string
	dbConf *MysqlDBConf
	mysqlQueueMap map[int]*queue.MysqlQueue
	tc.Utils
	tc.BaseJson
	sync.RWMutex
}

//construct
func NewMysqlService() *MysqlService {
	//self init
	this := &MysqlService{
		sonWorkers:make(map[string]*MysqlSonWorker),
	}
	return this
}

func NewMysqlSonWorker(tag string, dbConf *MysqlDBConf) *MysqlSonWorker {
	//self init
	this := &MysqlSonWorker{
		tag:tag,
		dbConf:dbConf,
		mysqlQueueMap:make(map[int]*queue.MysqlQueue),
	}
	//inter init
	go this.interInit()
	return this
}


////////////////////
//api for MysqlService
////////////////////

//quit
func (w *MysqlService) Quit() bool {
	if len(w.sonWorkers) <= 0 {
		return false
	}
	for _, sonWorker := range w.sonWorkers {
		sonWorker.quit()
	}
	return true
}

//send opt request
func (w *MysqlService) SendReq(
			tag, sql string,
			values []interface{},
			isLazy bool) *queue.MysqlResp {
	//basic check
	if sql == "" {
		return nil
	}

	//get son worker
	sonWorker := w.getSonWorker(tag)
	if sonWorker == nil {
		return nil
	}

	//init inter request
	req := queue.MysqlReq{
		Sql:sql,
		Values:values,
		RespChan:make(chan queue.MysqlResp, 1),
	}

	//send request pass son worker
	resp := sonWorker.sendReq(&req)
	return resp
}


//reload redis instance
//called when config updated
func (w *MysqlService) Reload() bool {
	////get db config
	//dbConf := conf.RunServiceConf.GetDBConf()
	//if dbConf == nil {
	//	return false
	//}
	//
	////get all config
	//allDBConf := dbConf.GetAllDBConf()
	//if allDBConf == nil || len(allDBConf) <= 0 {
	//	return false
	//}
	//
	////dynamic load new mysql
	//for tag, _ := range allDBConf {
	//	w.createSonWorker(tag)
	//}
	return true
}

//add son worker
func (w *MysqlService) AddSonWorker(tag string, dbConf *MysqlDBConf) bool {
	//basic check
	if tag == "" || dbConf == nil {
		return false
	}
	w.createSonWorker(tag, dbConf)
	return true
}

////////////////////
//api for MysqlSonWorker
////////////////////

func (ws *MysqlSonWorker) quit() {
	for _, mysqlQueue := range ws.mysqlQueueMap {
		mysqlQueue.Quit()
	}
}

//send request
func (ws *MysqlSonWorker) sendReq(req *queue.MysqlReq) (resp *queue.MysqlResp) {
	//init return result
	resp = &queue.MysqlResp{}

	//get random son queue
	sonQueue := ws.getRandomQueue()
	if sonQueue == nil {
		resp.Err = errors.New("can't get son mysql queue")
		return
	}

	//send request to son queue
	resp = sonQueue.SendReq(req)

	return
}

////////////////////////////////////
//private func for MysqlSonWorker
////////////////////////////////////

//get random queue
func (ws *MysqlSonWorker) getRandomQueue() *queue.MysqlQueue {
	//basic check
	if ws.mysqlQueueMap == nil {
		return nil
	}
	queueSize := len(ws.mysqlQueueMap)
	if queueSize <= 0 {
		return nil
	}

	//get random index
	idx := ws.GetRandomVal(queueSize) + 1
	v, ok := ws.mysqlQueueMap[idx]
	if !ok {
		return nil
	}
	return v
}

//inter init of MysqlSonWorker
func (ws *MysqlSonWorker) interInit() bool {
	//init batch mysql queue with locker
	ws.Lock()
	defer ws.Unlock()
	for i := 1; i <= ws.dbConf.Queues; i++ {
		//init son queue
		dbService := ws.createMysql()
		mysqlQueue := queue.NewMysqlQueue(dbService)
		ws.mysqlQueueMap[i] = mysqlQueue
	}
	return true
}

//init db instance
func (ws *MysqlSonWorker) createMysql() *tc.DBService {
	//basic check
	if ws.dbConf == nil || !ws.dbConf.Switcher {
		return nil
	}
	//init db service
	db := tc.NewDBServiceWithPool(
		ws.getDBAddress(),
		ws.dbConf.Pools,
	)
	return db
}

//format db address
func (ws *MysqlSonWorker) getDBAddress() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		ws.dbConf.User, ws.dbConf.Password,
		ws.dbConf.Host, ws.dbConf.Port,
		ws.dbConf.Name,
	)
}

/////////////////////////////////
//private func for MysqlService
/////////////////////////////////

//get son worker by tag
func (w *MysqlService) getSonWorker(tag string) *MysqlSonWorker {
	//get worker
	worker, ok := w.sonWorkers[tag]
	if !ok {
		return nil
	}
	return worker
}

//create son worker
func (w *MysqlService) createSonWorker(tag string, dbConf *MysqlDBConf) bool {
	if tag == "" {
		return false
	}
	w.Lock()
	defer w.Unlock()

	//check
	_, ok := w.sonWorkers[tag]
	if ok {
		return false
	}

	//init son worker
	sonWorker := NewMysqlSonWorker(tag, dbConf)
	w.sonWorkers[tag] = sonWorker

	return true
}
