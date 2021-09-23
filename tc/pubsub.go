package tc

import (
	"errors"
	"github.com/go-redis/redis"
	"log"
	"sync"
)

/*
 * face of message pub/sub
 */

//inter macro define
const (
	PubSubDelaySeconds = 2
	PublishChanSize = 1024 * 2
	ErrChanSize = 5
)

//pub sub conf
type PubSubConf struct {
	Channel string
	RedisServer string
	RedisPassword string
	RedisDB int
	CB func(data[]byte) bool
}

//publish request
type PublishReq struct {
	channel string
	jsonByte []byte
}

//pub/sub info
type pubSubInfo struct {
	subRedis *GRedis
	closeChan chan bool
}

//face info
type PubSub struct {
	channels map[string]func(data[]byte)bool //channel -> cb func
	subRedis map[string]*pubSubInfo //channel -> pubSubInfo
	publishRedis *GRedis //publish redis instance
	publishChan chan PublishReq
	closeChan chan bool
	closeMainChan chan bool
	sync.RWMutex
}

//construct
func NewPubSub() *PubSub {
	//self init
	this := &PubSub{
		channels:make(map[string]func(data []byte)bool),
		subRedis: make(map[string]*pubSubInfo),
		publishChan:make(chan PublishReq, PublishChanSize),
		closeChan:make(chan bool, 1),
		closeMainChan:make(chan bool, 1),
	}
	//spawn main process
	go this.runMainProcess()
	return this
}


/////////////////////
//implement of IPubSub
/////////////////////

//quit
func (f *PubSub) Quit() {
	//try catch panic
	defer func() {
		if err := recover(); err != nil {
			log.Println("PubSub:Quit panic, err:", err)
		}
	}()
	f.closeChan <- true
	f.closeMainChan <- true
}

//publish message to channel
func (f *PubSub) Publish(channel string, data []byte) (bRet bool) {
	//basic check
	if channel == "" || data == nil || f.publishRedis == nil {
		bRet = false
		return
	}

	//defer
	defer func() {
		if err := recover(); err != nil {
			log.Println("PubSub:Publish panic, err:", err)
			bRet = false
		}
	}()

	//send to chan
	req := PublishReq{
		channel:channel,
		jsonByte:data,
	}
	f.publishChan <- req
	bRet = true

	return
}

//pub/sub one channel with cb func
func (f *PubSub) PubSub(cfg *PubSubConf) bool {
	//basic check
	if cfg == nil || cfg.Channel == "" || cfg.CB == nil || f.subRedis == nil {
		return false
	}

	//get key data
	channel := cfg.Channel

	//check channel cb
	old := f.getCBForChannel(channel)
	if old != nil {
		return false
	}

	//init gRedis
	gRedis, err := f.initGRedis(cfg.RedisServer, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		return false
	}

	//pub/sub channel
	pubSubInfo := f.initPubSubRedis(channel, gRedis)
	if pubSubInfo == nil {
		return false
	}

	//sync channels
	f.channels[channel] = cfg.CB

	//spawn son process for receive
	go f.receiver(pubSubInfo)

	return true
}

//register publish
func (f *PubSub) RegisterPublish(server, password string, db int) error {
	//check
	if server == "" {
		return errors.New("invalid parameter")
	}
	if f.publishRedis != nil {
		return nil
	}

	//init gRedis
	gRedis, err := f.initGRedis(server, password, db)
	if err != nil {
		return err
	}

	//sync
	f.publishRedis = gRedis
	return nil
}

////////////////
//private func
////////////////

//main process for publish
func (f *PubSub) runMainProcess() {
	var (
		req PublishReq
		isOk bool
	)

	//defer
	defer func() {
		if err := recover(); err != nil {
			log.Println("PubSub:runMainProcess panic, err:", err)
		}
		close(f.publishChan)
		close(f.closeMainChan)
	}()

	//loop
	for {
		select {
		case req, isOk = <- f.publishChan:
			if isOk {
				f.publishToChannel(&req)
			}
		case <- f.closeMainChan:
			return
		}
	}
}

