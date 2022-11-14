package redis

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"sync"
	"time"
)

//face info
type Connection struct {
	client  *redis.Client
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

//disconnect
func (f *Connection) Disconnect() error {
	if f.client == nil {
		return nil
	}
	return f.client.Close()
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
	ctx, cancel := f.CreateContext()
	defer cancel()
	return script.Run(ctx, f.client, keys, args).Result()
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
	ctx, cancel := context.WithTimeout(context.Background(), f.timeout*time.Second)
	defer cancel()
	f.scripts[name].Load(ctx, f.client)
	return nil
}

//get client
func (f *Connection) GetClient() (*redis.Client, context.Context, context.CancelFunc) {
	ctx, cancel := f.CreateContext()
	return f.client, ctx, cancel
}

//create context
func (f *Connection) CreateContext() (context.Context, context.CancelFunc){
	return context.WithTimeout(context.Background(), f.timeout*time.Second)
}
