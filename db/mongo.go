package db

import (
	"context"
	"errors"
	"fmt"
	"github.com/andyzhou/tinycells/tc"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"log"
	"reflect"
	"sync"
	"time"
)

/*
 * Mongo db interface
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 * - base on `go.mongodb.org/mongo-driver/mongo`
 * - collection name like table name of mysql
 */

 //inter macro define
 const (
 	ConnServerTimeOut = 20 //xx seconds
 	PingServerTimeOut = 20 //xx seconds
 	GeneralOptTimeOut = 30 //xx seconds
 )

 //mongo account info
 type MongoAccount struct {
 	UserName string
 	Password string
 	DBUrl string
 	DBName string
 }

 //mongo db info
 type MongoDB struct {
 	account *MongoAccount
 	db *mongo.Database
 	client *mongo.Client
 	//genCtx context.Context
 	collections map[string]*mongo.Collection
 	tc.BaseJson
 	sync.Mutex
 }

 //construct
 //url like: `localhost:27017`
func NewMongoDB(account *MongoAccount) *MongoDB {
	//init db url
	//db url like: 'mongodb://localhost:27017'
	account.DBUrl = fmt.Sprintf("mongodb://%s", account.DBUrl)

	//self init
	this := &MongoDB{
		account:account,
		collections:make(map[string]*mongo.Collection),
	}

	//inter init
	this.interInit()

	return this
}

/////////
//api
////////

//get batch docs
//filter used as query condition
//jsonObj is a pointer of json type, used for decode data
//optionSlice used for limit, sort, etc.
//return json bytes slice and error
func (d *MongoDB) GetBatchDocs(collectName string, filter, jsonObj interface{},
				optionSlice []*options.FindOptions) ([][]byte, error) {
	var (
		//tempByteData = make([]byte, 0)
		result = make([][]byte, 0)
		err error
	)

	//basic check
	if collectName == "" || filter == nil {
		return nil, errors.New("collect name or filter is empty")
	}

	//check and init collection
	collection, genCtx, err := d.checkAndInitCollector(collectName)
	if err != nil {
		return nil, err
	}
	defer  genCtx.Done()

	//get batch doc
	//cur, err := collection.Find(d.genCtx, bson.D{})
	cur, err := collection.Find(genCtx, filter, optionSlice...)
	if err != nil {
		return nil, err
	}

	//loop records
	defer cur.Close(genCtx)
	for cur.Next(genCtx) {
		//clear json object
		//this is very important!!!
		d.clearObj(jsonObj)

		//try decode json object
		err = cur.Decode(jsonObj)
		if err != nil {
			log.Println("MongoDB::GetBatchDocs, decode json failed, err:", err.Error())
			continue
		}
		//add json object into slice
		byteData := make([]byte, 0)
		byteData = d.Encode(jsonObj)
		result = append(result, byteData)
		byteData = byteData[:0]
	}

	if err := cur.Err(); err != nil {
		log.Println("MongoDB::GetBatchDocs, cursor err:", err.Error())
		return nil, err
	}

	return result, nil
}

//get single doc
func (d *MongoDB) GetOneDoc(collectName string, filter, jsonObj interface{}) ([]byte, error) {
	//basic check
	if collectName == "" || filter == nil {
		return nil, errors.New("collect name or filter is empty")
	}

	//check and init collection
	collection, genCtx, err := d.checkAndInitCollector(collectName)
	if err != nil {
		return nil, err
	}
	defer  genCtx.Done()

	//get batch doc
	//cur, err := collection.Find(d.genCtx, bson.D{})
	resp := collection.FindOne(genCtx, filter)
	if resp == nil || resp.Err() != nil {
		return nil, resp.Err()
	}

	//try decode data
	err = resp.Decode(jsonObj)
	if err != nil {
		return nil, err
	}

	//encode json object into byte
	byteData := make([]byte, 0)
	byteData = d.Encode(jsonObj)
	return byteData, nil
}

//delete batch docs
func (d *MongoDB) DelBatchDocs(collectName string, filter interface{}) error {
	//basic check
	if collectName == "" || filter == nil {
		return errors.New("collect name or filter is empty")
	}

	//check and init collection
	collection, genCtx, err := d.checkAndInitCollector(collectName)
	if err != nil {
		return err
	}
	defer  genCtx.Done()

	//try delete batch doc
	_, err = collection.DeleteMany(genCtx, filter)

	return err
}

//delete doc
func (d *MongoDB) DelOneDoc(collectName string, filter interface{}) error {
	//basic check
	if collectName == "" || filter == nil {
		return errors.New("collect name or filter is empty")
	}

	//check and init collection
	collection, genCtx, err := d.checkAndInitCollector(collectName)
	if err != nil {
		return err
	}

	//try delete one doc
	defer  genCtx.Done()
	_, err = collection.DeleteOne(genCtx, filter)

	return err
}


