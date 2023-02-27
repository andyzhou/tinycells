package redis

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v7"
	"sync"
	"time"
)

//face info
type Connection struct {
	client  *redis.Client
	conn *redis.Conn
	config  *Config
	scripts map[string]*redis.Script
	timeout time.Duration
	sync.RWMutex
}

//construct
func NewConnection() *Connection {
	this := &Connection{
		timeout: DefaultTimeOut,
		scripts: map[string]*redis.Script{},
	}
	return this
}

//set timeout
func (f *Connection) SetTimeOut(timeoutSeconds int64) bool {
	if timeoutSeconds <= 0 {
		return false
	}
	f.timeout = time.Duration(timeoutSeconds)
	return true
}

//set config
func (f *Connection) SetConfig(cfg *Config) error {
	if cfg == nil {
		return errors.New("invalid config")
	}
	f.config = cfg
	return nil
}

//disconnect
func (f *Connection) Disconnect() error {
	var (
		err error
	)
	if f.client != nil {
		err = f.client.Close()
	}
	if f.conn != nil {
		err = f.conn.Close()
	}
	return err
}

//connect
func (f *Connection) Connect() error {
	if f.client == nil {
		return errors.New("client hadn't init")
	}
	//ctx, cancel := f.CreateContext()
	//defer cancel()
	conn := f.client.Conn()
	_, err := conn.Ping().Result()
	if err == nil {
		f.conn = conn
	}
	return err
}

func (f *Connection) GetConnect() *redis.Conn {
	return f.conn
}

//run script
func (f *Connection) RunScript(
				name string,
				keys []string,
				args ...interface{},
			) (interface{}, error) {
	script, ok := f.scripts[name]
	if !ok || script == nil {
		return nil, fmt.Errorf("scripter is not exist:%s", name)
	}
	//ctx, cancel := f.CreateContext()
	//defer cancel()
	return script.Run(f.client, keys, args).Result()
}

//add script
func (f *Connection) AddScript(name, script string) error {
	//check
	if name == "" || script == "" {
		return errors.New("invalid parameter")
	}
	if _, ok := f.scripts[name]; ok {
		return fmt.Errorf("ScriptAdd script is exist:%s", name)
	}
	f.scripts[name] = redis.NewScript(script)
	//ctx, cancel := context.WithTimeout(context.Background(), f.timeout*time.Second)
	//defer cancel()
	f.scripts[name].Load(f.client)
	return nil
}

//get client
//func (f *Connection) GetClient() (*redis.Client, context.Context, context.CancelFunc) {
//	ctx, cancel := f.CreateContext()
//	return f.client, ctx, cancel
//}
func (f *Connection) GetClient() *redis.Client {
	return f.client
}

//create context
func (f *Connection) CreateContext() (context.Context, context.CancelFunc){
	return context.WithTimeout(context.Background(), f.timeout*time.Second)
}
