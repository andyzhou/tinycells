package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"sync"
)

/*
 * json format config file processor
 */

//define config struct
type JsonConfig struct {
	cfgRootPath string
	kv map[string]interface{}
	sync.RWMutex
}

//construct
func NewJsonConfig() *JsonConfig {
	return NewJsonConfigWithPara(".")
}

func NewJsonConfigWithPara(cfgRootPath string) *JsonConfig {
	return &JsonConfig{
		cfgRootPath:cfgRootPath,
		kv:make(map[string]interface{}),
	}
}

//gt value as slice
func (c *JsonConfig) GetConfigAsSlice(key string) []interface{} {
	ret := make([]interface{}, 0)
	v := c.GetConfig(key)
	if v == nil {
		return ret
	}
	if value, ok := v.([]interface{}); ok {
		ret = value
	}
	return ret
}

//get value as map[string]interface{}
func (c *JsonConfig) GetConfigAsMap(key string) map[string] interface{} {
	ret := make(map[string]interface{})
	v := c.GetConfig(key)
	if v == nil {
		return ret
	}
	if value, ok := v.(map[string]interface{}); ok {
		ret = value
	}
	return ret
}


//get value as bool
func (c *JsonConfig) GetConfigAsBool(key string) bool {
	v := c.GetConfig(key)
	if v == nil {
		return false
	}
	if value, ok := v.(bool); ok {
		return value
	}
	return false
}

//get value as integer
func (c *JsonConfig) GetConfigAsInteger(key string) int {
	v := c.GetConfig(key)
	if v == nil {
		return 0
	}

	var ret string
	switch v.(type) {
	case float64:
		ret = fmt.Sprintf("%1.0f", v)
	case float32:
		ret = fmt.Sprintf("%1.0f", v)
	case string:
		ret = fmt.Sprintf("%s", v)
	default:
		ret = fmt.Sprintf("%d", v)
	}
	val, err := strconv.Atoi(ret)
	if err != nil {
		fmt.Println(err.Error())
		return 0
	}
	return val
}

//get value as string
func (c *JsonConfig) GetConfigAsString(key string) string {
	v := c.GetConfig(key)
	if v == nil {
		return ""
	}
	if value, ok := v.(string); ok {
		return value
	}
	return ""
}

//get single k/v
func (c *JsonConfig) GetConfig(key string) interface{} {
	//map k/v fetch
	if v, ok := c.kv[key];ok{
		return v
	}
	return nil
}

//get all config
func (c *JsonConfig) GetAllConfigs() map[string]interface{} {
	return c.kv
}

//load config
func (c *JsonConfig) LoadConfig(fileName string) error {
	bytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		return err
	}
	err = json.Unmarshal(bytes, &c.kv)
	return err
}
