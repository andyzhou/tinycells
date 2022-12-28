package kafka

import (
	"errors"
	"fmt"
	"github.com/Shopify/sarama"
	"log"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"
)

/*
 * face of producer
 */
//face info
type Producer struct {
	serverAddr []string
	conf *sarama.Config
	producer *sarama.SyncProducer
	topicMap map[string]*sarama.ProducerMessage
	produceChan chan ProduceReq
	closeChan chan bool
	counter atomic.Value
	sync.RWMutex
}

//construct
func NewProducer() *Producer {
	this := &Producer{
		serverAddr: []string{},
		topicMap: map[string]*sarama.ProducerMessage{},
		produceChan: make(chan ProduceReq, ProducerChanSize),
		closeChan: make(chan bool, 1),
		counter: atomic.Value{},
	}
	this.interInit()
	return this
}

//quit
func (f *Producer) Quit() {
	f.closeChan <- true
}

//set chan
func (f *Producer) SetChan(chanSize int32) error {
	if chanSize <= 0 {
		return errors.New("invalid parameter")
	}
	f.produceChan = make(chan ProduceReq, chanSize)
	return nil
}

//set address
func (f *Producer) SetAddress(address []string) error {
	if len(address) <= 0 {
		return errors.New("invalid parameter")
	}
	f.serverAddr = append(f.serverAddr, address...)
	return nil
}

//start
func (f *Producer) Start() error {
	//check
	if len(f.serverAddr) <= 0 {
		return errors.New("producer server address is nil")
	}

	//check process counter
	v := f.counter.Load()
	counterInt, _ := v.(int)
	if counterInt > 0 {
		return errors.New("producer process had started")
	}

	//init new producer
	producer, err := sarama.NewSyncProducer(f.serverAddr, f.conf)
	if err != nil {
		return err
	}
	f.producer = &producer

	//spawn main process
	go f.runMainProcess()

	//init counter
	f.counter.Store(1)
	return nil
}

//send message
func (f *Producer) SendMessage(topic, key, message string) error {
	//check
	if topic == "" || key == "" {
		return errors.New("invalid parameter")
	}
	//init request
	req := ProduceReq{
		Topic: topic,
		Key: key,
		Val: message,
	}
	//async send to chan
	select {
	case f.produceChan <- req:
	}
	return nil
}

///////////////
//private func
///////////////

//run main process
func (f *Producer) runMainProcess() {
	var (
		req ProduceReq
		isOk bool
	)

	//defer
	defer func() {
		if err := recover(); err != nil {
			log.Println("kafka.Producer panic, err:", err)
			log.Println("kafka.Producer trace:", string(debug.Stack()))
		}
		//close relate obj
		(*f.producer).Close()
		close(f.closeChan)
		//decr counter
		f.counter.Store(0)
	}()

	//loop check
	for {
		select {
		case req, isOk = <- f.produceChan:
			if isOk {
				//send message
				f.sendMessage(&req)
			}
		case <- f.closeChan:
			return
		}
	}
}

//send message
func (f *Producer) sendMessage(req *ProduceReq) error {
	//check
	if req == nil || req.Key == "" || req.Val == "" {
		return errors.New("invalid parameter")
	}
	//get topic
	topic := f.getOrInitTopic(req.Topic, req.Key)
	if topic == nil {
		return fmt.Errorf("init topic %v, %v failed", req.Topic, req.Key)
	}
	//init and send message
	log.Printf("kafka.Producer:sendMessage, topic:%v, key:%v, val:%v\n", req.Topic, req.Key, req.Val)
	topic.Value = sarama.StringEncoder(req.Val)
	(*f.producer).SendMessage(topic)
	return nil
}

//get or init topic
func (f *Producer) getOrInitTopic(topic, key string) *sarama.ProducerMessage {
	//check
	if topic == "" || key == "" {
		return nil
	}
	//init unique key
	uniqueKey := fmt.Sprintf("%s:%s", topic, key)

	//check cache with locker
	f.Lock()
	defer f.Unlock()
	v, ok := f.topicMap[uniqueKey]
	if ok {
		return v
	}

	//init and sync new
	v = &sarama.ProducerMessage{
		Topic:topic,
		Key:sarama.StringEncoder(key),
	}
	f.topicMap[uniqueKey] = v
	return v
}

//init config
func (f *Producer) initConf() {
	conf := sarama.NewConfig()
	conf.Producer.Return.Successes = true
	conf.Producer.Return.Errors = true
	conf.Producer.Timeout = time.Second * TimeOut
	conf.Version = sarama.V2_2_0_0
	f.conf = conf
}

//inter init
func (f *Producer) interInit() {
	//init config
	f.initConf()
}