package redis

import (
	"errors"
	"log"
	"sync"
)

//inter data
type (
	PubSubCallback func(interface{})
)

//face info
type PubSub struct {
	conn *Connection //reference
	chanMap map[string]chan struct{} //channel -> chan struct{}
	sync.RWMutex
}

//construct
func NewPubSub() *PubSub {
	this := &PubSub{}
	return this
}

//close sub process
func (f *PubSub) Close() {
	if f.chanMap == nil || len(f.chanMap) <= 0 {
		return
	}
	for _, v := range f.chanMap {
		close(v)
	}
	f.chanMap = map[string]chan struct{}{}
}

//publish message
func (f *PubSub) Publish(channelName string, message interface{}) error {
	//check
	if channelName == "" {
		return errors.New("invalid parameter")
	}
	if f.conn == nil {
		return errors.New("inter conn not init")
	}
	//key opt
	c, ctx, cancel := f.conn.GetClient()
	defer cancel()
	_, err := c.Publish(ctx, channelName, message).Result()
	return err
}

//subscript channel
func (f *PubSub) Subscript(channelName string, cb PubSubCallback) error {
	//check
	if channelName == "" || cb == nil {
		return errors.New("invalid parameter")
	}
	if f.conn == nil {
		return errors.New("inter conn not init")
	}
	f.Lock()
	defer f.Unlock()
	_, ok := f.chanMap[channelName]
	if ok {
		return errors.New("channel had subscript")
	}
	closeChan := make(chan struct{}, 1)
	f.chanMap[channelName] = closeChan

	//run sub process
	sf := func(channelName string, ch chan struct{}) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("PubSub:Subscript channel %v panic, err %v", channelName, err)
			}
			f.Lock()
			defer f.Unlock()
			delete(f.chanMap, channelName)
		}()

		//key opt
		c, ctx, cancel := f.conn.GetClient()
		defer cancel()
		ps := c.Subscribe(ctx, channelName)
		dataChan := ps.Channel()

		//loop
		for {
			select {
			case data, ok := <- dataChan:
				if ok {
					cb(data.Payload)
				}
			case <- ch:
				return
			}
		}
	}
	go sf(channelName, closeChan)
	return nil
}

//set base redis connect
func (f *PubSub) SetConn(conn *Connection) error {
	//check
	if conn == nil {
		return errors.New("invalid parameter")
	}
	f.conn = conn
	return nil
}