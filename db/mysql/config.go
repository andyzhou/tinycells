package mysql

type Config struct {
	Host 		string 	`yaml:"host" json:"host"`
	Port 		int 	`yaml:"port" json:"port"`
	User 		string 	`yaml:"user" json:"user"`
	Password 	string 	`yaml:"password" json:"password"`
	DBName 		string	`yaml:"dbName" json:"dbName"`
	PoolSize 	int    	`yaml:"poolSize" json:"poolSize"`
}