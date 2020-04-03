package rpc

import (
	"google.golang.org/grpc"
	"context"
	"log"
	"sync"
	"time"
	"io"
)

/*
 * rpc client base interface
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 * - use stream data for communicate
 * - one node one client
 * - use outside call back for received stream data
 */

 //internal macro define
 const (
 	NodeCheckRate = 5
 	NodeDataChanSize = 128
 	MaxTryTimes = 5
 )

 //rpc packet kind
 const (
 	PacketKindReq = iota
 	PacketKindResp
 )

 //rpc mode
 const (
 	ModeOfRpcGen = iota
 	ModeOfRpcStream
 )

 //single rpc client info
 type GRPCClient struct {
	address string `node remote address`
	mode int `rpc mode`
	cb func(packet Packet)bool `call back for received stream data of outside`
	conn *grpc.ClientConn `rpc connect`
	client PacketServiceClient `service client`
	stream PacketService_StreamReqClient `stream packet client`
	receiveChan chan Packet `packet chan for receive outside`
	sendChan chan Packet `packet chan for send`
	closeChan chan bool
	ctx context.Context
	sync.Mutex
 }

 //construct
func NewGRPCClient(address string, mode int) *GRPCClient {
	this := &GRPCClient{
		address:address,
		mode:mode,
		sendChan:make(chan Packet, NodeDataChanSize),
		receiveChan:make(chan Packet, NodeDataChanSize),
		closeChan:make(chan bool),
		ctx:context.Background(),
	}

	//spawn main process
	go this.runMainProcess()

	//try ping server
	go this.ping(false)

	return this
}

///////
//api
///////

//quit
func (n *GRPCClient) Quit() {
	n.closeChan <- true
}

//send general data to remote server
//sync mode
func (n *GRPCClient) SendRequest(req *Packet) (bool, *Packet) {
	if req.Json == "" || n.client == nil {
		return false, nil
	}

	//try catch panic
	defer func() {
		if err := recover(); err != nil {
			log.Println("GRPCClient::SendRequest panic happened, err:", err)
		}
	}()

	//check connect lost or not
	bRet := n.checkServerConn()
	if !bRet {
		log.Println("GRPCClient::SendRequest lost connect...")
		return false, nil
	}

	resp, err := n.client.SendReq(context.Background(), req)
	if err != nil {
		log.Println("GRPCClient::SendRequest failed, err:", err.Error())
		return false, nil
	}

	return true, resp
}


//send stream data to remote server
//async mode
func (n *GRPCClient) SendStreamData(data Packet) bool {
	if data.Json == "" {
		return false
	}

	//try catch panic
	defer func() {
		if err := recover(); err != nil {
			log.Println("NodeInfo::SendData, panic happened, err:", err)
		}
	}()

	//send to data chan
	n.sendChan <- data

	return true
}

//set call back for received stream data of outside
func (n *GRPCClient) SetCallBack(cb func(Packet)bool) {
	n.cb = cb
}

//check server connect
func (n *GRPCClient) CheckConn() bool {
	return n.checkServerConn()
}

/////////////////
//private func
////////////////

//ping remote server
func (n *GRPCClient) ping(isReConn bool) bool {
	var (
		stream PacketService_StreamReqClient
		err error
		isFailed bool
		maxTryTimes int
	)
	if isReConn {
		if n.conn != nil {
			n.conn.Close()
			n.conn = nil
		}
	}

	//check and init rpc connect
	if n.conn == nil {
		//try connect remote server
		conn, err := grpc.Dial(n.address, grpc.WithInsecure())
		if err != nil {
			log.Println("Can't pind ", n.address, ", err:", err.Error())
			return false
		}

		//set rpc client and connect
		n.Lock()
		n.client = NewPacketServiceClient(conn)
		n.conn = conn
		n.Unlock()
	}

	if n.mode == ModeOfRpcGen {
		return true
	}

	//create stream of both side
	for {
		stream, err = n.client.StreamReq(n.ctx)
		if err != nil {
			if maxTryTimes > MaxTryTimes {
				isFailed = true
				break
			}
			log.Println("Create stream with server ", n.address, " failed, err:", err.Error())
			time.Sleep(time.Second)
			maxTryTimes++
			continue
		}
		log.Println("Create stream with server ", n.address, " success")
		break
	}

	if isFailed {
		return false
	}

	//sync stream object
	n.stream = stream

	//spawn stream receiver
	go n.receiveServerStream(stream)

	return true
}


//receive stream data from server
func (n *GRPCClient) receiveServerStream(stream PacketService_StreamReqClient) {
	var (
		in *Packet
		err error
	)

	//try catch panic
	defer func() {
		if err := recover(); err != nil {
			log.Println("GRPCClient::receiveServerStream, panic happened, err:", err)
		}
	}()

	log.Println("GRPCClient::receiveServerStream.....")

	//receive data use for loop
	for {
		in, err = stream.Recv()
		if err != nil {
			if err == io.EOF {
				continue
			}
			log.Println("Receive data failed, err:", err.Error())
			break
		}

		//send data to receive chan of outside
		if n.receiveChan != nil {
			log.Println("GRPCClient::receiveServerStream, in00:", in)
			n.receiveChan <- *in
		}
	}
}

//send data to remote server
func (n *GRPCClient) sendDataToServer(data *Packet) bool {
	log.Println("RpcNode::sendDataToServer, data:", data)
	if data == nil || n.stream == nil {
		return false
	}
	err := n.stream.Send(data)
	if err != nil {
		log.Println("GRPCClient::sendDataToServer, send data failed, err:", err.Error())
		return false
	}
	log.Println("GRPCClient::sendDataToServer, success")
	return true
}

//check remote server status
func (n *GRPCClient) checkServerStatus() {
	var (
		needPing bool
	)
	//log.Println("GRPCClient::checkServerStatus...")
	if n.conn == nil {
		//try reconnect
		needPing = true
	}else{
		state := n.conn.GetState().String()
		if state == "TRANSIENT_FAILURE" || state == "SHUTDOWN" {
			needPing = true
		}
	}
	if needPing {
		n.ping(true)
	}
}

//check remote connect is lost or not
func (n *GRPCClient) checkServerConn() bool {
	state := n.conn.GetState().String()
	if state == "TRANSIENT_FAILURE" || state == "SHUTDOWN" {
		return false
	}
	return true
}

//run main process
func (n *GRPCClient) runMainProcess() {
	var (
		ticker = time.Tick(time.Second * NodeCheckRate)
		data Packet
		needQuit, isOk bool
	)
	for {
		if needQuit && len(n.sendChan) <= 0 {
			break
		}
		select {
		case data, isOk = <- n.sendChan:
			if isOk {
				//try send to remote server
				n.sendDataToServer(&data)
			}
		case data, isOk = <- n.receiveChan:
			if isOk {
				log.Println("GRPCClient, received data:", data)
				//run callback func to process received data
				n.cb(data)
			}
		case <- ticker:
			//check server status
			n.checkServerStatus()
		case <- n.closeChan:
			needQuit = true
		}
	}
	log.Println("GRPCClient::runMainProcess of ", n.address, " need quit..")
}
