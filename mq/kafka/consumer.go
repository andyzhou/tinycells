package kafka

import (
	"errors"
	"github.com/Shopify/sarama"
	"log"
	"runtime/debug"
	"sync"
)

/*
 * consumer face
 */

//face info
type Consumer struct {
	serverAddr []string
	conf *sarama.Config
	rootConsumer *sarama.Consumer
	consumerMap map[string]*DynamicConsumer //topic -> DynamicConsumer
	sync.RWMutex
}

//construct
func NewConsumer() *Consumer {
	this := &Consumer{
		serverAddr: []string{},
		consumerMap: map[string]*DynamicConsumer{},
	}
	this.interInit()
	return this
}

//quit
func (f *Consumer) Quit() {
	if f.rootConsumer != nil {
		(*f.rootConsumer).Close()
		f.rootConsumer = nil
	}
}

//remove consumer
func (f *Consumer) RemoveConsumer(topic string) error {
	//basic check
	if topic == "" {
		return errors.New("invalid parameter")
	}
	//check
	f.Lock()
	defer f.Unlock()
	v, ok := f.consumerMap[topic]
	if !ok || v == nil {
		return errors.New("no such topic consumer")
	}
	//close and remove it
	v.CloseChan <- true
	delete(f.consumerMap, topic)
	return nil
}

//register call back for consumer
//cb function like cb(key, value []byte)bool, used for outside
func (f *Consumer) RegisterConsumer(topic string, cb func([]byte, []byte) bool) error {
	//basic check
	if topic == "" || cb == nil {
		return errors.New("invalid parameter")
	}

	//check and init with locker
	f.Lock()
	_, ok := f.consumerMap[topic]
	f.Unlock()
	if ok {
		//had registered
		return nil
	}

	//init new dynamic consumer
	dc := &DynamicConsumer{
		Topic: topic,
		CB: cb,
		CloseChan: make(chan bool, 1),
	}
	err := f.initDynamicConsumer(dc)
	return err
}

//start
func (f *Consumer) Start() error {
	//check
	if len(f.serverAddr) <= 0 {
		return errors.New("consumer server address is nil")
	}
	//init root consumer
	err := f.initRootConsumer()
	return err
}

//set address
func (f *Consumer) SetAddress(address []string) error {
	if len(address) <= 0 {
		return errors.New("invalid parameter")
	}
	f.serverAddr = append(f.serverAddr, address...)
	return nil
}

///////////////
//private func
///////////////

//sub consumer process
func (f *Consumer) runConsumerProcess(dc *DynamicConsumer, pc *sarama.PartitionConsumer) {
	var (
		message *sarama.ConsumerMessage
		isOk bool
		err error
	)

	//defer
	defer func() {
		if err := recover(); err != nil {
			log.Println("kafka.Consumer:runConsumerProcess panic, err:", err)
			log.Println("kafka.Consumer:runConsumerProcess trace:", string(debug.Stack()))
		}
		(*pc).Close()
		log.Println("kafka.Consumer:runConsumerProcess end..")
	}()

	//loop wait message
	log.Println("kafka.Consumer:runConsumerProcess loop..")
	for {
		select {
		case message, isOk = <- (*pc).Messages():
			if isOk {
				//dynamic run cb func
				log.Printf("kafka.Consumer:runConsumerProcess, topic:%v, keys:%v, message:%v\n",
							dc.Topic, string(message.Key), string(message.Value))
				if dc.CB != nil {
					dc.CB(message.Key, message.Value)
				}
			}
		case err = <- (*pc).Errors():
			{
				log.Printf("kafka.Consumer:runConsumerProcess, topic:%v, err:%v\n",  dc.Topic, err.Error())
			}
		case <- dc.CloseChan:
			{
				return
			}
		}
	}
}

//init dynamic consumer
func (f *Consumer) initDynamicConsumer(dc *DynamicConsumer) error {
	//check
	if dc == nil || f.rootConsumer == nil {
		return errors.New("invalid parameter or root consumer is nil")
	}

	//init partition consumer
	pConsumer, err := (*f.rootConsumer).ConsumePartition(dc.Topic, 0, sarama.OffsetNewest)
	log.Printf("kafka.Consumer:initDynamicConsumer, topic:%v, err:%v", dc.Topic, err)
	if err != nil {
		return err
	}

	//spawn son process
	go f.runConsumerProcess(dc, &pConsumer)
	log.Printf("kafka.Consumer:initDynamicConsumer, topic:%v, sync", dc.Topic)

	//sync into dynamic map
	f.Lock()
	defer f.Unlock()
	f.consumerMap[dc.Topic] = dc
	log.Printf("kafka.Consumer:initDynamicConsumer, topic:%v, done", dc.Topic)
	return nil
}

//init root consumer
func (f *Consumer) initRootConsumer() error {
	consumer, err := sarama.NewConsumer(f.serverAddr, f.conf)
	if err != nil {
		return err
	}
	f.rootConsumer = &consumer
	return nil
}

//init config
func (f *Consumer) initConf() {
	conf := sarama.NewConfig()
	conf.Consumer.Return.Errors = true
	conf.Consumer.Offsets.Initial = sarama.OffsetNewest
	conf.Version = sarama.V2_2_0_0
	f.conf = conf
}

//inter init
func (f *Consumer) interInit() {
	//init config
	f.initConf()
}