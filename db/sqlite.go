package db


import (
	_ "github.com/mattn/go-sqlite3"
	"database/sql"
	"log"
	"fmt"
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
func NewSqlLite(dbFile string) *SqlLite {
	this := &SqlLite{
		dbFile:dbFile,
	}

	//init db
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		log.Println("SqlLite, open db failed, err:", err.Error())
	}else{
		//sync db
		this.db = db
	}

	return this
}

///////
//api
//////

//close db
func (s *SqlLite) Close() {
	if s.db != nil {
		s.db.Close()
		s.db = nil
	}
}

//execute
func (s *SqlLite) Execute(sql string) (bool, int64, int64) {
	var (
		lastInsertId, effectRows int64
		err error
	)

	if sql == "" || s.db == nil {
		return false, lastInsertId, effectRows
	}

	result, err := s.db.Exec(sql)
	if err != nil {
		log.Println("SqlLite, exec sql:", sql, " failed, err:", err.Error())
		return false, lastInsertId, effectRows
	}

	lastInsertId, err = result.LastInsertId()
	effectRows, err = result.RowsAffected()

	if err != nil {
		log.Println("SqlLite, exec convert failed, err:", err.Error())
		return false, lastInsertId, effectRows
	}

	return true, lastInsertId, effectRows
}

//query
func (s *SqlLite) Query(sql string) (bool, []map[string]string) {
	var (
		colSize, i int
		err error
		tempStr string
		tempSlice = make([]interface{}, 0)

		results = make([]map[string]string, 0)
	)
	if sql == "" || s.db == nil {
		return false, nil
	}

	rows, err := s.db.Query(sql)
	if err != nil {
		log.Println("SqlLite, query sql:", sql, " failed, err:", err.Error())
		return false, nil
	}

	defer rows.Close()
	cols, err := rows.Columns()
	if err != nil {
		log.Println("SqlLite, can't get column info")
		return false, nil
	}

	//init temp slice
	colSize = len(cols)
	for i = 0; i < colSize; i++ {
		tempSlice = append(tempSlice, new([]byte))
	}

	for rows.Next() {
		//process single row record
		err = rows.Scan(tempSlice...)
		//log.Println("err:", err, ", tempSlice:", tempSlice)
		i = 0
		tempMap := make(map[string]string)
		for _, col := range cols {
			//log.Println("col:", col, ", type:", reflect.TypeOf(tempSlice[i]))
			tempStr = ""
			switch v := tempSlice[i].(type) {
			case *[]uint8:
				//log.Println("col:", col, ", v:", string(*v))
				tempStr = fmt.Sprintf("%s", string(*v))
			}
			tempMap[col] = tempStr
			i++
		}
		results = append(results, tempMap)
	}

	//clear temp variable
	tempSlice = tempSlice[:0]

	//log.Println("results:", results)
	return true, results
}

