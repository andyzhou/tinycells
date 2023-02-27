package mongo

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type (
	D                       = bson.D
	E                       = bson.E
	M                       = bson.M
	Client                  = mongo.Client
	UpdateOptions           = options.UpdateOptions
	FindOptions             = options.FindOptions
	FindOneOptions          = options.FindOneOptions
	InsertOneOptions        = options.InsertOneOptions
	InsertManyOptions       = options.InsertManyOptions
	DeleteOptions           = options.DeleteOptions
	FindOneAndUpdateOptions = options.FindOneAndUpdateOptions
	WriteModel              = mongo.WriteModel
	BulkWriteResult         = mongo.BulkWriteResult
	UpdateResult            = mongo.UpdateResult
	Cursor                  = mongo.Cursor
	BulkWriteException      = mongo.BulkWriteException
	UpdateOneModel          = mongo.UpdateOneModel
)

type BulkWriteOp struct {
	WriteModel   []WriteModel
	BulkWriteOpt *options.BulkWriteOptions
}

//index key info
type MGIndex struct {
	Key interface{} `bson:"key" json:"key"`
	Name string `bson:"name" json:"name"`
}