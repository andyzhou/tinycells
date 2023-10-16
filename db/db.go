package db

import (
	"github.com/andyzhou/tinycells/db/mongo"
	"github.com/andyzhou/tinycells/db/mysql"
	"github.com/andyzhou/tinycells/db/redis"
	"github.com/andyzhou/tinycells/db/sqlite"
)

//face info
type DB struct {
	mongo *mongo.Mongo
	mysql *mysql.Mysql
	redis *redis.Redis
	sqlite *sqlite.SqlLite
}

//construct
func NewDB() *DB {
	this := &DB{
		mongo: mongo.NewMongo(),
		mysql: mysql.NewMysql(),
		redis: redis.NewRedis(),
		sqlite: sqlite.NewSqlLite(),
	}
	return this
}

//get sub instance
func (f *DB) GetMongo() *mongo.Mongo {
	return f.mongo
}
func (f *DB) GetMysql() *mysql.Mysql {
	return f.mysql
}
func (f *DB) GetRedis() *redis.Redis {
	return f.redis
}
func (f *DB) GetSqlite() *sqlite.SqlLite {
	return f.sqlite
}