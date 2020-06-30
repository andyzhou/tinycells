package db

import (
	"github.com/kr/beanstalk"
	"strings"
	"log"
	"time"
)

/*
 * Beanstalk client service
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 * base on github.com/kr/beanstalk
 * https://github.com/asukakenji/cheatsheets/blob/master/beanstalkd.md
 */

 //internal macro define
 const (
 	BeanstalkChanSize = 128
 )

 //received chan data
 type BeanstalkConsumerData struct {
 	QueueId uint64
 	Data []byte
 }

 //beanstalk data
 type BeanstalkData struct {
 	data []byte
 	priority int `default 0`
 	delay int `delay seconds, default 0`
 	ttr int `data up to time`
 }

 //beanstalk client
 type BeanstalkClient struct {
 	tube string
 	address string `server address`
 	receiveChan chan BeanstalkConsumerData `receiver chan from outside`
 	cbForConsumer func(uint64, []byte)bool `callback for consumer`
 	producerConn *beanstalk.Conn `connect for producer`
 	consumerConn *beanstalk.Conn `connect for consumer`
 	needQuit bool
 	dataChan chan BeanstalkData `lazy data chan`
 	closeChan chan bool
 }



//////////////////////////
//api for BeanstalkClient
//////////////////////////

//construct (STEP-1)
//serverAddr like: "127.0.0.1:11300"
func NewBeanstalkClient(
				tube, serverAddr string,
				receiveChan chan BeanstalkConsumerData,
			) *BeanstalkClient {
	this := &BeanstalkClient{
		tube:tube,
		address:serverAddr,
		receiveChan:receiveChan,
		dataChan:make(chan BeanstalkData, BeanstalkChanSize),
		closeChan:make(chan bool),
	}
	//init client
	this.initClient()
	return this
}

//quit
func (c *BeanstalkClient) Quit() {
	//c.needQuit = true
	c.closeChan <- true
}

//set callback for received data from consumer (STEP-2)
func (c *BeanstalkClient) SetCBForConsumer(cb func(uint64,[]byte)bool) {
	c.cbForConsumer = cb
}

//remove queue
func (c *BeanstalkClient) Remove(queueId uint64) bool {
	if queueId <= 0 || c.producerConn == nil {
		return false
	}
	err := c.producerConn.Delete(queueId)
	if err != nil {
		log.Println("BeanstalkClient::Remove for queue id:", queueId, " failed, err:", err.Error())
		return false
	}
	return true
}

//speed up queue
//leftSeconds means total seconds left for queue ready
func (c *BeanstalkClient) SpeedUp(queueId uint64, leftSeconds int) bool {
	if queueId <= 0 || c.producerConn == nil {
		return false
	}
	if leftSeconds < 0 {
		leftSeconds = 0
	}

	delay := time.Duration(leftSeconds) * time.Second
	err := c.producerConn.Release(queueId, 0, delay)
	if err != nil {
		log.Println("BeanstalkClient::SpeedUp for queue:", queueId, " failed, err:", err.Error())
		return false
	}
	return true
}

//producer queue (STEP-3)
//return bool, newQueueId
func (c *BeanstalkClient) Producer(data []byte, priority, ttr int) (bool, uint64) {
	var (
		queueId uint64
	)
	if len(data) <= 0 {
		return false, queueId
	}
	if c.producerConn == nil {
		return false, queueId
	}

	//format data
	priorityInt32 := uint32(priority)
	delayDuration := time.Second * time.Duration(ttr)

	//send to server
	queueId, err := c.producerConn.Put(data, priorityInt32, delayDuration, 0)
	if err != nil {
		log.Println("BeanstalkClient::Producer failed, err:", err.Error())
		return false, queueId
	}

	return true, queueId
}

func (c *BeanstalkClient) ProducerLazy(data []byte, priority, ttr int) bool {
	if len(data) <= 0 {
		return false
	}

	if c.producerConn == nil {
		return false
	}

	//try catch panic
	defer func() {
		if err := recover(); err != nil {
			log.Println("BeanstalkClient::ProducerLazy, panic happened, err:", err)
		}
	}()

	//init lazy data
	lazyData := BeanstalkData{
		data:data,
		priority:priority,
		delay:0,
		ttr:ttr,
	}

	//send to chan
	c.dataChan <- lazyData

	return true
}

////////////////////////////////
//private func for BeanstalkClient
////////////////////////////////

//client init
func (c *BeanstalkClient) initClient() bool {
	var (
		err error
	)

	//try connect server
	//init producer
	producerConn, err := beanstalk.Dial("tcp", c.address)
	if err != nil {
		log.Println("BeanstalkClient::initClient producer connect for ", c.address,
					" failed, err:", err.Error())
		return false
	}

	//init consumer
	consumerConn, err := beanstalk.Dial("tcp", c.address)
	if err != nil {
		log.Println("BeanstalkClient::initClient consumer connect for ", c.address,
					" failed, err:", err.Error())
		return false
	}

	//set tube
	producerConn.Tube.Name = c.tube
	producerConn.TubeSet.Name[c.tube] = true

	consumerConn.Tube.Name = c.tube
	consumerConn.TubeSet.Name[c.tube] = true

	//sync connect
	c.producerConn = producerConn
	c.consumerConn = consumerConn

	//spawn consumer process
	go c.consumeDataProcess()

	//spawn main process
	go c.runMainProcess()

	return true
}

//consume data
func (c *BeanstalkClient) consumeDataProcess() {
	var (
		id uint64
		body []byte
		err error
		substr = "timeout"
	)
	for {
		if c.needQuit {
			break
		}
		//pick ended data from queue
		id, body, err = c.consumerConn.Reserve(time.Second)
		if err != nil {
			if !strings.Contains(err.Error(), substr) {
				log.Println("BeanstalkClient::consumeData of tube:", c.tube,
							" timeout, err:", err.Error())
			}
			continue
		}

		//cast to chan of outside
		consumerData := BeanstalkConsumerData{
			QueueId:id,
			Data:body,
		}
		c.receiveChan <- consumerData

		//remove from queue
		c.consumerConn.Delete(id)
	}
}

//produce data to server
func (c *BeanstalkClient) produceData(data *BeanstalkData) (bool, uint64) {
	var (
		id uint64
	)

	if data == nil {
		return false, id
	}

	delayDuration := time.Duration(data.delay) * time.Second
	ttrDuration := time.Duration(data.ttr) * time.Second
	priorityInt32 := uint32(data.priority)

	//put to server
	id, err := c.producerConn.Put(data.data, priorityInt32, delayDuration, ttrDuration)
	if err != nil {
		log.Println("BeanstalkClient::produceData failed, err:", err.Error())
		return false, id
	}

	return true, id
}

//clean up
func (c *BeanstalkClient) cleanUp() {
	//data clean up
	if c.producerConn != nil {
		c.producerConn.Close()
		c.producerConn = nil
	}
	if c.consumerConn != nil {
		c.consumerConn.Close()
		c.consumerConn = nil
	}

	//close chan
	close(c.dataChan)
	close(c.closeChan)
}

//main process
func (c *BeanstalkClient) runMainProcess() {
	var (
		data BeanstalkData
		needQuit, isOk bool
	)
	for {
		if needQuit && len(c.dataChan) <= 0 {
			break
		}
		select {
		case data, isOk = <- c.dataChan:
			if isOk {
				c.produceData(&data)
			}
		case <- c.closeChan:
			needQuit = true
			c.needQuit = true
		}
	}
	//data clean up
	c.cleanUp()
}
