package redis

import "time"

type Config struct {
	DBTag 	string 		   `yaml:"dbTag" json:"dbTag"`
	DBNum       int    	   `yaml:"dbNum" json:"dbNum"`
	Addr     string        `yaml:"addr" json:"addr"`
	Password string        `yaml:"password" json:"password"`
	PoolSize int    `yaml:"poolSize" json:"poolSize"`
	TimeOut  time.Duration `yaml:"timeout" json:"timeout"`
}
