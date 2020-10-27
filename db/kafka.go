package db

import (
	"fmt"
	"github.com/Shopify/sarama"
	"log"
	"sync"
	"time"
)

/*
 * Kafka service interface
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 * - base on `https://github.com/Shopify/sarama`
 */

 //inter macro define
 const (
 	KafkaTimeOut = 5 //xx seconds
 	KafkaProducerChanSize = 128
 )

 //kafka config
 type KafkaConf struct {
 	producer *sarama.Config
 	consumer *sarama.Config
 }

 //producer request chan
 type KafkaProduceReq struct {
 	topic string
 	key string
 	val string
 }

 //dynamic consumer info
 type KafkaDynamicConsumer struct {
 	topic string
 	cb func([]byte)[]byte `cb of outside`
 	closeChan chan bool `used for dynamic process`
 }

 //kafka info
 type Kafka struct {
 	addr []string `server address`
 	conf *KafkaConf `basic config`
 	producer *sarama.AsyncProducer
 	consumer *sarama.Consumer
 	topicMap map[string]*sarama.ProducerMessage `producer topic map`
 	consumerMap map[string]*KafkaDynamicConsumer `dynamic consumer map`
 	produceChan chan KafkaProduceReq
 	produceCloseChan chan bool
 	sync.Mutex
 }
 
 //construct
func NewKafka(address []string) *Kafka {
	//self init
	this := &Kafka{
		addr:address,
		topicMap:make(map[string]*sarama.ProducerMessage),
		consumerMap:make(map[string]*KafkaDynamicConsumer),
		produceChan:make(chan KafkaProduceReq, KafkaProducerChanSize),
		produceCloseChan:make(chan bool),
	}

	//inter init
	this.interInit()

	return this
}

//quit
func (k *Kafka) Quit() {
	k.produceCloseChan <- true
	for _, dc := range k.consumerMap {
		dc.closeChan <- true
	}
}

//register call back for consumer
//cb function like cb(key, value []byte)bool, used for outside
func (k *Kafka) RegisterConsumer(topic string, cb func([]byte)[]byte) bool {
	//basic check
	if topic == "" || cb == nil {
		return false
	}

	//check
	_, ok := k.consumerMap[topic]
	if ok {
		return true
	}

	//init new
	dc := &KafkaDynamicConsumer{
		topic:topic,
		cb:cb,
		closeChan:make(chan bool),
	}
	k.initDynamicConsumer(dc)
	return true
}

//send message
func (k *Kafka) SendMessage(topic, key, message string) bool {
	if topic == "" || key == "" {
		return false
	}

	//try catch panic
	defer func() {
		if err := recover(); err != nil {
			log.Println("Kafka::SendMessage panic happened, err:", err)
		}
	}()

	//init request
	req := KafkaProduceReq{
		topic:topic,
		key:key,
		val:message,
	}
	k.produceChan <- req
	return true
}

//////////////
//private func
//////////////

//internal testing
func (k *Kafka) testing() {
	topic := "test-1"
	key := "k-1-0"
	message := "hello"

	//send message
	k.SendMessage(topic, key, message)
}

//send single message
func (k *Kafka) sendMessage(req *KafkaProduceReq) bool {
	if req == nil {
		return false
	}

	//get topic object
	topic := k.getOrInitTopic(req.topic, req.key)
	if topic == nil {
		return false
	}

	//init message
	topic.Value = sarama.StringEncoder(req.val)

	//send to chain
	(*k.producer).Input() <- topic

	return true
}

//init dynamic consumer
func (k *Kafka) initDynamicConsumer(dc *KafkaDynamicConsumer) bool {
	if dc == nil || k.consumer == nil {
		return false
	}

	//init partition consumer
	pConsumer, err := (*k.consumer).ConsumePartition(dc.topic, 0, sarama.OffsetNewest)
	if err != nil {
		log.Println("Kafka::initDynamicConsumer failed, err:", err.Error())
		return false
	}

	//spawn son process
	go k.runConsumerProcess(dc, &pConsumer)

	//sync into map
	k.Lock()
	defer k.Unlock()
	k.consumerMap[dc.topic] = dc

	return true
}

