package kafka

import "github.com/Shopify/sarama"

//kafka config
type KaConf struct {
	producer *sarama.Config
	consumer *sarama.Config
}

//dynamic consumer
type DynamicConsumer struct {
	Topic string
	CB func([]byte, []byte) bool //cb of outside, func(key, value []byte)bool
	CloseChan chan bool //used for dynamic process
}

//produce request
type ProduceReq struct {
	Topic string
	Key string
	Val string
}
