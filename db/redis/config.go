package redis

import "time"

type Config struct {
	DBTag 	string 		   `yaml:"dbTag" json:"dbTag"`
	DBNum       int    	   `yaml:"dbNum" json:"dbNum"`
	Addr     string        `yaml:"addr" json:"addr"`
	Password string        `yaml:"password" json:"password"`
	TimeOut  time.Duration `yaml:"timeout" json:"timeout"`
}
