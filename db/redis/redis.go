package redis

import (
	"errors"
	"github.com/go-redis/redis"
	"sync"
)

//face info
type Redis struct {
	connMap map[string]*Connection //dbTag -> *Connection
	pubSub *PubSub
	sync.RWMutex
}

//construct
func NewRedis() *Redis {
	this := &Redis{
		connMap: map[string]*Connection{},
		pubSub: NewPubSub(),
	}
	return this
}

//get pub sub instance
func (f *Redis) GetPubSub() *PubSub {
	return f.pubSub
}

//access connect
func (f *Redis) C(dbName string) *Connection {
	if dbName == "" {
		return nil
	}
	f.Lock()
	defer f.Unlock()
	v, ok := f.connMap[dbName]
	if ok && v != nil {
		return v
	}
	return nil
}

//create new connect
func (f *Redis) CreateConn(cfg *Config) (*Connection, error) {
	//check
	if cfg == nil {
		return nil, errors.New("invalid redis db config")
	}
	//check and release old
	v, ok := f.connMap[cfg.DBTag]
	if ok && v != nil {
		v.Disconnect()
	}
	//init new
	connect := NewConnection()
	connect.config = cfg
	connect.client = redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DBNum,
	})
	//sync into run env
	f.Lock()
	defer f.Unlock()
	f.connMap[cfg.DBTag] = connect
	return connect, nil
}

//gen new config
func (f *Redis) GenNewConfig() *Config {
	return &Config{}
}