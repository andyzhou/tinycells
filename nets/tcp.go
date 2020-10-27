package nets

import (
	"log"
	"math"
	"net"
)

/*
 * TCP Service with grace support
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

//internal macro define
const (
	MaxClientIdPow = 30
)

//tcp service struct
type TcpService struct {
	port int `tcp port`
	tcpKind int `tcp kind`
	ids int32 `client id generator`
	maxIds int32 `max id value`
	listener net.Listener `net listener`
	grace *Grace `grace instance`
	clientsObj *Clients `clients object`
}

//construct
func NewTcpService(port, tcpKind int) *TcpService {
	//self init
	maxPowVal := int32(math.Pow(2, MaxClientIdPow))
	this := &TcpService{
		port:port,
		tcpKind:tcpKind,
		ids:0,
		maxIds:maxPowVal,
	}
	//init grace
	this.graceInit(tcpKind)
	return this
}


/////////
//api
////////

//quit
func (s *TcpService) Quit() {
	s.grace.Quit()
	if s.clientsObj != nil {
		s.clientsObj.Quit()
	}
	s.listener.Close()
}

//set clients object
func (s *TcpService) SetClients(obj *Clients)  {
	s.clientsObj = obj
}

//start tcp server
func (s *TcpService) Start() {
	log.Println("TcpService::Start, begin..")
	defer s.listener.Close()

	//accept client in loop
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			log.Fatalln("connet failed", err)
			continue
		}

		//if running ids value exceed max
		//need reset to zero
		if s.ids >= s.maxIds {
			s.ids = 0
		}

		//generate new client id
		s.ids += 1
		newClientId := s.ids

		log.Println("TcpService::Start, newClientId:", newClientId)

		//add into running clients
		if s.clientsObj != nil {
			s.clientsObj.AddClient(newClientId, &conn)
		}
	}
}

///////////////
//private func
//////////////

func (s *TcpService) graceInit(tcpKind int) {
	//init grace
	s.grace = NewGrace(s.port, tcpKind)

	//set attached for tcp
	s.grace.SetAttached(true)

	//get listener for tcp service
	s.listener = s.grace.GetListener()

	//start grace
	go s.grace.Start()
}
