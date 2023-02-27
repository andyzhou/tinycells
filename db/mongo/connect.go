package mongo

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"reflect"
	"time"
)

//face info
type Connection struct {
	client *mongo.Client
	db     *mongo.Database
	config *Config
	optionSlice []*options.ClientOptions
}

//construct
func NewConnection(cfg *Config) *Connection {
	if cfg.PoolSize <= 0 {
		cfg.PoolSize = DefaultPoolSize
	}
	this := &Connection{
		config: cfg,
		optionSlice: []*options.ClientOptions{},
	}
	this.interInit()
	return this
}

////////////////
//bulk opt api
////////////////

//write begin
func (f *Connection) BulkWriteBegin(ordered bool) *BulkWriteOp {
	return &BulkWriteOp{
		WriteModel: []mongo.WriteModel{},
		BulkWriteOpt: &options.BulkWriteOptions{
			Ordered: &ordered,
		},
	}
}

func (f *Connection) BulkWriteOpInsertOne(bwOp *BulkWriteOp, doc interface{}) {
	op := mongo.NewInsertOneModel().SetDocument(doc)
	bwOp.WriteModel = append(bwOp.WriteModel, op)
}

func (f *Connection) BulkWriteOpUpdateOne(bwOp *BulkWriteOp,
	filter interface{}, update interface{}, upsert bool) {
	op := mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(D{{Key: "$set", Value: update}}).SetUpsert(upsert)
	bwOp.WriteModel = append(bwOp.WriteModel, op)
}

//write end
func (f *Connection) BulkWriteEnd(col string, bwOp *BulkWriteOp) (*BulkWriteResult, error) {
	ctx, cancel := f.createContext()
	defer cancel()
	return f.db.Collection(col).BulkWrite(ctx, bwOp.WriteModel, bwOp.BulkWriteOpt)
}

////////////////
//base opt api
////////////////

//insert many
func (f *Connection) InsertMany(
				col string,
				docs []interface{},
				opts ...*InsertManyOptions,
			) error {
	ctx, cancel := f.createContext()
	defer cancel()
	_, err := f.db.Collection(col).InsertMany(ctx, docs, opts...)
	return err
}

//insert one
func (f *Connection) InsertOne(
				col string,
				doc interface{},
				opts ...*InsertOneOptions,
			) error {
	ctx, cancel := f.createContext()
	defer cancel()
	_, err := f.db.Collection(col).InsertOne(ctx, doc, opts...)
	return err
}

//delete one
func (f *Connection) DeleteOne(col string, filter interface{},
	opts ...*DeleteOptions) error {
	ctx, cancel := f.createContext()
	defer cancel()
	_, err := f.db.Collection(col).DeleteOne(ctx, filter, opts...)
	return err
}

//delete many
func (f *Connection) DelMany(
				col string,
				filter interface{},
				opts ...*DeleteOptions,
			) error {
	ctx, cancel := f.createContext()
	defer cancel()
	_, err := f.db.Collection(col).DeleteMany(ctx, filter, opts...)
	return err
}

//update batch
func (f *Connection) UpdateMany(
				col string,
				filter interface{},
				update interface{},
				opts ...*UpdateOptions,
			) error {
	ctx, cancel := f.createContext()
	defer cancel()
	_, err := f.db.Collection(col).UpdateMany(ctx, filter, update, opts...)
	return err
}

//update one
func (f *Connection) UpdateOne(
				col string,
				filter interface{},
				update interface{},
				opts ...*UpdateOptions,
			) error {
	ctx, cancel := f.createContext()
	defer cancel()
	_, err := f.db.Collection(col).UpdateOne(ctx, filter, update, opts...)
	return err
}

//find and update one
func (f *Connection) FindOneAndUpdate(
				col string,
				filter interface{},
				update interface{},
				resp interface{},
				opts ...*FindOneAndUpdateOptions,
			) error {
	ctx, cancel := f.createContext()
	defer cancel()
	return f.db.Collection(col).FindOneAndUpdate(ctx, filter, update, opts...).Decode(resp)
}

//find one
func (f *Connection) FindOne(
				col string,
				filter interface{},
				resp interface{},
				opts ... *FindOneOptions,
			) error {
	ctx, cancel := f.createContext()
	defer cancel()
	return f.db.Collection(col).FindOne(ctx, filter, opts...).Decode(resp)
}

//find batch doc by cond
func (f *Connection) Find(
				col string,
				filter interface{},
				opts ...*FindOptions,
			) (*mongo.Cursor, error) {
	ctx, cancel := f.createContext()
	defer cancel()
	return f.db.Collection(col).Find(ctx, filter, opts...)
}