//sub consumer process
func (k *Kafka) runConsumerProcess(dc *KafkaDynamicConsumer, pc *sarama.PartitionConsumer) {
	var (
		message *sarama.ConsumerMessage
		err error
		needQuit, isOk bool
	)

	//defer close
	defer (*pc).Close()

	//wait message
	for {
		if needQuit {
			break
		}
		select {
		case message, isOk = <- (*pc).Messages():
			if isOk {
				key := string(message.Key)
				value := string(message.Value)
				log.Println("Kafka::runConsumerProcess, message:", message)
				log.Println("Kafka::runConsumerProcess, key:", key, ", val:", value)
			}
		case err = <- (*pc).Errors():
			{
				log.Println("Kafka::runConsumerProcess, err:", err.Error())
			}
		case <- dc.closeChan:
			{
				needQuit = true
			}
		}
	}
}

//producer main process
func (k *Kafka) runProducerProcess() {
	var (
		req KafkaProduceReq
		needQuit, isOk bool
	)

	//close resource
	defer (*k.producer).AsyncClose()

	for {
		if needQuit && len(k.produceChan) <= 0 {
			break
		}
		select {
		case req, isOk = <- k.produceChan:
			if isOk {
				//send message
				k.sendMessage(&req)
			}
		case <- k.produceCloseChan:
			needQuit = true
		}
	}
}

//get or init topic
func (k *Kafka) getOrInitTopic(topic, key string) *sarama.ProducerMessage {
	//init unique key
	uniqueKey := fmt.Sprintf("%s:%s", topic, key)

	//check cache
	v, ok := k.topicMap[uniqueKey]
	if ok {
		return v
	}

	//init new with locker
	k.Lock()
	defer k.Unlock()
	v = &sarama.ProducerMessage{
		Topic:topic,
		Key:sarama.StringEncoder(key),
	}

	//sync into map
	k.topicMap[uniqueKey] = v

	return v
}

//inter init
func (k *Kafka) interInit() {
	//init kafka config
	k.conf = k.initKafkaConf()

	//init producer
	k.initProducer()

	//init consumer
	k.initConsumer()

	//inter testing
	time.AfterFunc(time.Second * 3, k.testing)
}

//init consumer
func (k *Kafka) initConsumer() bool {
	//init
	consumer, err := sarama.NewConsumer(k.addr, k.conf.consumer)
	if err != nil {
		log.Println("Kafka::initConsumer failed, err:", err.Error())
		return false
	}
	k.consumer = &consumer

	return true
}

//init producer
func (k *Kafka) initProducer() bool {
	//init
	producer, err := sarama.NewAsyncProducer(k.addr, k.conf.producer)
	if err != nil {
		log.Println("Kafka::initProducer failed, err:", err.Error())
		return false
	}
	k.producer = &producer

	//spawn main process
	go k.runProducerProcess()

	return true
}

//init kafka config
func (k *Kafka)initKafkaConf() *KafkaConf{
	//init producer config
	producerConf := sarama.NewConfig()
	producerConf.Producer.Return.Successes = true
	producerConf.Producer.Return.Errors = true
	producerConf.Producer.Timeout = time.Second * KafkaTimeOut
	producerConf.Version = sarama.V2_2_0_0

	//init consumer config
	consumerConf := sarama.NewConfig()
	consumerConf.Consumer.Return.Errors = true
	consumerConf.Consumer.Offsets.Initial = sarama.OffsetNewest
	consumerConf.Version = sarama.V2_2_0_0

	//init config
	conf := &KafkaConf{
		producer:producerConf,
		consumer:consumerConf,
	}
	return conf
}

