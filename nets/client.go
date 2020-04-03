package nets

import (
	"github.com/andyzhou/tinycells/tc"
	"log"
	"time"
	"net"
	"sync"
	"math"
)

/**
 * client data and interface
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

const (
	LoginMaxTryTimes = 3
	MaxTcpBuffSize = 1024
	CastChanSize = 64
	ActiveCheckRate = 30
	//UseLostConnectionErr = "use of closed network connection"
)

//real client info
type Client struct {
	id int32 `original client id`
	playerId int64 `logon player id`
	token string `game token`
	keyDataId string `key data id`
	conn *net.Conn `client tcp original connect`
	tcpBuff []byte `shared tcp buff`
	clientPack *ClientPack `shared client whole pack`
	castChan chan []byte `chan for service data cast`
	closeChan chan bool
	activeTime int64 `client last active time`
	tryTimes int `try times for first login`
	//cb
	cbForLogin func(int32, *net.Conn, []byte) int64 `login callback`
	cbForQuit func(int32,int64) `quit callback`
	cbForRequest func(int64,*ClientPack) `general request callback`
	Packet `anonymous instance`
	sync.Mutex `data locker`
	tc.Utils
}


//construct for client
func NewClient(clientId int32, originConn *net.Conn) *Client {
	//init client pack
	clientPack := &ClientPack{
		HeaderPack:make([]byte, 0),
		BodyPack:make([]byte, 0),
	}

	//init new client instance
	this := &Client{
		id:clientId,
		conn:originConn,
		clientPack:clientPack,
		tcpBuff:make([]byte, MaxTcpBuffSize),
		castChan:make(chan []byte, CastChanSize),
		closeChan:make(chan bool),
		activeTime:time.Now().Unix(),
	}

	//run client currency process for handle tcp data
	go this.runMainProcess()
	go this.handleTcpDataProcess()

	return this
}


////////////////////////
//callback from outside
////////////////////////

//set client login callback
func (c *Client) SetLoginCallBack(cb func(int32, *net.Conn, []byte)int64) {
	log.Println("Client::SetLoginCallBack...")
	c.cbForLogin = cb
}

//set client quit callback
func (c *Client) SetQuitCallBack(cb func(int32,int64)) {
	log.Println("Client::SetQuitCallBack...")
	c.cbForQuit = cb
}

//set client request data callback
func (c *Client) SetRequestCallBack(cb func(int64,*ClientPack)) {
	c.cbForRequest = cb
}


////////////////////
//API for client
///////////////////

//client quit
func (c *Client) Quit() {
	//process quit gentleman
	defer func() {
		if err := recover(); err != nil {
			log.Fatalln("Client::Quit, fatal happend, err:", err)
		}
	}()
	c.closeChan <- true
	time.Sleep(time.Second/10)
}


//cast data to client
func (c *Client) CastData(data []byte) bool {
	if len(data) <= 0 {
		return false
	}

	//process cast date gentleman
	defer func() {
		if err := recover(); err != nil {
			log.Fatalln("Client::CastData, fatal happened, err:", err)
		}
	}()

	//cast data to channel
	c.castChan <- data
	return true
}

//set player id
func (c *Client) SetPlayerId(playerId int64) bool {
	if playerId <= 0 {
		return false
	}
	c.Lock()
	defer c.Unlock()
	c.playerId = playerId
	return true
}

//set key data id
func (c *Client) SetKeyDataId(keyDataId string) bool {
	if keyDataId == "" {
		return false
	}
	c.Lock()
	defer c.Unlock()
	c.keyDataId = keyDataId
	return true
}

//get client id
func (c *Client) GetId() int32 {
	return c.id
}


////////////////////////////////
//private functions for client
///////////////////////////////

//clean up for closed client
func (c *Client) cleanUp() bool {
	if c.playerId <= 0 {
		return false
	}
	//sync player running data and clean up
	//mod.RunPlayers.PlayerQuit(c.playerId)
	return true
}

//send byte data to client
func (c *Client) sendToClient(data []byte) bool {
	var (
		needQuit bool
		timeoutDuration = time.Second
	)

	//basic check
	if len(data) <= 0 || c.conn == nil {
		return false
	}

	//try send data to client side with timeout
	(*c.conn).SetWriteDeadline(time.Now().Add(timeoutDuration))
	size, err := (*c.conn).Write([]byte(data))
	if err != nil {
		log.Println("Client::sendToClient, send failed, clientId:", c.id, ", size:", size, ", err:", err)
		needQuit = c.CheckTcpError(err)
		if needQuit {
			//run client quit callback from outside
			go c.cbForQuit(c.id, c.playerId)
		}
	}

	return true
}


//heart beat check
//func (c *Client) checkHeartBeat() bool {
//	activeTimeConf := int64(conf.RunPlazaConfig.GetBasic().GetClientActiveTime())
//	now := time.Now().Unix()
//	diff := now - c.activeTime
//	if diff > activeTimeConf {
//		//need close un-active client
//		RunClients.ForceCloseClient(c.id)
//	}else{
//		//cast current server time to client?
//		//TODO..
//	}
//	return true
//}

//client main process
func (c *Client) runMainProcess() {
	var (
		needQuit bool
		data = make([]byte, 0)
		ticker = time.Tick(time.Second * ActiveCheckRate)
		isOk bool
	)
	for {
		if needQuit && len(c.castChan) <= 0 {
			break
		}
		select {
		case data, isOk = <- c.castChan:
			if isOk {
				c.sendToClient(data)
			}
		case <- ticker:
			//heart beat check
			//c.checkHeartBeat()
		case <- c.closeChan:
			//mark close and clean up
			needQuit = true
			c.cleanUp()
		}
	}

	//close related chan
	close(c.castChan)
	close(c.closeChan)

	log.Println("Client::runMainProcess, client:", c.id, " main process need quit..")
}

//handle client tcp data process
func (c *Client) handleTcpDataProcess() {
	var (
		playerId int64
		//arenaId string
		needQuit bool
	)

	for {
		if c.conn == nil {
			log.Println("Client::handleTcpDataProcess... connect is null")
			break
		}

		//read and analyze packet header
		needQuit = c.analyzeHeader()
		if needQuit {
			log.Println("Client::handleTcpDataProcess, analyze header failed, clientId:", c.id, ", playerId:", c.playerId)
			break
		}

		////get command id
		//commandId = c.clientPack.CommandId
		//if commandId <= 0 {
		//	//no any header, need continue
		//	log.Println("Client::handleTcpDataProcess, invalid header info, clientId:", c.id, ", playerId:", c.playerId)
		//	log.Println("Client::handleTcpDataProcess, commandId:", commandId)
		//	continue
		//}

		//continue read body
		//log.Println("begin analyze body, commandId:", commandId, ", bodySize:", c.clientPack.BodySize)
		needQuit = c.analyzeBody(c.clientPack.BodySize)
		if needQuit {
			log.Println("Client::handleTcpDataProcess, analyze body failed, clientId:", c.id, ", playerId:", c.playerId)
			break
		}

		//need check player logon or not
		if c.playerId <= 0 {
			//check try times
			if c.tryTimes >= LoginMaxTryTimes {
				//force close tcp connect
				break
			}
			//need unpack login packet
			//run callback from outside
			//this is sync operate!!
			playerId = c.cbForLogin(c.id, c.conn, c.clientPack.BodyPack)
			log.Println("player login, playerId:", playerId)
			if playerId <= 0 {
				c.tryTimes++
			}else{
				//sync player id
				c.Lock()
				c.playerId = playerId
				c.Unlock()
			}
		} else {
			//just cast internal data
			//cast request to related gate
			//skip heart beat packet
			//run callback from outside
			//RunGameGate.SyncGate(internalPacket)
			c.clientPack.PlayerId = c.playerId
			c.cbForRequest(c.playerId, c.clientPack)
		}
	}

	//cleanup related data
	c.clientPack.HeaderPack = c.clientPack.HeaderPack[:0]
	c.clientPack.BodyPack = c.clientPack.BodyPack[:0]
	c.tcpBuff = c.tcpBuff[:0]

	//some error happened
	log.Println("Client ", c.id, " disconnect.")

	//force close client
	//run client quit callback from outside
	//RunClients.ForceCloseClient(c.id)
	//c.cbForQuit(c.id, c.playerId)
}

//analyze packet body
func (c *Client) analyzeBody(bodySize int) bool {
	//continue read body
	var (
		err error
		needQuit bool
		size, readSize int
		i int
		left = bodySize
		blocks = int(math.Ceil(float64(bodySize)/MaxTcpBuffSize))
	)

	//reset body pack buff
	c.clientPack.BodyPack = c.clientPack.BodyPack[:0]

	//log.Println("Client::analyzeBody, bodySize:", bodySize, ", blocks:", blocks)
	for {
		if left <= 0 || i >= 2 * blocks {
			break
		}

		if left > MaxTcpBuffSize {
			size, err = (*c.conn).Read(c.tcpBuff)
		}else{
			size, err = (*c.conn).Read(c.tcpBuff[0:left])
		}
		//log.Println("handleTcpDataProcess, blocks:", blocks, ", i:", i, ", size:", size, ", left:", left, " err:", err)

		if err != nil {
			log.Println("Client ", c.id, " read data failed, err:", err.Error())
			//special error of net
			needQuit = c.CheckTcpError(err)
			if needQuit {
				log.Println("Lost connect with client:", c.id, " for read")
				break
			}
		}
		if size > 0 {
			c.clientPack.BodyPack = append(c.clientPack.BodyPack, c.tcpBuff[0:size]...)
			readSize += size
		}
		left -= size
		i++
	}

	if needQuit || readSize == 0 {
		//reset temp buff
		return needQuit
	}

	//check temp buff
	if readSize < bodySize {
		//read dirty data, force close client
		log.Println("Read dirty data, force close client:", c.id)
		needQuit = true
		return needQuit
	}

	//copy body packet
	//if !needQuit {
	//	*wholePacket = append(*wholePacket, c.clientPack.BodyPack...)
	//	//log.Println("wholePacket02:", *wholePacket)
	//}

	//get function packet
	return needQuit
}

//analyze packet header
func (c *Client) analyzeHeader() bool {
	var (
		size int
		needQuit bool
		err error
	)

	//reset shared client pack
	c.clientPack.CommandId = 0
	c.clientPack.BodySize = 0
	c.clientPack.HeaderPack = c.clientPack.HeaderPack[:0]

	//read packet header
	header := c.tcpBuff[0:c.GetPackHeadSize()]
	size, err = (*c.conn).Read(header)

	log.Println("client::analyzeHeader, header:", header)
	log.Println("client::analyzeHeader, size:", size, ", err:", err)

	if err != nil {
		log.Println("client::analyzeHeader, Client ", c.id, " read data failed, err:", err.Error())
		//special error of net
		needQuit = c.CheckTcpError(err)
		if needQuit {
			header = header[:0]
			return needQuit
		}
	}

	//read empty header
	if size == 0 {
		header = header[:0]
		return needQuit
	}

	//update client active time
	c.activeTime = time.Now().Unix()

	//analyze packet header
	bodySize := c.DecodePackHeader(header)
	log.Println("client::analyzeHeader, bodySize:", bodySize)

	//copy into shared head pack
	//c.clientPack.CommandId = commandId
	c.clientPack.BodySize = bodySize
	c.clientPack.HeaderPack = append(c.clientPack.HeaderPack, header...)

	//clean header buff
	header = header[:0]
	return needQuit
}

