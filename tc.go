package tinycells

import (
	"errors"
	"github.com/andyzhou/tinycells/config"
	"sync"
)

//global variable
var (
	_tc *TinyCells
	_tcOnce sync.Once
)

//interface
type TinyCells struct {
	cfg *config.Config
}

//get single instance
func GetTC() *TinyCells {
	_tcOnce.Do(func() {
		_tc = NewTinyCells()
	})
	return _tc
}

//construct
func NewTinyCells() *TinyCells {
	this := &TinyCells{}
	return this
}

//////////////////////////////////////////////
//setup for first init
//this should called before use sub instance
//////////////////////////////////////////////

//setup config
func (f *TinyCells) SetConfig(params ...interface{}) error {
	if f.cfg != nil {
		return errors.New("config instance had setup")
	}
	f.cfg = config.NewConfig(params...)
	return nil
}

///////////////////////
//get sub instance
///////////////////////

//get config
func (f *TinyCells) GetConfig() *config.Config {
	return f.cfg
}