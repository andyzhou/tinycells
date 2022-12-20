package db

import (
	"github.com/andyzhou/tinycells/db/mongo"
	"github.com/andyzhou/tinycells/db/mysql"
	"github.com/andyzhou/tinycells/db/redis"
)

//face info
type DB struct {
	mysql *mysql.Mysql
	mongo *mongo.Mongo
	redis *redis.Redis
}

//construct
func NewDB() *DB {
	this := &DB{
		mysql: mysql.NewMysql(),
		mongo: mongo.NewMongo(),
		redis: redis.NewRedis(),
	}
	return this
}

//get sub instance
func (f *DB) GetMysql() *mysql.Mysql {
	return f.mysql
}

func (f *DB) GetMongo() *mongo.Mongo {
	return f.mongo
}

func (f *DB) GetRedis() *redis.Redis {
	return f.redis
}
