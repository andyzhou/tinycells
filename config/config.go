package config

//main face
type Config struct {
	ini *IniConfig
	json *JsonConfig
}

//construct
func NewConfig(params ...interface{}) *Config {
	//get key param
	cfgRootPath := "."
	if params != nil && len(params) > 0 {
		v, ok := params[0].(string)
		if ok && v != "" {
			cfgRootPath = v
		}
	}
	//self init
	this := &Config{
		ini: NewIniConfigWithPara(cfgRootPath),
		json: NewJsonConfigWithPara(cfgRootPath),
	}
	return this
}

//get sub face
func (c *Config) GetIniConf() *IniConfig {
	return c.ini
}

func (c *Config) GetJsonConf() *JsonConfig {
	return c.json
}