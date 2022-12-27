package mq

import (
	"github.com/andyzhou/tinycells/mq/kafka"
	"sync"
)

/*
 * message queue face
 */

//global variable
var (
	_mq *MQ
	_mqOnce sync.Once
)

//face info
type MQ struct {
	kafka *kafka.Kafka
}

//get single instance
func GetMQ() *MQ {
	_mqOnce.Do(func() {
		_mq = NewMQ()
	})
	return _mq
}

//construct
func NewMQ() *MQ {
	this := &MQ{
		kafka: kafka.NewKafka(),
	}
	return this
}

//get relate face
func (f *MQ) GetKafka() *kafka.Kafka {
	return f.kafka
}