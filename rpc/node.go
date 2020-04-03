package rpc

import (
	"sync"
	"log"
)

/*
 * Node interface for service
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 * used for manage remote rpc client address of stream mode
 */

 //node info
 type GRPCNode struct {
 	remoteStreams map[string]PacketService_StreamReqServer `remoteAddr:stream`
 	sync.Mutex
 }

 //construct
func NewGRPCNode() *GRPCNode {
	this := &GRPCNode{
		remoteStreams:make(map[string]PacketService_StreamReqServer),
	}
	return this
}

/////////
//api
////////

//quit
func (r *GRPCNode) Quit() {
	r.Lock()
	defer r.Unlock()
	for k, _ := range r.remoteStreams {
		delete(r.remoteStreams, k)
	}
}

//cast packet to all streams
func (r *GRPCNode) CastToAll(packet Packet) bool {
	var (
		err error
	)
	//basic check
	if packet.Json == "" || len(r.remoteStreams) <= 0 {
		return false
	}
	for _, stream := range r.remoteStreams {
		err = stream.Send(&packet)
		if err != nil {
			log.Println("GRPCNode::CastToAll, send failed, err:", err.Error())
		}
	}
	return true
}

//cast packet to single node
func (r *GRPCNode) CastToNode(remoteAddr string, packet Packet) bool {
	stream, ok := r.remoteStreams[remoteAddr]
	if !ok {
		return false
	}
	stream.Send(&packet)
	return true
}

//cast packet to assigned remote streams
func (r *GRPCNode) CastToNodes(nodes map[string]bool, packet Packet) bool {
	var (
		stream PacketService_StreamReqServer
		isOk bool
		err error
	)

	//basic check
	if len(nodes) <= 0 || packet.Json == "" || len(r.remoteStreams) <= 0 {
		return false
	}

	//pick and send
	for node, _ := range nodes {
		stream, isOk = r.remoteStreams[node]
		if !isOk {
			continue
		}
		//send data pass stream
		err = stream.Send(&packet)
		if err != nil {
			log.Println("GRPCNode::CastToNodes, send failed, err:", err.Error())
		}
	}
	return true
}

//clean up
func (r *GRPCNode) CleanUp() {
	r.Lock()
	defer r.Unlock()
	for k, _ := range r.remoteStreams {
		delete(r.remoteStreams, k)
	}
}

//remove stream
func (r *GRPCNode) RemoveStream(remoteAddr string) bool {
	if remoteAddr == "" {
		return false
	}
	r.Lock()
	defer r.Unlock()
	delete(r.remoteStreams, remoteAddr)
	return true
}

//check or add remote client stream info
func (r *GRPCNode) CheckOrAddRemoteStream(remoteAddr string, stream PacketService_StreamReqServer) bool {
	if remoteAddr == "" || stream == nil {
		return false
	}
	_, ok := r.remoteStreams[remoteAddr]
	if ok {
		return false
	}
	//add new record with locker
	r.Lock()
	defer r.Unlock()
	r.remoteStreams[remoteAddr] = stream
	return true
}
