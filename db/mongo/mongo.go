package mongo

import (
	"errors"
	"sync"
)

//face info
type Mongo struct {
	connMap map[string]*Connection //dbName -> *Connection
	sync.RWMutex
}

//construct
func NewMongo() *Mongo {
	this := &Mongo{
		connMap: map[string]*Connection{},
	}
	return this
}

//access connect
func (f *Mongo) C(dbName string) *Connection {
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
func (f *Mongo) CreateConn(cfg *Config) (*Connection, error) {
	//check
	if cfg == nil {
		return nil, errors.New("invalid redis db config")
	}
	//init new
	connect := NewConnection(cfg)
	err := connect.Connect()
	if err != nil {
		return nil, err
	}
	err = connect.Ping()
	if err != nil {
		return nil, err
	}

	//sync into run env
	f.Lock()
	defer f.Unlock()
	f.connMap[cfg.DBName] = connect
	return connect, nil
}

//gen new config
func (f *Mongo) GenNewConfig() *Config {
	return &Config{}
}