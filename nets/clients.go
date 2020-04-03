package nets

import (
	"sync"
	"net"
	"log"
	"time"
)

/**
 * Running clients and related interface
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

 //internal macro define
 const (
 	ClientMaxActiveSeconds = 120
 	ClientCheckRate = 30
 )

//running clients info
type Clients struct {
	clientMap map[int32]*Client `real clients map`
	closeChan chan bool
	//cb
	cbForLogin func(int32, *net.Conn, []byte)int64 `login callback`
	cbForQuit func(int32,int64) `quit callback`
	cbForRequest func(int64,*ClientPack) `general request callback`
	sync.Mutex `data locker`
}

//construct for clients
func NewClients() *Clients {
	//init self
	this := &Clients{
		clientMap:make(map[int32]*Client),
		closeChan:make(chan bool),
	}
	//spawn main process
	return this
}


////////////////////////
//callback from outside
////////////////////////

//set client login callback
func (cs *Clients) SetLoginCallBack(cb func(int32, *net.Conn, []byte)int64) {
	log.Println("Clients::SetLoginCallBack..")
	cs.cbForLogin = cb
}

//set client quit callback
func (cs *Clients) SetQuitCallBack(cb func(int32,int64)) {
	cs.cbForQuit = cb
}

//set client request data callback
func (cs *Clients) SetRequestCallBack(cb func(int64,*ClientPack)) {
	cs.cbForRequest = cb
}

//////////////////
//API for clients
//////////////////

//clients quit
func (cs *Clients) Quit() {
	cs.closeChan <- true
	for _, client := range cs.clientMap {
		client.Quit()
	}
	time.Sleep(time.Second/10)
}

//cast data to client
func (cs *Clients) CastClient(clientId int32, data []byte) bool {
	if clientId <= 0 || len(data) <= 0 {
		return false
	}
	//try get client instance
	client := cs.getClient(clientId)
	if client == nil {
		return  false
	}
	//begin cast data
	client.CastData(data)
	return true
}

//set player id
func (cs *Clients) SetPlayerId(clientId int32, playerId int64) bool {
	if clientId <= 0 || playerId <= 0 {
		return false
	}
	client, ok := cs.clientMap[clientId]
	if !ok {
		return false
	}
	bRet := client.SetPlayerId(playerId)
	return bRet
}

//set key data id
func (cs *Clients) SetKeyDataId(clientId int32, keyDataId string) bool {
	if clientId <= 0 || keyDataId == "" {
		return false
	}
	client, ok := cs.clientMap[clientId]
	if !ok {
		return false
	}
	bRet := client.SetKeyDataId(keyDataId)
	return bRet
}

//add new original client with origin tcp connect
func (cs *Clients) AddClient(clientId int32, originConn *net.Conn) bool {
	if clientId <= 0 || originConn == nil {
		return false
	}

	//init new client
	client := NewClient(clientId, originConn)
	log.Println("Clients::AddClient....clientId:", clientId)

	//set callback for client
	client.SetLoginCallBack(cs.cbForLogin)
	client.SetRequestCallBack(cs.cbForRequest)
	client.SetQuitCallBack(cs.cbForQuit)

	//add new client into global clients map
	cs.Lock()
	defer cs.Unlock()
	cs.clientMap[clientId] = client

	return true
}

//closed from client side
//just force close client connect and notify service close self
func (cs *Clients) ForceCloseClient(id int32) bool {
	cs.Lock()
	defer cs.Unlock()
	client := cs.getClient(id)
	log.Println("ForceCloseClient, clientId:", id)
	if client == nil {
		return false
	}

	//close client main process
	client.Quit()

	//clear buff
	//client.clientPack.HeaderPack = client.clientPack.HeaderPack[:0]
	//client.clientPack.BodyPack = client.clientPack.BodyPack[:0]

	//try close client connect
	//if client.conn != nil {
	//	log.Println("ForceCloseClient, force close client connect for clientId:", id)
	//	//(*client.conn).Close()
	//	//client.conn = nil
	//}

	//remove from running map
	cs.Lock()
	delete(cs.clientMap, id)
	cs.Unlock()

	return true
}

////////////////////////////////
//private functions for clients
///////////////////////////////

//get client by id
func (cs *Clients) getClient(id int32) *Client {
	if id <= 0 {
		return nil
	}
	if client, ok := cs.clientMap[id]; ok {
		return client
	}
	return nil
}


//data clean up
func (cs *Clients) cleanUp() bool {
	if len(cs.clientMap) <= 0 {
		return false
	}
	for clientId, _ := range cs.clientMap {
		cs.ForceCloseClient(clientId)
	}
	return true
}

//check un active clients
func (cs *Clients) checkUnActiveClients() bool {
	var (
		diff int64
	)

	if len(cs.clientMap) <= 0 {
		return false
	}
	return false

	now := time.Now().Unix()
	for k, v := range cs.clientMap {
		diff = now - v.activeTime
		if diff > ClientMaxActiveSeconds {
			//un-active, need force kick
			cs.ForceCloseClient(k)
		}
	}
	return true
}

//run main process
func (cs *Clients) runMainProcess() {
	var (
		ticker = time.Tick(time.Second * ClientCheckRate)
		needQuit bool
	)
	for {
		if needQuit {
			break
		}
		select {
		case <- ticker:
			cs.checkUnActiveClients()
		case <- cs.closeChan:
			needQuit = true
		}
	}
	log.Println("Clients::runMainProcess, need quit..")
}
