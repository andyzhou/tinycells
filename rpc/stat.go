package rpc

import (
	"context"
	"google.golang.org/grpc/stats"
	"log"
	"sync"
)

/*
 * RPC stat handler for service
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 * Need apply `TagConn`, `TagRPC`, `HandleConn`, `HandleRPC` methods.
 */

//connect ctx key info
type connCtxKey struct{}

type GRPCStat struct {
	nodeFace *GRPCNode
	connMap map[*stats.ConnTagInfo]string
	sync.Mutex
}

//declare global variable
var RunRpcStat *GRPCStat

//construct
func NewGRPCStat(nodeFace *GRPCNode) *GRPCStat {
	this := &GRPCStat{
		nodeFace:nodeFace,
		connMap:make(map[*stats.ConnTagInfo]string),
	}
	return this
}

//declare global variables
//var connMutex sync.Mutex
//var connMap = make(map[*stats.ConnTagInfo]string)

///////
//api
///////

//clean up
func (h *GRPCStat) CleanUp() {
	h.Lock()
	defer h.Unlock()
	for k, _ := range h.connMap {
		delete(h.connMap, k)
	}
}

//get connect tag
func (h *GRPCStat) GetConnTagFromContext(ctx context.Context) (*stats.ConnTagInfo, bool) {
	tag, ok := ctx.Value(connCtxKey{}).(*stats.ConnTagInfo)
	return tag, ok
}

//cb for rpc api
func (h *GRPCStat) TagConn(ctx context.Context, info *stats.ConnTagInfo) context.Context {
	log.Println("TagConn, from address:", info.RemoteAddr)
	return context.WithValue(ctx, connCtxKey{}, info)
}

//cb for rpc api
func (h *GRPCStat) TagRPC(ctx context.Context, info *stats.RPCTagInfo) context.Context {
	log.Println("TagRPC, method name:", info.FullMethodName)
	return ctx
}

//cb for rpc api
func (h *GRPCStat) HandleConn(ctx context.Context, s stats.ConnStats) {
	//log.Println("HandleConn")

	//get tag current ctx
	tag, ok := h.GetConnTagFromContext(ctx)
	if !ok {
		log.Fatal("can not get conn tag")
	}

	switch s.(type) {
	case *stats.ConnBegin:
		//connMap[tag] = ""
		h.connMap[tag] = ""
		log.Printf("begin conn, tag = (%p)%#v, now connections = %d\n", tag, tag, len(h.connMap))
	case *stats.ConnEnd:
		delete(h.connMap, tag)
		log.Printf("end conn, tag = (%p)%#v, now connections = %d\n", tag, tag, len(h.connMap))
		//run node face to remove end connect
		if h.nodeFace != nil {
			h.nodeFace.RemoveStream(tag.RemoteAddr.String())
		}
	default:
		log.Printf("illegal ConnStats type\n")
	}
}

//cb for rpc api
func (h *GRPCStat) HandleRPC(ctx context.Context, s stats.RPCStats) {
	//fmt.Println("HandleRPC, IsClient:", s.IsClient())
}

