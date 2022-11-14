package tinycells

import (
	"errors"
	"github.com/andyzhou/tinycells/config"
	"github.com/andyzhou/tinycells/crypt"
	"github.com/andyzhou/tinycells/db"
	"github.com/andyzhou/tinycells/util"
	"sync"
)

//global variable
var (
	_tc *TinyCells
	_tcOnce sync.Once
)

//interface
type TinyCells struct {
	crypt *crypt.Crypt
	db *db.DB
	cfg *config.Config
	util *util.Util
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
	this := &TinyCells{
		crypt: crypt.NewCrypt(),
		util: util.NewUtil(),
	}
	return this
}

//////////////////////////////////////////////
//setup for first init
//this should called before use sub instance
//////////////////////////////////////////////

//setup db
func (f *TinyCells) SetupDB(params ...interface{}) error {
	if f.db != nil {
		return errors.New("db instance had setup")
	}
	f.db = db.NewDB()
	return nil
}

//setup config
func (f *TinyCells) SetupConfig(params ...interface{}) error {
	if f.cfg != nil {
		return errors.New("config instance had setup")
	}
	f.cfg = config.NewConfig(params...)
	return nil
}

///////////////////////
//get sub instance
///////////////////////

//get db
func (f *TinyCells) GetDB() *db.DB {
	return f.db
}

//get config
func (f *TinyCells) GetConfig() *config.Config {
	return f.cfg
}

//get crypt
func (f *TinyCells) GetCrypt() *crypt.Crypt {
	return f.crypt
}

//get util
func (f *TinyCells) GetUtil() *util.Util {
	return f.util
}