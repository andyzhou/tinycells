package kafka

import (
	"errors"
)

/*
 * face of kafka
 *  - base on `https://github.com/Shopify/sarama`
 */

//face info
type Kafka struct {
	producer *Producer
	consumer *Consumer
}

//construct, step-1
func NewKafka() *Kafka {
	this := &Kafka{
		producer: NewProducer(),
		consumer: NewConsumer(),
	}
	return this
}

//set relate face, step-2
func (f *Kafka) SetAddress(address []string) error {
	if len(address) <= 0 {
		return errors.New("invalid parameter")
	}
	f.producer.SetAddress(address)
	f.consumer.SetAddress(address)
	return nil
}

//start, step-3
func (f *Kafka) Start() error {
	err := f.producer.Start()
	if err != nil {
		return err
	}
	err = f.consumer.Start()
	return err
}

//quit
func (f *Kafka) Quit() {
	f.producer.Quit()
	f.consumer.Quit()
}

//get relate face
func (f *Kafka) GetProducer() *Producer {
	return f.producer
}
func (f *Kafka) GetConsumer() *Consumer {
	return f.consumer
}