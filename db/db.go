package db

import (
	"github.com/andyzhou/tinycells/db/mongo"
	"github.com/andyzhou/tinycells/db/redis"
)

//face info
type DB struct {
	mongo *mongo.Mongo
	redis *redis.Redis
}

//construct
func NewDB() *DB {
	this := &DB{
		mongo: mongo.NewMongo(),
		redis: redis.NewRedis(),
	}
	return this
}

//get sub instance
func (f *DB) GetMongo() *mongo.Mongo {
	return f.mongo
}

func (f *DB) GetRedis() *redis.Redis {
	return f.redis
}
