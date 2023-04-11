package tinycells

import (
	"errors"
	"github.com/andyzhou/tinycells/cmd"
	"github.com/andyzhou/tinycells/config"
	"github.com/andyzhou/tinycells/crypt"
	"github.com/andyzhou/tinycells/db"
	"github.com/andyzhou/tinycells/logger"
	"github.com/andyzhou/tinycells/sys"
	"github.com/andyzhou/tinycells/web"

	//"github.com/andyzhou/tinycells/mq"
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
	//mq *mq.MQ
	wb *web.Web
	db *db.DB
	single *sys.Signal
	logger *logger.Logger
	cmd *cmd.Cmd
	crypt *crypt.Crypt
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
		//mq: mq.NewMQ(),
		wb: web.NewWeb(),
		db: db.NewDB(),
		single: sys.NewSignal(),
		logger: logger.NewLogger(),
		cmd: cmd.NewCmd(),
		crypt: crypt.NewCrypt(),
		util: util.NewUtil(),
	}
	return this
}

//////////////////////////////////////////////
//setup for first init
//this should called before use sub instance
//////////////////////////////////////////////

//setup logger
func (f *TinyCells) SetUpLogger(params ...interface{}) error {
	if params == nil || len(params) < 0 {
		return errors.New("invalid parameter")
	}
	config, ok := params[0].(*logger.Config)
	if !ok || config == nil {
		return errors.New("invalid logger config")
	}
	f.logger.SetConfig(config)
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

////get mq
//func (f *TinyCells) GetMQ() *mq.MQ {
//	return f.mq
//}

//get single
func (f *TinyCells) GetSingle() *sys.Signal {
	return f.single
}

//get web
func (f *TinyCells) GetWeb() *web.Web {
	return f.wb
}

//get db
func (f *TinyCells) GetDB() *db.DB {
	return f.db
}

//get logger
func (f *TinyCells) GetLogger() *logger.Logger {
	return f.logger
}

//get cmd
func (f *TinyCells) GetCmd() *cmd.Cmd {
	return f.cmd
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