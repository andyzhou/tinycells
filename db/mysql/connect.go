package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/andyzhou/tinycells/util"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"sync"
	"time"
)

//face info
type Connect struct {
	dbConf *Config
	poolMap map[int]*sql.DB
	poolSize int
	address string
	checkChan chan struct{}
	closeChan chan struct{}
	util.Util
	sync.RWMutex
}

//construct
func NewConnect(conf *Config) *Connect {
	this := &Connect{
		dbConf: conf,
		poolMap: map[int]*sql.DB{},
		checkChan: make(chan struct{}, 1),
		closeChan: make(chan struct{}, 1),
	}
	go this.poolChecker()
	return this
}

//quit
func (f *Connect) Quit() {
	f.closeChan <- struct{}{}
}

//get db instance
func (f *Connect) GetDB() *sql.DB {
	return f.getRandomDB()
}

//transaction
func (f *Connect) Transaction(query string, args ...interface{}) (int64, int64, error) {
	//get random db
	db := f.getRandomDB()
	if db == nil {
		return 0, 0, errors.New("can't get db instance")
	}
	//begin transaction
	tx, err := db.Begin()
	if err != nil {
		return 0, 0, err
	}
	//execute
	result, err := tx.ExecContext(context.Background(), query, args...)
	if err != nil {
		return 0, 0, err
	}
	//commit
	err = tx.Commit()
	if err != nil {
		//rollback
		tx.Rollback()
		return 0, 0, err
	}
	lastInsertId, _ := result.LastInsertId()
	effectRows, _ := result.RowsAffected()
	return lastInsertId, effectRows, nil
}

//execute sql
//return lastInsertId, effectRows, error
func (f *Connect) Execute(query string, args ...interface{}) (int64, int64, error) {
	//get random db
	db := f.getRandomDB()
	if db == nil {
		return 0, 0, errors.New("can't get db instance")
	}
	//exec sql
	result, err := db.Exec(query, args...)
	if err != nil {
		return 0, 0, err
	}
	lastInsertId, _ := result.LastInsertId()
	effectRows, _ := result.RowsAffected()
	return lastInsertId, effectRows, nil
}

//get one row record
func (f *Connect) GetRow(query string, args ...interface{}) (map[string]interface{}, error) {
	recordMap := make(map[string]interface{})
	queryNew := fmt.Sprintf("%s LIMIT 1", query)
	records, err := f.GetRows(queryNew, args...)
	if err != nil {
		return nil, err
	}
	//return first record of slice
	for _, record := range records {
		if len(record) <= 0 {
			continue
		}
		recordMap = record
		break
	}
	return recordMap, nil
}

//get batch row records
func (f *Connect) GetRows(query string, args ...interface{}) ([]map[string]interface{}, error) {
	//get random db
	db := f.getRandomDB()
	if db == nil {
		return nil, errors.New("can't get db instance")
	}
	//format result
	records := make([]map[string]interface{}, 0)
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	//init map for return
	columns, _ := rows.Columns()

	scanArgs := make([]interface{}, len(columns))
	values := make([]interface{}, len(columns))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	//begin get record and copy into slice for return
	k := 0
	for rows.Next() {
		//storage row data into record map
		err = rows.Scan(scanArgs...)
		record := make(map[string]interface{})
		for i, col := range values {
			if col != nil {
				record[columns[i]] = col
			}
		}
		//append record into slice list
		records = append(records, record)
		k++
	}
	return records, nil
}

//ping server
func (f *Connect) Ping() error {
	//get random db
	db := f.getRandomDB()
	if db == nil {
		return errors.New("can't get db instance")
	}
	return db.Ping()
}

////////////////
//private func
////////////////

//get rand db from pool
func (f *Connect) getRandomDB() *sql.DB {
	if f.poolSize <= 0 {
		return nil
	}
	randIdx := f.GetRandomVal(f.poolSize)
	f.Lock()
	defer f.Unlock()
	v, ok := f.poolMap[randIdx]
	if ok && v != nil {
		return v
	}
	return nil
}

//pool init and check process
func (f *Connect) poolChecker() {
	defer func() {
		if err := recover(); err != nil {
			log.Println("mysql.Connect:poolChecker panic, err:", err)
		}
		//release pool
		f.releasePool()
	}()

	//check
	if f.dbConf.PoolSize <= 0 {
		f.dbConf.PoolSize = DBPoolMin
	}

	//format address
	f.address = f.getDBAddress()

	//fill pool map
	f.fillPoolMap()

	//start first checker
	f.checkChan <- struct{}{}

	//loop
	for {
		select {
		case <- f.checkChan:
			{
				//connect check
				f.checkOrConnect()
				//next ticker
				time.Sleep(time.Second * ConnCheckRate)
				f.checkChan <- struct{}{}
			}
		case <- f.closeChan:
			return
		}
	}
}

//check or connect server
func (f *Connect) checkOrConnect() {
	var (
		err error
	)
	if f.poolMap == nil {
		return
	}
	f.Lock()
	defer f.Unlock()
	for k, v := range f.poolMap {
		if v == nil {
			//try connect
			newConn, _ := f.connectServer()
			if newConn != nil {
				f.poolMap[k] = newConn
			}
			continue
		}
		err = v.Ping()
		if err == nil {
			continue
		}
		//ping failed, try connect
		v.Close()
		newConn, _ := f.connectServer()
		if newConn != nil {
			f.poolMap[k] = newConn
		}
	}
}

//release pool
func (f *Connect) releasePool() {
	if f.poolMap == nil {
		return
	}
	f.Lock()
	defer f.Unlock()
	for _, v := range f.poolMap {
		if v != nil {
			v.Close()
		}
	}
	f.poolMap = map[int]*sql.DB{}
}

//fill pool map
func (f *Connect) fillPoolMap() {
	for i := 0; i < f.dbConf.PoolSize; i++ {
		f.poolMap[i] = nil
	}
	f.poolSize = len(f.poolMap)
}

//try connect
func (f *Connect) connectServer() (*sql.DB, error) {
	//check
	if f.dbConf == nil {
		return nil, errors.New("db conf is nil")
	}
	//init db driver
	db, err := sql.Open("mysql", f.address)
	if err != nil {
		return nil, err
	}
	//try ping db server
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return db, nil
}

//format db address
func (f *Connect) getDBAddress() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		f.dbConf.User, f.dbConf.Password,
		f.dbConf.Host, f.dbConf.Port,
		f.dbConf.DBName,
	)
}
