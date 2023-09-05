package config


import (
	"log"
	"os"
	"time"
)

/*
 * single json config process face
 */

//inter macro define
const (
	ConfCheckConfRate = 30 //seconds
)

//sub config info
type SubConfig struct {
	confFile string
	confCheckRate int
	cbForAnalyze func(map[string]interface{}) bool `CB for analyze config`
	conf *JsonConfig `config instance`
	confMap map[string]interface{}
	lastTime int64 `last update time`
	closeChan chan bool
}

//construct
func NewSubConfig(
			confFile string,
			cb func(map[string]interface{}) bool,
			checkRate ... int,
		) *SubConfig {
	//self init
	this := &SubConfig{
		confFile:confFile,
		cbForAnalyze:cb,
		conf:NewJsonConfig(),
		confMap:make(map[string]interface{}),
		closeChan:make(chan bool, 1),
	}

	//get and set check rate
	if checkRate != nil && len(checkRate) > 0 {
		this.confCheckRate = checkRate[0]
	}else{
		this.confCheckRate = ConfCheckConfRate
	}

	//pre load config
	this.preLoadConfig()

	//spawn main process
	go this.runMainProcess()

	return this
}

//quit
func (c *SubConfig) Quit() {
	c.closeChan <- true
}

//get map data
func (c *SubConfig) GetConfMap() map[string]interface{} {
	return c.confMap
}

///////////////
//private func
///////////////

//pre load all config
func (c *SubConfig) preLoadConfig() bool {
	//begin load config
	err := c.conf.LoadConfig(c.confFile)
	if err != nil {
		log.Printf("SubConfig::preLoadConfig failed, file:%v, error:%v\n", c.confFile, err.Error())
		return false
	}

	//get all data
	c.confMap = c.conf.GetAllConfigs()

	//run call back
	if c.cbForAnalyze != nil && c.confMap != nil {
		c.cbForAnalyze(c.confMap)
	}

	return true
}

//check config stat
func (c *SubConfig) checkFileStat() bool {
	if c.confFile == "" {
		return false
	}

	//check config file modify time
	modifyTime := c.getFileModifyTime(c.confFile)
	if modifyTime <= 0 || modifyTime <= c.lastTime {
		return false
	}

	//need reload new config file
	c.preLoadConfig()

	//update last modify time
	c.lastTime = modifyTime

	return true
}


//main process
func (c *SubConfig) runMainProcess() {
	//set ticker
	tickDuration := time.Duration(c.confCheckRate) * time.Second
	ticker := time.Tick(tickDuration)
	neeQuit := false

	defer func() {
		if err := recover(); err != nil {
			log.Println("SubConfig:mainProcess panic, err:", err)
		}
		//close chan
		close(c.closeChan)
	}()

	//loop
	for {
		if neeQuit {
			break
		}
		select {
		case <- ticker:
			c.checkFileStat()
		case <- c.closeChan:
			neeQuit = true
		}
	}
}

//check file stat and last modify time
func (c *SubConfig) getFileModifyTime(filePath string) int64 {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return 0
	}
	modifyTime := fileInfo.ModTime().Unix()
	return modifyTime
}