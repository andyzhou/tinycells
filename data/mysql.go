package data

import (
	"github.com/andyzhou/tinycells/tc"
	"runtime/debug"
	"strconv"
	"bytes"
	"log"
	"fmt"
)

/*
 * base mysql db data face
 * - only for json format data
 * - use as anonymous class
 */

 //const db field
 const (
 	TableFieldOfTotal = "total"
 	TableFieldOfData = "data" //db data field
 )

 //where kind
 const (
 	WhereKindOfGen = iota
 	WhereKindOfIn	  //for in('x','y')
 	WhereKindOfInSet  //for FIND_IN_SET(val, `x`, 'y')
 )

 //where para
 type WherePara struct {
 	Kind int
 	Val interface{}
 }

 //face info
 type BaseMysql struct {
 	tc.Utils
 }

 //get total num
func (d *BaseMysql) GetTotalNum(
					whereMap map[string]WherePara,
					table string,
					db *tc.DBService,
				) int {
	var (
		values = make([]interface{}, 0)
		total int
	)

	//basic check
	if table == "" || db == nil {
		return total
	}

	//format where sql
	whereBuffer, whereValues := d.formatWhereSql(whereMap)
	if whereValues != nil {
		values = append(values, whereValues...)
	}

	//format sql
	sql := fmt.Sprintf("SELECT count(*) as total FROM %s %s",
						table,
						whereBuffer.String(),
					)

	//query one record
	recordMap, err := db.GetRow(sql, values...)
	if err != nil {
		log.Println("BaseMysql::GetTotalNum failed, err:", err.Error())
		log.Println("track:", string(debug.Stack()))
		return total
	}

	total = d.getTotalVal(recordMap)
	return total
}

 //get batch data with condition
func (d *BaseMysql) GetBatchData(
				whereMap map[string]WherePara,
				orderBy string,
				offset int,
				size int,
				table string,
				db *tc.DBService,
			) [][]byte {
	var (
		limitSql, orderBySql string
		values = make([]interface{}, 0)
	)

	//basic check
	if table == "" || db == nil {
		return nil
	}

	//format where sql
	whereBuffer, whereValues := d.formatWhereSql(whereMap)
	if whereValues != nil {
		values = append(values, whereValues...)
	}

	//format limit sql
	if size > 0 {
		limitSql = fmt.Sprintf("LIMIT %d, %d", offset, size)
	}

	//format order by sql
	if orderBy != "" {
		orderBySql = fmt.Sprintf(" ORDER BY %s", orderBy)
	}

	//format sql
	sql := fmt.Sprintf("SELECT data FROM %s %s %s %s",
						table,
						whereBuffer.String(),
						orderBySql,
						limitSql,
					)

	//query records
	recordsMap, err := db.GetArray(sql, values...)
	if err != nil {
		log.Println("BaseMysql::GetBathData failed, err:", err.Error())
		log.Println("track:", string(debug.Stack()))
		return nil
	}

	//check records map
	if recordsMap == nil || len(recordsMap) <= 0 {
		return nil
	}

	//init result
	result := make([][]byte, 0)

	//analyze original record
	for _, recordMap := range recordsMap {
		jsonByte := d.GetByteData(recordMap)
		if jsonByte == nil {
			continue
		}
		result = append(result, jsonByte)
	}
	return result
}


 //get one data
 //dataField default value is 'data'
func (d *BaseMysql) GetOneData(
				dataField string,
				whereMap map[string]WherePara,
				table string,
				db *tc.DBService,
			) []byte {
	var (
		assignedDataField string
		values = make([]interface{}, 0)
	)

	//basic check
	if whereMap == nil || table == "" || db == nil {
		return nil
	}

	//format where sql
	whereBuffer, whereValues := d.formatWhereSql(whereMap)
	if whereValues != nil {
		values = append(values, whereValues...)
	}

	if dataField != "" {
		assignedDataField = dataField
	}else{
		assignedDataField = "data"
	}

	//format sql
	sql := fmt.Sprintf("SELECT %s FROM %s %s",
						assignedDataField,
						table,
						whereBuffer.String(),
					)

	//query records
	recordMap, err := db.GetRow(sql, values...)
	if err != nil {
		log.Println("BaseMysql::GetOneData failed, err:", err.Error())
		log.Println("track:", string(debug.Stack()))
		return nil
	}

	//check record map
	if recordMap == nil || len(recordMap) <= 0 {
		return nil
	}

	//get json byte data
	jsonByte := d.GetByteDataByField(assignedDataField, recordMap)
	return jsonByte
}


