package rpc

import (
	"golang.org/x/net/context"
	"log"
	"io"
	"time"
	"errors"
	"sync"
)

/*
 * RPC service callback for service
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

//internal macro define
const (
	ErrCodeLostParameter = 10000
)

 //rpc callback info
 type GRPCCallBack struct {
 	//callback for service from outside
 	streamCB func(int64,string,string)bool `cb for stream data`
 	//callback for service from outside
 	generalCB func(int64,string,string)string `cb for general data` //input/return packet json string
 	nodeFace *GRPCNode `node interface from outside`
 	sync.Mutex
 }

 //construct
func NewGRPCCallBack(nodeFace *GRPCNode) *GRPCCallBack {
	this := &GRPCCallBack{
		nodeFace:nodeFace,
	}
	return this
}

//set callback for stream request
func (r *GRPCCallBack) SetCBForStream(cb func(int64,string,string)bool) {
	r.streamCB = cb
}

//set callback for general request
func (r *GRPCCallBack) SetCBForGen(cb func(int64,string,string)string) {
	r.generalCB = cb
}

//grpc call back for stream data from client
func (r *GRPCCallBack) StreamReq(stream PacketService_StreamReqServer) error {
	var (
		in *Packet
		err error
		tips string
	)

	//get tag by stream
	tag, ok := RunRpcStat.GetConnTagFromContext(stream.Context())
	if !ok {
		tips = "Can't get tag from node stream."
		log.Println(tips)
		return errors.New(tips)
	}

	log.Println("StreamReq-------tag:", tag)

	//check or sync remote rpc client info
	//r.checkOrAddRemoteStream(tag.RemoteAddr.String(), stream)
	if r.nodeFace != nil {
		r.nodeFace.CheckOrAddRemoteStream(tag.RemoteAddr.String(), stream)
	}

	//try receive stream data from node
	for {
		in, err = stream.Recv()
		log.Println("StreamReq....., in:", in, ", err:", err)
		if err != nil {
			if err == io.EOF {
				log.Println("Read done")
				return nil
			}
			log.Println("Read error:", err.Error())
			return err
		}
		log.Println("in:",in)

		//received real packet data from client, process it.
		//run callback of outside rpc server
		if r.streamCB != nil {
			r.streamCB(in.PlayerId, tag.RemoteAddr.String(), in.Json)
		}
		time.Sleep(time.Second/20)
	}
	return nil
}

 //grpc call back for general request
func (r *GRPCCallBack) SendReq(ctx context.Context, in *Packet) (*Packet, error) {
	var (
		remoteAddr string
	)
	//get key parameter
	packetJsonStr := in.Json

	log.Println("GRPCCallBack::SendReq...")

	if packetJsonStr == "" {
		in.BRet = false
		in.ErrCode = ErrCodeLostParameter
		return in, errors.New("lost parameter data")
	}

	//run callback of outside to process general data
	if r.generalCB != nil {
		//get tag by stream
		tag, ok := RunRpcStat.GetConnTagFromContext(ctx)
		if ok {
			remoteAddr = tag.RemoteAddr.String()
		}
		log.Println("GRPCCallBack::SendReq, remoteAddr:", remoteAddr, ", ok:", ok)
		packetJsonStr = r.generalCB(in.PlayerId, remoteAddr, packetJsonStr)
		in.BRet = true
		in.Json = packetJsonStr
	}

	return in, nil
}

