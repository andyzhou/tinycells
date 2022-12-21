package mysql

import (
	"errors"
	"sync"
)

//face info
type Mysql struct {
	connectMap map[string]*Connect //dbTag -> *Connect
	JsonData
	sync.RWMutex
}

//construct
func NewMysql() *Mysql {
	this := &Mysql{
		connectMap: map[string]*Connect{},
	}
	return this
}

//get connect
func (f *Mysql) GetConnect(tag string) *Connect {
	f.Lock()
	defer f.Unlock()
	v, ok := f.connectMap[tag]
	if ok && v != nil {
		return v
	}
	return nil
}

//create connect
func (f *Mysql) CreateConnect(tag string, conf *Config) (*Connect, error) {
	//check
	if tag == "" || conf == nil {
		return nil, errors.New("invalid parameter")
	}
	//init new connect
	conn := NewConnect(conf)
	f.Lock()
	defer f.Unlock()
	f.connectMap[tag] = conn
	return conn, nil
}

//gen new config
func (f *Mysql)  GenNewConfig() *Config {
	return &Config{}
}