package main

import (
	"github.com/andyzhou/tinycells"
	"github.com/urfave/cli"
	"log"
)

func main() {
	redisExample()
}

//redis
func redisExample() {
	//get sub instance
	tc := tinycells.GetTC()
	rd := tc.GetDB().GetRedis()

	//set key data
	dbName := "base"

	//gen config
	config := rd.GenNewConfig()
	config.Addr = "127.0.0.1:6380"
	config.DBTag = dbName
	config.DBNum = 0

	//create connect
	_, err := rd.CreateConn(config)
	if err != nil {
		log.Printf("connect redis failed, err:%v", err)
		return
	}
	defer rd.C(dbName).Disconnect()

	//get keys
	rc, ctx, _ := rd.C(dbName).GetClient()
	keys, err := rc.Keys(ctx, "*").Result()
	log.Printf("keys:%v, err:%v", keys, err)
}

//mongo
func mongoExample() {
	//get sub instance
	tc := tinycells.GetTC()
	mgo := tc.GetDB().GetMongo()

	//set key data
	dbName := "battle_artist_v2"
	dbCol := "clients"

	//gen config
	config := mgo.GenNewConfig()
	config.DBName = dbName
	config.DBUrl = "mongodb://127.0.0.1:27017/battle_artist_v2"

	//create connect
	_, err := mgo.CreateConn(config)
	if err != nil {
		log.Printf("connect mongo failed, err:%v", err)
		return
	}
	defer mgo.C(dbName).Disconnect()
	count, err := mgo.C(dbName).Count(dbCol, nil)
	log.Printf("get count:%v, err:%v", count, err)
}

//cmd
func cmdExample() {
	//setup argv name
	argNameOfName := "name"

	//get sub instance
	tc := tinycells.GetTC()
	cmd := tc.GetCmd()

	//register arg name
	err := cmd.RegisterStringFlag(argNameOfName)
	if err != nil {
		log.Printf("err:%v", err.Error())
		return
	}
	sf := func(c *cli.Context) error {
		nameVal := c.String(argNameOfName)
		log.Printf("nameVal:%v", nameVal)
		return nil
	}
	err = cmd.InitApp(sf)
	if err != nil {
		log.Printf("err:%v", err.Error())
		return
	}
	err = cmd.StartApp()
	if err != nil {
		log.Printf("err:%v", err.Error())
		return
	}
	log.Printf("init app end")
}

//logger
func loggerExample()  {
	tc := tinycells.GetTC()
	logger := tc.GetLogger()
	config := logger.BuildDefaultConfig()
	err := logger.SetConfig(config)
	logger.SS().Infof("test logger")
	log.Printf("err:%v", err)
}