package sqlite


import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

/*
 * sql lite db interface
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

//sql lite info
type SqlLite struct {
	dbFile string
	db *sql.DB
}

//construct
func NewSqlLite() *SqlLite {
	this := &SqlLite{
	}
	return this
}

//open db file
func (s *SqlLite) OpenDBFile(dbFile string) error {
	//check
	if dbFile == "" {
		return errors.New("invalid parameter")
	}

	//init db
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		log.Println("SqlLite, open db failed, err:", err.Error())
		return err
	}

	//sync db
	s.db = db
	return nil
}

//close db
func (s *SqlLite) Close() {
	if s.db != nil {
		s.db.Close()
		s.db = nil
	}
}

//execute
func (s *SqlLite) Execute(sql string, args []interface{}) (int64, int64, error) {
	var (
		lastInsertId, effectRows int64
		err error
	)
	//check
	if sql == "" || s.db == nil {
		return lastInsertId, effectRows, errors.New("invalid parameter")
	}
	//exec sql
	result, err := s.db.Exec(sql, args...)
	if err != nil {
		return lastInsertId, effectRows, err
	}
	lastInsertId, err = result.LastInsertId()
	effectRows, err = result.RowsAffected()
	if err != nil {
		return lastInsertId, effectRows, err
	}

	return lastInsertId, effectRows, nil
}

//query
func (s *SqlLite) Query(sql string, args []interface{}) ([]map[string]string, error) {
	var (
		colSize, i int
		err error
		tempStr string
		tempSlice = make([]interface{}, 0)

		results = make([]map[string]string, 0)
	)
	if sql == "" || s.db == nil {
		return nil, errors.New("invalid parameter")
	}
	rows, err := s.db.Query(sql, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	//init temp slice
	colSize = len(cols)
	for i = 0; i < colSize; i++ {
		tempSlice = append(tempSlice, new([]byte))
	}

	for rows.Next() {
		//process single row record
		err = rows.Scan(tempSlice...)
		i = 0
		tempMap := make(map[string]string)
		for _, col := range cols {
			tempStr = ""
			switch v := tempSlice[i].(type) {
			case *[]uint8:
				tempStr = fmt.Sprintf("%s", string(*v))
			default:
				tempStr = fmt.Sprintf("%v", v)
			}
			tempMap[col] = tempStr
			i++
		}
		results = append(results, tempMap)
	}

	//clear temp variable
	tempSlice = tempSlice[:0]
	return results, nil
}