//delete data
func (d *BaseMysql) DelOneData(
				whereMap map[string]WherePara,
				table string,
				db *tc.DBService,
			) bool {
	var (
		values = make([]interface{}, 0)
	)

	//basic check
	if whereMap == nil || table == "" || db == nil {
		return false
	}

	//format where sql
	whereBuffer, whereValues := d.formatWhereSql(whereMap)
	if whereValues != nil {
		values = append(values, whereValues...)
	}

	//format sql
	sql := fmt.Sprintf("DELETE FROM %s %s", table, whereBuffer.String())

	//remove from db
	_, _, err := db.Execute(sql, values...)
	if err != nil {
		log.Println("BaseMysql::DelOneData failed, err:", err.Error())
		log.Println("track:", string(debug.Stack()))
		return false
	}

	return true
}

//update one base data
func (d *BaseMysql) UpdateOneBaseData(
					dataByte []byte,
					whereMap map[string]interface{},
					table string,
					db *tc.DBService,
				) bool {
	var (
		tempStr string
		whereBuffer = bytes.NewBuffer(nil)
		values = make([]interface{}, 0)
	)

	//basic check
	if dataByte == nil || whereMap == nil ||
	   table == "" || db == nil {
		return false
	}

	//fit values
	values = append(values, dataByte)

	//format where sql
	if len(whereMap) > 0 {
		whereBuffer.WriteString(" WHERE ")
	}
	i := 0
	for field, val := range whereMap {
		tempStr = fmt.Sprintf("%s = ?", field)
		if i > 0 {
			whereBuffer.WriteString(" AND ")
		}
		whereBuffer.WriteString(tempStr)
		values = append(values, val)
		i++
	}

	//format sql
	sql := fmt.Sprintf("UPDATE %s SET data = ? %s",
		table,
		whereBuffer.String(),
	)

	//save into db
	_, _, err := db.Execute(sql, values...)
	if err != nil {
		log.Println("BaseMysql::UpdateOneBaseData failed, err:", err.Error())
		log.Println("track:", string(debug.Stack()))
		return false
	}
	return true
}

//increase or decrease field value
func (d *BaseMysql) UpdateCountOfOneData(
					updateMap map[string]interface{},
					whereMap map[string]WherePara,
					table string,
					db *tc.DBService,
				) bool {
	var (
		tempStr string
		updateBuffer = bytes.NewBuffer(nil)
		whereBuffer = bytes.NewBuffer(nil)
		values = make([]interface{}, 0)
	)

	//basic check
	if updateMap == nil || whereMap == nil ||
		table == "" || db == nil {
		return false
	}

	if len(updateMap) <= 0 || len(whereMap) <= 0 {
		return false
	}

	//format update field sql
	updateBuffer.WriteString("json_set(data ")
	for field, val := range updateMap {
		tempStr = fmt.Sprintf(", '$.%s', " +
					"GREATEST(json_extract(data, '$.%s') + ?, 0)",
					field, field)
		updateBuffer.WriteString(tempStr)
		values = append(values, val)
	}
	updateBuffer.WriteString(")")

	//format where sql
	whereBuffer, whereValues := d.formatWhereSql(whereMap)
	if whereValues != nil {
		values = append(values, whereValues...)
	}

	//format sql
	sql := fmt.Sprintf("UPDATE %s SET data = %s %s",
		table,
		updateBuffer.String(),
		whereBuffer.String(),
	)

	//save into db
	_, _, err := db.Execute(sql, values...)
	if err != nil {
		log.Println("BaseMysql::UpdateCountOfOneData failed, err:", err.Error())
		log.Println("track:", string(debug.Stack()))
		return false
	}
	return true
}

