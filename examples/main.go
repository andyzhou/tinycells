package main

import (
	"github.com/andyzhou/tinycells"
	"github.com/andyzhou/tinycells/web"
	"github.com/urfave/cli/v2"
	"log"
	"sync"
	"time"
)

func main() {
	var (
		wg sync.WaitGroup
	)

	cmdExample()
	loggerExample()
	mongoExample()
	redisExample()
	mysqlExample()
	webAppExample()

	//wait
	wg.Add(1)
	wg.Wait()
}

//web app
func webAppExample()  {
	app := web.NewApp()
	app.SetTplPath("./tpl")
	app.RegisterSubApp("/", NewSubApp())
	go app.Start(8090)
}

//kafka
//func kafkaExample() {
//	var (
//		wg sync.WaitGroup
//		recvTimes int
//		sendMaxTimes int
//	)
//
//	//topic and key setup
//	topic := "topic-1"
//	key := "key-1"
//	message := "this is test message %v"
//	sendMaxTimes = 5
//
//	//get sub instance
//	tc := tinycells.GetTC()
//	kafka := tc.GetMQ().GetKafka()
//
//	//setup
//	address := []string{
//		"127.0.0.1:9092",
//	}
//	kafka.SetAddress(address)
//
//	//start
//	err := kafka.Start()
//	if err != nil {
//		log.Printf("start kafka failed, err:%v\n", err.Error())
//		return
//	}
//	wg.Add(1)
//	log.Printf("start kafka succeed\n")
//	log.Printf("topic:%v, key:%v, message:%v\n", topic, key, message)
//	log.Printf("kafka consumer register begin...\n")
//
//	//receive message inf son process
//	receiver := func(key, message []byte) bool {
//		log.Printf("received topic:%v, key:%v, message:%v \n\n", topic, string(key), string(message))
//		recvTimes++
//		return true
//	}
//
//	//register consumer
//	consumer := kafka.GetConsumer()
//	err = consumer.RegisterConsumer(topic, receiver)
//	if err != nil {
//		log.Printf("kafka consumer register failed, err:%v\n", err.Error())
//		return
//	}
//	log.Printf("kafka consumer register success.\n")
//
//	//loop sync sender
//	sf := func(wg *sync.WaitGroup) {
//		//send message
//		i := 0
//		for {
//			if i < sendMaxTimes {
//				producer := kafka.GetProducer()
//				msg := fmt.Sprintf(message, i)
//				err = producer.SendMessage(topic, key, msg)
//				if err != nil {
//					log.Printf("kafka producer send message failed, err:%v\n", err.Error())
//					break
//				}
//				log.Printf("kafka producer send message succeed\n")
//				time.Sleep(time.Second)
//				i++
//			}
//			if recvTimes >= sendMaxTimes {
//				wg.Done()
//				break
//			}
//		}
//	}
//	go sf(&wg)
//	log.Printf("kafka example running...\n")
//	wg.Wait()
//	log.Printf("kafka example done!\n")
//}

//mysql
func mysqlExample() {
	//get sub instance
	tc := tinycells.GetTC()
	mysql := tc.GetDB().GetMysql()

	//db tag
	dbTag := "sys"

	//gen config
	config := mysql.GenNewConfig()
	config.Host = "127.0.0.1"
	config.Port = 3306
	config.User = "root"
	config.Password = "123456"
	config.DBName = "sys"

	//create connect
	conn, err := mysql.CreateConnect(dbTag, config)
	if err != nil {
		log.Printf("connect mysql failed, err:%v\n", err)
		return
	}
	sf := func() {
		for {
			err = conn.Ping()
			log.Printf("connect mysql succeed, ping result %v\n", err)
			if err == nil {
				record, err := conn.GetRow("SELECT * FROM sys_config")
				log.Printf("record:%v, err:%v\n", record, err)
			}
			time.Sleep(time.Second)
		}
	}
	go sf()
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
	config.Addr = "127.0.0.1:6379"
	config.DBTag = dbName
	config.DBNum = 0

	//create connect
	_, err := rd.CreateConn(config)
	if err != nil {
		log.Printf("connect redis failed, err:%v\n", err)
		return
	}
	defer rd.C(dbName).Disconnect()

	//get client connect
	rc := rd.C(dbName).GetConnect()

	//ping
	result, err := rc.Ping().Result()
	if err != nil {
		log.Printf("ping redis failed, err:%v\n", err)
		return
	}
	log.Printf("pind redis result:%v\n", result)

	//get keys
	keys, err := rc.Keys("*").Result()
	log.Printf("keys:%v, err:%v\n", keys, err)
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
	conn := mgo.C(dbName)
	if conn == nil {
		log.Println("can't get conn")
		return
	}
	defer conn.Disconnect()
	count, err := conn.Count(dbCol, nil)
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
func loggerExample() {
	tc := tinycells.GetTC()
	logger := tc.GetLogger()
	config := logger.BuildDefaultConfig()
	err := logger.SetConfig(config)
	logger.SS().Infof("test logger")
	log.Printf("err:%v", err)
}
