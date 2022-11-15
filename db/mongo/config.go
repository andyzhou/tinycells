package mongo

type Config struct {
	UserName string `yaml:"userName" json:"userName"`
	Password string `yaml:"password" json:"password"`
	DBUrl string	`yaml:"dbUrl" json:"dbUrl"`
	DBName string	`yaml:"dbName" json:"dbName"`
	PoolSize int    `yaml:"poolSize" json:"poolSize"`
}