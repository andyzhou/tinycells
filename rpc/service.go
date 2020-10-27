package rpc

import (
	"fmt"
	"github.com/andyzhou/tinycells/nets"
	"google.golang.org/grpc"
	"log"
	"net"
)

/*
 * rpc service interface
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

//rpc service info
type GRPCService struct {
	port int `rpc port`
	address string `rpc service address`
	listener net.Listener `tcp listener`
	service *grpc.Server
	rpcStat *GRPCStat `inter rpc stat`
	rpcCB *GRPCCallBack `inter rpc service callback`
	grace *nets.Grace `grace instance`
}

//construct (STEP1)
func NewGRPCService(port int, nodeFace *GRPCNode) *GRPCService {
	address := fmt.Sprintf(":%d", port)
	this := &GRPCService{
		port:port,
		address:address,
		service:nil,
	}

	//init internal rpc stat
	this.rpcStat = NewGRPCStat(nodeFace)

	//init internal rpc callback
	this.rpcCB = NewGRPCCallBack(nodeFace)

	//grace init
	this.graceInit()

	return this
}

////////
//api
///////

//quit
func (r *GRPCService) Quit() {
	if r.service != nil {
		r.service.Stop()
		log.Println("rpc service stopped.")
	}
}

//start service (STEP3)
func (r *GRPCService) Start() {
	//create rpc service
	r.createService()
}

//set server callback for stream request (STEP2-1)
func (r *GRPCService) SetCBForStream(cb func(int64,string,string)bool) {
	r.rpcCB.SetCBForStream(cb)
}

//set server callback for general request (STEP2-2)
func (r *GRPCService) SetCBForGeneral(cb func(int64,string,string)string) {
	r.rpcCB.SetCBForGen(cb)
}

////////////////
//private func
///////////////

//create rpc service
func (r *GRPCService) createService() {
	//var (
	//	tips string
	//	err error
	//)

	//try listen tcp port
	//listen, err := net.Listen("tcp", r.address)
	//if err != nil {
	//	tips = "Create rpc service failed, error:" + err.Error()
	//	log.Println(tips)
	//	panic(tips)
	//}
	//log.Println("Create rpc service success")

	//create rpc server with rpc stat support
	r.service = grpc.NewServer(grpc.StatsHandler(r.rpcStat))

	//register call back
	//cb := NewGRPCCallBack()
	RegisterPacketServiceServer(r.service, r.rpcCB)

	//begin rpc service
	go r.beginService(r.listener)
}

//begin rpc service
func (r *GRPCService) beginService(listen net.Listener) {
	//service listen
	err := r.service.Serve(listen)
	if err != nil {
		tips := "Failed for rpc service, error:" + err.Error()
		log.Println(tips)
		panic(tips)
	}
}

//grace init
func (r *GRPCService) graceInit() {
	//init grace
	r.grace = nets.NewGrace(r.port, nets.TcpKindGen)

	//get listener for tcp service
	r.listener = r.grace.GetListener()

	//start grace
	go r.grace.Start()
}