//count
func (f *Connection) Count(col string, filter interface{}) (int64, error) {
	if filter == nil {
		filter = D{}
	}
	ctx, cancel := f.createContext()
	defer cancel()
	count, err := f.db.Collection(col).CountDocuments(ctx, filter)
	return count, err
}

//drop collection
func (f *Connection) DropCollection(col string) error {
	//check
	if col == "" {
		return errors.New("invalid parameter")
	}
	ctx, cancel := f.createContext()
	defer cancel()
	collection := f.db.Collection(col)
	if collection == nil {
		return errors.New("can't get collection")
	}
	err := collection.Drop(ctx)
	return err
}

//get collections
func (f *Connection) GetCollections(col string) ([]string, error) {
	//check
	if col == "" {
		return nil, errors.New("invalid parameter")
	}
	ctx, cancel := f.createContext()
	defer cancel()
	filter := bson.D{}
	collections, err := f.db.ListCollectionNames(ctx, filter)
	return collections, err
}

//drop index
func (f *Connection) DropIndex(col string, indexNames ...string) error {
	var (
		err error
	)
	//check
	if col == "" {
		return errors.New("invalid parameter")
	}
	ctx, cancel := f.createContext()
	defer cancel()
	if indexNames != nil && len(indexNames) > 0 {
		for _, indexName := range indexNames {
			f.db.Collection(col).Indexes().DropOne(ctx, indexName)
		}
	}else{
		_, err = f.db.Collection(col).Indexes().DropAll(ctx)
	}
	return err
}

//create index
func (f *Connection) CreateIndex(col string, keys, opts interface{}) error {
	//check
	if col == "" || keys == nil {
		return errors.New("invalid parameter")
	}
	//begin create index
	ctx, cancel := f.createContext()
	defer cancel()
	indexModel := mongo.IndexModel{
		Keys:    keys,
		Options: opts.(*options.IndexOptions),
	}
	_, err := f.db.Collection(col).Indexes().CreateOne(ctx, indexModel)
	return err
}

//get all indexes
func (f *Connection) GetAllIndex(col string) (map[string]interface{}, error) {
	//check
	if col == "" {
		return nil, errors.New("invalid parameter")
	}
	//list all index of col
	ctx, cancel := f.createContext()
	defer cancel()
	cur, err := f.db.Collection(col).Indexes().List(ctx)
	if err != nil {
		return nil, err
	}
	//format result
	allKeyMap := make(map[string]interface{})
	for cur.Next(context.TODO()) {
		//try decode doc obj
		indexVal := &MGIndex{}
		cur.Decode(indexVal)
		if indexVal == nil || indexVal.Name == "" {
			continue
		}
		allKeyMap[indexVal.Name] = indexVal.Key
	}
	return allKeyMap, nil
}

//ping server
func (f *Connection) Ping() error {
	//check
	if f.client == nil {
		return errors.New("client not init")
	}
	ctx, cancel := f.createContext()
	defer cancel()
	err := f.client.Ping(ctx, readpref.Primary())
	return err
}

//disconnect
func (f *Connection) Disconnect() error {
	//check
	if f.client == nil {
		return errors.New("client not init")
	}
	ctx, cancel := f.createContext()
	defer cancel()
	err := f.client.Disconnect(ctx)
	return err
}

//connect
func (f *Connection) Connect() error {
	//init client
	client, err := mongo.NewClient(f.optionSlice...)
	if err != nil {
		return err
	}

	//init db
	db := client.Database(f.config.DBName)

	//try connect server
	ctx, cancel := f.createContext()
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		return err
	}

	//sync inter variables
	f.db = db
	f.client = client
	return nil
}

////////////////
//private func
////////////////

//create context
func (f *Connection) createContext() (context.Context, context.CancelFunc){
	return context.WithTimeout(context.Background(), ServerOptTimeOut*time.Second)
}

//reset dynamic json object
func (f *Connection) clearObj(v interface{}) {
	p := reflect.ValueOf(v).Elem()
	p.Set(reflect.Zero(p.Type()))
}

//inter init
func (f *Connection) interInit() {
	//init options
	optionSlice := make([]*options.ClientOptions, 0)
	optionsMain := options.Client().ApplyURI(f.config.DBUrl)
	optionsMain.SetMinPoolSize(uint64(DefaultPoolSize))
	if f.config.PoolSize > 0 {
		optionsMain.SetMaxPoolSize(uint64(f.config.PoolSize))
	}
	optionSlice = append(optionSlice, optionsMain)
	if f.config.UserName != "" {
		auth := options.Client().SetAuth(options.Credential{
			AuthSource:"admin",
			Username:f.config.UserName,
			Password:f.config.Password,
			PasswordSet:true,
		})
		optionSlice = append(optionSlice, auth)
	}
	f.optionSlice = optionSlice
}