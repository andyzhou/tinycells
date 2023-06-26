package main

import (
	"fmt"
	"github.com/andyzhou/tinycells"
	"github.com/andyzhou/tinycells/db/mysql"
	"github.com/andyzhou/tinycells/db/redis"
	"github.com/andyzhou/tinycells/media"
	"github.com/andyzhou/tinycells/util"
	"github.com/andyzhou/tinycells/web"
	genRedis "github.com/go-redis/redis/v7"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"sync"
	"time"
)

const (
	redisServer = "127.0.0.1:6381"
	pubSubChan = "ps_channel"
)

func main() {
	var (
		wg sync.WaitGroup
	)
	dfaExample()
	return

	timeExample()
	cmdExample()
	loggerExample()
	mongoExample()
	redisExample()
	mysqlExample()
	webAppExample()
	imageExample()

	//wait
	wg.Add(1)
	wg.Wait()
}

//dfa example
func dfaExample() {
	dfa := util.NewDFA()
	dfa.AddFilterWords("你妈逼的", "狗日")
	text := "我曹你妈逼的, 逼的你这个狗 日的，怎么这么傻啊。我也是服了，狗日的,这些话我都说不出口"
	found, newStr := dfa.ChangeSensitiveWords(text)
	fmt.Println("found:", found)
	fmt.Println(newStr)
}

//image example
func imageExample() {
	file := "test.png"
	img := media.NewImageResize(64)
	f, err := img.LoadFile(file)
	if err != nil {
		log.Println("err:", err)
		return
	}
	defer f.Close()
	//img.ResizeFromFile(file)
	imgByte, err := img.ResizeFromIOReader(f)
	log.Println("imgByte:", len(imgByte), ", err:", err)

	newFile := fmt.Sprintf("%v.png", time.Now().Unix())
	f1, err := os.Create(newFile)
	if f1 != nil {
		f1.Write(imgByte)
		f1.Close()
	}
}

//time example
func timeExample() {
	now := time.Now().Unix()
	u := util.Util{}
	dayStr := u.TimeStampToDayStr(now)
	log.Println("dayStr:", dayStr)
	diffTimeStr := u.DiffTimeStampToStr(1677334000)
	log.Println("diffTimeStr:", diffTimeStr)
}

//web app
func webAppExample()  {
	app := web.NewApp()
	app.SetTplPattern("./tpl/*.html")
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

//global variable
var (
	pubSubWg sync.WaitGroup
)

//mysql
func mysqlExample() {
	//get sub instance
	tc := tinycells.GetTC()
	db := tc.GetDB().GetMysql()

	//db tag
	dbTag := "sys"

	//gen config
	config := db.GenNewConfig()
	config.Host = "127.0.0.1"
	config.Port = 3306
	config.User = "root"
	config.Password = "123456"
	config.DBName = "adam"


	for {
		//create connect
		conn, err := db.CreateConnect(dbTag, config)
		if err != nil {
			log.Printf("connect mysql failed, err:%v\n", err)
			return
		}
		time.Sleep(time.Second)
		err = conn.Ping()
		log.Printf("connect mysql succeed, ping result %v\n", err)
		if err != nil {
			return
		}
		break
	}

	//query
	conn := db.GetConnect(dbTag)
	record, err := conn.GetRow("SELECT * FROM sys_config")
	log.Printf("record:%v, err:%v\n", record, err)
	return

	//update
	updateMap := map[string]interface{}{
		"reviews":1,
	}
	where := map[string]mysql.WherePara{
		"subjectId":{
			Val: "31",
		},
	}
	tab := "sparrow_subject"
	tabField := "count"
	err = db.UpdateCountOfDataAdv(
				updateMap,
				where,
				tabField,
				tab,
				conn,
			)
	log.Println(err)
}

//redis
func cbForPubSub(msg *genRedis.Message) error {
	channel := msg.Channel
	info := msg.Payload
	fmt.Println("channel:", channel, ", info:", info)
	pubSubWg.Done()
	return nil
}

func redisExample() {
	//get sub instance
	tc := tinycells.GetTC()
	rd := tc.GetDB().GetRedis()

	//set key data
	dbName := "base"

	//gen config
	config := rd.GenNewConfig()
	config.Addr = redisServer
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
	client := rd.C(dbName)
	rc := client.GetConnect()

	//ping
	result, err := rc.Ping().Result()
	if err != nil {
		log.Printf("ping redis failed, err:%v\n", err)
		return
	}
	log.Printf("pind redis result:%v\n", result)

	//test pub sub
	ps := redis.NewPubSub()
	ps.SetConn(client)
	ps.Subscript(pubSubChan, cbForPubSub)

	sf := func() {
		ps.Publish(pubSubChan, "test")
	}
	time.AfterFunc(time.Second * 2, sf)

	//init wg
	pubSubWg.Add(1)
	pubSubWg.Wait()

	//get keys
	//keys, err := rc.Keys("*").Result()
	//log.Printf("keys:%v, err:%v\n", keys, err)

	log.Println("redis test done!")
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
