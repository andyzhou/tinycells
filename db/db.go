package db

import (
	"github.com/andyzhou/tinycells/db/redis"
)

//face info
type DB struct {
	redis *redis.Redis
}

//construct
func NewDB() *DB {
	this := &DB{
		redis: redis.NewRedis(),
	}
	return this
}

//get sub instance
func (f *DB) GetRedis() *redis.Redis {
	return f.redis
}

////setup redis
//func (f *DB) SetupRedis(params ...interface{}) error {
//	if params == nil || len(params) <= 0 {
//		return errors.New("invalid parameter")
//	}
//	f.redis.CreateConnect()
//}