//update one data
func (d *BaseMysql) UpdateOneData(
				updateMap map[string]interface{},
				ObjArrMap map[string][]interface{},
				whereMap map[string]interface{},
				table string,
				db *tc.DBService,
			) bool {
	var (
		tempStr string
		updateBuffer = bytes.NewBuffer(nil)
		whereBuffer = bytes.NewBuffer(nil)
		values = make([]interface{}, 0)
		objectValSlice []interface{}
	)

	//basic check
	if updateMap == nil || whereMap == nil ||
	   table == "" || db == nil {
		return false
	}

	if len(updateMap) <= 0 || len(whereMap) <= 0 {
		return false
	}

	//format update field sql
	updateBuffer.WriteString("json_set(data ")
	for field, val := range updateMap {
		//reset object value slice
		objectValSlice = objectValSlice[:0]

		//check value kind
		//if hash map, need convert to json object kind
		v, isHashMap := val.(map[string]interface{})
		if isHashMap {
			//convert hash map into json object
			tempStr, objectValSlice = d.GenJsonObject(v)
			tempStr = fmt.Sprintf(",'$.%s', %s", field, tempStr)
		}else {
			tempStr = fmt.Sprintf(",'$.%s', ?", field)
		}
		if isHashMap {
			values = append(values, objectValSlice...)
		}else{
			values = append(values, val)
		}
		updateBuffer.WriteString(tempStr)
	}
	updateBuffer.WriteString(")")

	//check object array map
	if ObjArrMap != nil && len(ObjArrMap) > 0 {
		for field, objSlice := range ObjArrMap {
			tempSql, tempValues := d.GenJsonArrayAppendObject("data", field, objSlice)
			updateBuffer.WriteString(tempSql)
			values = append(values, tempValues...)
		}
	}

	//format where sql
	i := 0
	if whereMap != nil {
		whereBuffer.WriteString(" WHERE ")
	}
	for field, val := range whereMap {
		tempStr = fmt.Sprintf("%s = ?", field)
		if i > 0 {
			whereBuffer.WriteString(" AND ")
		}
		whereBuffer.WriteString(tempStr)
		values = append(values, val)
		i++
	}

	//format sql
	sql := fmt.Sprintf("UPDATE %s SET data = %s %s",
						table,
						updateBuffer.String(),
						whereBuffer.String(),
					)

	//save into db
	_, _, err := db.Execute(sql, values...)
	if err != nil {
		log.Println("BaseMysql::UpdateOneData failed, err:", err.Error())
		log.Println("track:", string(debug.Stack()))
		return false
	}
	return true
}

 //add data
func (d *BaseMysql) AddData(
				jsonByte []byte,
				table string,
				db *tc.DBService,
			) bool {
	//basic check
	if jsonByte == nil || db == nil {
		return false
	}

	//format sql
	sql := fmt.Sprintf("INSERT INTO %s(data)  VALUES(?)", table)
	values := []interface{}{
		jsonByte,
	}

	//save into db
	_, _, err := db.Execute(sql, values...)
	if err != nil {
		log.Println("BaseMysql::AddData failed, err:", err.Error())
		log.Println("track:", string(debug.Stack()))
		return false
	}

	return true
}


 //check and get json byte
func (d *BaseMysql) GetByteDataByField(
					field string,
					recordMap map[string]interface{},
				) []byte {
	v, ok := recordMap[field]
	if !ok {
		return nil
	}
	v2, ok := v.([]byte)
	if !ok {
		return nil
	}
	return v2
}

func (d *BaseMysql) GetByteData(recordMap map[string]interface{}) []byte {
	v, ok := recordMap[TableFieldOfData]
	if !ok {
		return nil
	}
	v2, ok := v.([]byte)
	if !ok {
		return nil
	}
	return v2
}

////////////////
//private func
////////////////

func (d *BaseMysql) getTotalVal(recordMap map[string]interface{}) int {
	v, ok := recordMap[TableFieldOfTotal]
	if !ok {
		return 0
	}
	v2, ok := v.([]uint8)
	if !ok {
		return 0
	}
	total, _ := strconv.Atoi(string(v2))
	return total
}

func (d *BaseMysql) formatWhereSql(
					whereMap map[string]WherePara,
				) (*bytes.Buffer, []interface{}) {
	var (
		tempStr string
		whereKind int
		whereBuffer = bytes.NewBuffer(nil)
		tempBuffer = bytes.NewBuffer(nil)
		values = make([]interface{}, 0)
	)

	if whereMap == nil || len(whereMap) <= 0 {
		return whereBuffer, nil
	}

	//format where sql
	i := 0
	whereBuffer.WriteString(" WHERE ")
	for field, wherePara := range whereMap {
		if i > 0 {
			whereBuffer.WriteString(" AND ")
		}
		whereKind = wherePara.Kind
		switch whereKind {
		case WhereKindOfIn:
			{
				tempSlice, ok := wherePara.Val.([]interface{})
				if ok {
					tempBuffer.Reset()
					tempStr = fmt.Sprintf("%s IN(", field)
					tempBuffer.WriteString(tempStr)
					k := 0
					for _, v := range tempSlice {
						if k > 0 {
							tempBuffer.WriteString(",")
						}
						tempBuffer.WriteString("?")
						values = append(values, v)
						k++
					}
					tempBuffer.WriteString(")")
					whereBuffer.WriteString(tempBuffer.String())
				}
			}
		case WhereKindOfInSet:
			{
				tempStr = fmt.Sprintf(" FIND_IN_SET(?, %s)", field)
				whereBuffer.WriteString(tempStr)
				values = append(values, wherePara.Val)
			}
		case WhereKindOfGen:
			fallthrough
		default:
			{
				tempStr = fmt.Sprintf("%s = ?", field)
				whereBuffer.WriteString(tempStr)
				values = append(values, wherePara.Val)
			}
		}
		i++
	}
	return whereBuffer, values
}