//update batch doc
func (d *MongoDB) UpdateBathDocs(collectName string, filter, update interface{}) error {
	//check and init collection
	collection, genCtx, err := d.checkAndInitCollector(collectName)
	if err != nil {
		return err
	}

	//update batch
	defer  genCtx.Done()
	_, err = collection.UpdateMany(genCtx, filter, update)

	return err
}

//update doc 
func (d *MongoDB) UpdateDoc(collectName string, filter, data interface{}) error {
	//check and init collection
	collection, genCtx, err := d.checkAndInitCollector(collectName)
	if err != nil {
		return err
	}
	defer  genCtx.Done()

	//update one
	_, err = collection.UpdateOne(genCtx, filter, data)
	//log.Println("resp:", resp)

	return err
}


//add batch docs
func (d *MongoDB) AddBatchDocs(collectName string, jsonSlice []interface{}) error {
	if collectName == "" || jsonSlice == nil {
		return errors.New("lost parameters")
	}

	//check and init collection
	collection, genCtx, err := d.checkAndInitCollector(collectName)
	if err != nil {
		return err
	}

	//add new batch doc
	defer genCtx.Done()

	_, err = collection.InsertMany(genCtx, jsonSlice)
	return err
}

//add new doc
//return new doc id and error
func (d *MongoDB) AddDoc(collectName string, json interface{}) (string, error) {
	var (
		docId string
		err error
	)

	//basic check
	if collectName == ""{
		return docId, errors.New("collect name is empty")
	}

	//init collection
	collection := d.getOrInitCollection(collectName)
	if collection == nil {
		return docId, errors.New("init collection failed")
	}

	//init general ctx
	genCtx, _ := context.WithTimeout(context.TODO(), GeneralOptTimeOut * time.Second)

	//add new record
	_, err = collection.InsertOne(genCtx, json)
	if err != nil {
		return docId, err
	}

	//try get doc id as string format
	//if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
	//	docId = oid.Hex()
	//}

	return docId, err
}


//ping server
func (d *MongoDB) Ping() error {
	ctx, _ := context.WithTimeout(context.Background(), PingServerTimeOut * time.Second)
	err := d.client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Println("MongoDB::Ping failed, err:", err.Error())
		return err
	}
	return nil
}

/////////////////
//private func
////////////////

//check and init collector
func (d *MongoDB) checkAndInitCollector(collectName string) (*mongo.Collection, context.Context, error) {
	//basic check
	if collectName == "" {
		return nil, nil, errors.New("collect name or filter is empty")
	}

	//init collection
	collection := d.getOrInitCollection(collectName)
	if collection == nil {
		return nil, nil, errors.New("init collection failed")
	}

	//init general ctx
	genCtx, _ := context.WithTimeout(context.TODO(), GeneralOptTimeOut * time.Second)
	return collection, genCtx, nil
}


//reset dynamic json object
func (d *MongoDB)clearObj(v interface{}) {
	p := reflect.ValueOf(v).Elem()
	p.Set(reflect.Zero(p.Type()))
}


//get or init collection
func (d *MongoDB) getOrInitCollection(collectionName string) *mongo.Collection {
	if collectionName == "" {
		return nil
	}

	d.Lock()
	defer d.Unlock()
	v, ok := d.collections[collectionName]
	if ok {
		return v
	}

	//init it
	collection := d.db.Collection(collectionName)

	//sync into map
	d.collections[collectionName] = collection

	return collection
}

//inter init
func (d *MongoDB) interInit() {
	//init options
	optionsMain := options.Client().ApplyURI(d.account.DBUrl)
	auth := options.Client().SetAuth(options.Credential{
					AuthSource:"admin",
					Username:d.account.UserName,
					Password:d.account.Password,
					PasswordSet:true,
			})

	//options.SetAuth(Credential{
	//		AuthSource: "admin", Username: "foo",
	//		Password: "bar", PasswordSet: true,
	//	})



	//author := options.Client().SetAuth(Credential{
	//	AuthSource: "admin", Username: "foo",
	//	Password: "bar", PasswordSet: true,
	//}),

	//init client
	client, err := mongo.NewClient(optionsMain, auth)
	if err != nil {
		log.Println("MongoDB::interIinit, init failed, err:", err.Error())
		panic(err)
	}

	//try connect server
	ctx, _ := context.WithTimeout(context.TODO(), ConnServerTimeOut * time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Println("MongoDB::interIinit, connect failed, err:", err.Error())
		panic(err)
	}

	//init general ctx
	//genCtx, _ := context.WithTimeout(context.TODO(), GeneralOptTimeOut * time.Second)

	//init db
	db := client.Database(d.account.DBName)

	//sync client and db
	d.Lock()
	defer d.Unlock()
	d.db = db
	d.client = client
	//d.genCtx = genCtx
}