//pub sub receiver
func (f *PubSub) receiver(pubSubInfo *pubSubInfo) {
	var (
		message interface{}
		err error
		needReboot, needQuit bool
	)

	//check
	if pubSubInfo == nil {
		return
	}

	//get pub/sub instance
	pubSub := pubSubInfo.subRedis.GetPubSub()
	if pubSub == nil {
		log.Println("PubSub::receiver, get pubSub failed")
		return
	}

	//defer close
	defer func() {
		if err := recover(); err != nil {
			log.Println("PubSub:receiver panic, err:", err)
		}
		pubSub.Close()
	}()

	//create error receive chan
	done := make(chan error, ErrChanSize)

	//loop receive message
	for {
		if needQuit || needReboot {
			break
		}

		select {
		case <- pubSubInfo.closeChan:
			{
				needQuit = true
				break
			}
		default:
			{
				//receive message
				message, err = pubSub.Receive()
				if err != nil {
					log.Println("PubSub::receiver, err:", err.Error())
					done <- err
					break
				}
				//check message
				switch msg := message.(type) {
				case error:
					{
						done <- err
						needReboot = true
						break
					}
				case *redis.Message:
					{
						f.onMessage(msg.Channel, []byte(msg.Payload))
					}
				case redis.Subscription:
					{
						log.Println("PubSub::receiver, count:", msg.Count)
					}
				}
			}
		}
	}

	if len(done) > 0 {
		err, ok := <- done
		if ok && err != nil {
			log.Println("PubSub::receiver, need reboot, err:", err.Error())
		}
	}

	if needReboot {
		log.Println("PubSub::receiver, need reboot..")
		go f.reboot(pubSubInfo)
	}
}

//process pub/sub received message
func (f *PubSub) onMessage(channel string, data []byte) bool {
	//basic check
	if channel == "" || data == nil || f.channels == nil {
		return false
	}

	//get cb for assigned channel
	cb, ok := f.channels[channel]
	if !ok {
		return false
	}

	//dynamic call back
	bRet := cb(data)
	return bRet
}

//publish to channel
func (f *PubSub) publishToChannel(req *PublishReq) bool {
	//basic check
	if req == nil || f.publishRedis == nil {
		return false
	}

	//get pub/sub instance
	client := f.publishRedis.GetClient()
	_, err := client.Publish(req.channel, req.jsonByte).Result()
	if err != nil {
		log.Println("PubSub:publishToChannel failed, err:", err.Error())
		return false
	}

	return true
}

//get cb for channel
func (f *PubSub) getCBForChannel(channel string) func(data []byte) bool {
	//basic check
	if channel == "" || f.channels == nil {
		return nil
	}

	//try hit cb in map
	v, ok := f.channels[channel]
	if !ok {
		return nil
	}
	return v
}

//reboot init
func (f *PubSub) reboot(pubSubInfo *pubSubInfo) {
	if f.channels == nil || pubSubInfo == nil {
		return
	}

	//pub/sub channel
	client := pubSubInfo.subRedis.GetClient()

	//check with locker
	f.Lock()
	defer f.Unlock()
	for channel, _ := range f.channels {
		pubSub := client.Subscribe(channel)
		if pubSub == nil {
			continue
		}
		//set
		pubSubInfo.subRedis.SetPubSub(pubSub)
	}

	//spawn receiver
	go f.receiver(pubSubInfo)
}

//init pub/sub redis
func (f *PubSub) initPubSubRedis(
					channel string,
					subRedis *GRedis,
				) *pubSubInfo {
	//pub/sub channel
	client := subRedis.GetClient()
	pubSub := client.Subscribe(channel)
	if pubSub == nil {
		return nil
	}
	//set pub/sub
	subRedis.SetPubSub(pubSub)
	//sub info
	pubSubInfo := &pubSubInfo{
		subRedis: subRedis,
		closeChan: make(chan bool, 1),
	}
	//sync into env
	f.subRedis[channel] = pubSubInfo
	return pubSubInfo
}

//init gRedis
func (f *PubSub) initGRedis(server, password string, db int) (*GRedis, error) {
	gRedis := NewGRedis(server, password, db)
	if gRedis == nil {
		return nil, errors.New("create redis instance failed")
	}
	//try check redis is worked
	client := gRedis.GetClient()
	_, err := client.Ping().Result()
	if err != nil {
		return nil, err
	}
	return gRedis, nil
}