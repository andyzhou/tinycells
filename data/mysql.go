package data

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/andyzhou/tinycells/db"
	"github.com/andyzhou/tinycells/tc"
	"log"
	"runtime/debug"
	"strconv"
)

/*
 * base mysql db data face
 * - only for json format data
 * - use as anonymous class
 */

 //const db field
 const (
 	TableFieldOfMax = "max"
 	TableFieldOfTotal = "total"
 	TableFieldOfData = "data" //db data field
 )

 //where kind
 const (
 	WhereKindOfGen = iota
 	WhereKindOfIn	  //for in('x','y')
 	WhereKindOfInSet  //for FIND_IN_SET(val, `x`, 'y')
 	WhereKindOfAssigned //for assigned condition, like '>', '<', '!=', etc.
 )

 //where para
 type WherePara struct {
 	Kind int
 	Condition string //used for `WhereKindOfAssigned`, for example ">", "<=", etc.
 	Val interface{}
 }

 //face info
 type BaseMysql struct {
 	tc.Utils
 }

 //get max value for assigned field
//field should be integer kind
func (d *BaseMysql) GetMaxVal(
						jsonField string,
						whereMap map[string]WherePara,
						table string,
						db *db.Mysql,
					) (int64, error) {
	var (
		values = make([]interface{}, 0)
		max int64
	)

	//basic check
	if jsonField == "" || table == "" || db == nil {
		return max, errors.New("invalid parameter")
	}

	//format where sql
	whereBuffer, whereValues := d.formatWhereSql(whereMap)
	if whereValues != nil {
		values = append(values, whereValues...)
	}

	//format sql
	sql := fmt.Sprintf("SELECT max(json_extract(data, '$.%s')) as max FROM %s %s",
		jsonField, table, whereBuffer.String(),
	)

	//query one record
	recordMap, err := db.GetRow(sql, values...)
	if err != nil {
		log.Println("BaseMysql::GetMaxVal failed, err:", err.Error())
		log.Println("track:", string(debug.Stack()))
		return max, err
	}
	max = d.getIntegerVal(TableFieldOfMax, recordMap)
	return max, nil
}

//sum assigned field count
 //field should be integer kind
func (d *BaseMysql) SumCount(
						jsonField string,
						whereMap map[string]WherePara,
						table string,
						db *db.Mysql,
					) int64 {
	var (
		values = make([]interface{}, 0)
		total int64
	)

	//basic check
	if jsonField == "" || table == "" || db == nil {
		return total
	}

	//format where sql
	whereBuffer, whereValues := d.formatWhereSql(whereMap)
	if whereValues != nil {
		values = append(values, whereValues...)
	}

	//format sql
	sql := fmt.Sprintf("SELECT sum(json_extract(data, '$.%s')) as total FROM %s %s",
						jsonField, table, whereBuffer.String(),
					)

	//query one record
	recordMap, err := db.GetRow(sql, values...)
	if err != nil {
		log.Println("BaseMysql::SumCount failed, err:", err.Error())
		log.Println("track:", string(debug.Stack()))
		return total
	}
	total = d.getIntegerVal(TableFieldOfTotal, recordMap)
	return total
}


 //get total num
func (d *BaseMysql) GetTotalNum(
					whereMap map[string]WherePara,
					table string,
					db *db.Mysql,
				) int64 {
	var (
		values = make([]interface{}, 0)
		total int64
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

	total = d.getIntegerVal(TableFieldOfTotal, recordMap)
	return total
}

 //get batch data with condition
func (d *BaseMysql) GetBatchData(
				whereMap map[string]WherePara,
				orderBy string,
				offset int,
				size int,
				table string,
				db *db.Mysql,
			) ([][]byte, error) {

	recordsMap, err := d.GetBatchDataAdv(
				"",
				whereMap,
				orderBy,
				offset,
				size,
				table,
				db,
			)
	//check records map
	if err != nil {
		return nil, err
	}
	if recordsMap == nil || len(recordsMap) <= 0 {
		return nil, nil
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
	return result, nil
}

func (d *BaseMysql) GetBatchDataAdv(
				selectFields string,
				whereMap map[string]WherePara,
				orderBy string,
				offset int,
				size int,
				table string,
				db *db.Mysql,
			) ([]map[string]interface{}, error) {
	var (
		limitSql, orderBySql string
		values = make([]interface{}, 0)
	)

	//basic check
	if table == "" || db == nil {
		return nil, errors.New("invalid paramter")
	}

	if selectFields == "" {
		selectFields = "data"
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
	sql := fmt.Sprintf("SELECT %s FROM %s %s %s %s",
						selectFields,
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
		return nil, err
	}

	//check records map
	if recordsMap == nil || len(recordsMap) <= 0 {
		return nil, nil
	}

	return recordsMap, nil
}

//get batch random data
func (d *BaseMysql) GetBathRandomData(
						whereMap map[string]WherePara,
						size int,
						table string,
						db *db.Mysql,
					) ([][]byte, error) {
	var (
		limitSql string
		values = make([]interface{}, 0)
	)

	//basic check
	if table == "" || db == nil {
		return nil, errors.New("invalid parameter")
	}

	//format limit sql
	if size > 0 {
		limitSql = fmt.Sprintf("LIMIT %d", size)
	}

	//format where sql
	whereBuffer, whereValues := d.formatWhereSql(whereMap)
	if whereValues != nil {
		values = append(values, whereValues...)
	}

	//format sql
	sql := fmt.Sprintf("SELECT data FROM %s %s ORDER BY RAND() %s",
		table,
		whereBuffer.String(),
		limitSql,
	)

	//query records
	recordsMap, err := db.GetArray(sql, values...)
	if err != nil {
		log.Println("BaseMysql::GetBathRandomData failed, err:", err.Error())
		log.Println("track:", string(debug.Stack()))
		return nil, err
	}

	//check records map
	if recordsMap == nil || len(recordsMap) <= 0 {
		return nil, nil
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
	return result, nil
}

 //get one data
 //dataField default value is 'data'
func (d *BaseMysql) GetOneData(
				dataField string,
				whereMap map[string]WherePara,
				needRand bool,
				table string,
				db *db.Mysql,
			) ([]byte, error) {
	if dataField == "" {
		dataField = "data"
	}
	dataFields := []string{
		dataField,
	}
	byteMap, err := d.GetOneDataAdv(
			dataFields,
			whereMap,
			needRand,
			table,
			db,
		)
	if err != nil {
		return nil, err
	}
	if byteMap == nil {
		return nil, nil
	}
	v, ok := byteMap[dataField]
	if !ok {
		return nil, nil
	}
	return v, nil
}

func (d *BaseMysql) GetOneDataAdv(
				dataFields []string,
				whereMap map[string]WherePara,
				needRand bool,
				table string,
				db *db.Mysql,
			) (map[string][]byte, error) {
	var (
		//assignedDataField string
		dataFieldBuffer = bytes.NewBuffer(nil)
		orderBy string
		values = make([]interface{}, 0)
	)

	//basic check
	if table == "" || db == nil {
		return nil, errors.New("invalid parameter")
	}

	//format where sql
	whereBuffer, whereValues := d.formatWhereSql(whereMap)
	if whereValues != nil {
		values = append(values, whereValues...)
	}

	if dataFields != nil && len(dataFields) > 0 {
		i := 0
		for _, dataField := range dataFields {
			if i > 0 {
				dataFieldBuffer.WriteString(",")
			}
			dataFieldBuffer.WriteString(dataField)
			i++
		}
	}else{
		dataFieldBuffer.WriteString("data")
	}

	if needRand {
		orderBy = fmt.Sprintf(" ORDER BY RAND()")
	}

	//format sql
	sql := fmt.Sprintf("SELECT %s FROM %s %s %s",
						dataFieldBuffer.String(),
						table,
						whereBuffer.String(),
						orderBy,
					)

	//query records
	recordMap, err := db.GetRow(sql, values...)
	if err != nil {
		log.Println("BaseMysql::GetOneData failed, err:", err.Error())
		log.Println("track:", string(debug.Stack()))
		return nil, err
	}

	//check record map
	if recordMap == nil || len(recordMap) <= 0 {
		return nil, nil
	}

	//format result
	result := make(map[string][]byte)

	//get json byte data
	for _, dataField := range dataFields {
		jsonByte := d.GetByteDataByField(dataField, recordMap)
		if jsonByte != nil {
			result[dataField] = jsonByte
		}
	}
	return result, nil
}


//delete data
func (d *BaseMysql) DelOneData(
				whereMap map[string]WherePara,
				table string,
				db *db.Mysql,
			) error {
	return d.DelData(
			whereMap,
			table,
			db,
		)
}

func (d *BaseMysql) DelData(
				whereMap map[string]WherePara,
				table string,
				db *db.Mysql,
			) error {
	var (
		values = make([]interface{}, 0)
	)

	//basic check
	if whereMap == nil || table == "" || db == nil {
		return errors.New("invalid parameter")
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
		return err
	}

	return nil
}

//update one base data
func (d *BaseMysql) UpdateBaseData(
					dataByte []byte,
					whereMap map[string]WherePara,
					table string,
					db *db.Mysql,
				) error {
	return d.UpdateBaseDataAdv("", dataByte, whereMap, table, db)
}

func (d *BaseMysql) UpdateBaseDataAdv(
					dataField string,
					dataByte []byte,
					whereMap map[string]WherePara,
					table string,
					db *db.Mysql,
				) error {
	var (
		whereBuffer = bytes.NewBuffer(nil)
		values = make([]interface{}, 0)
	)

	//basic check
	if dataByte == nil || whereMap == nil ||
	   table == "" || db == nil {
		return errors.New("invalid parameter")
	}

	if dataField == "" {
		dataField = "data"
	}

	//fit values
	values = append(values, dataByte)

	//format where sql
	whereBuffer, whereValues := d.formatWhereSql(whereMap)
	if whereValues != nil {
		values = append(values, whereValues...)
	}

	//format sql
	sql := fmt.Sprintf("UPDATE %s SET %s = ? %s",
		table,
		dataField,
		whereBuffer.String(),
	)

	//save into db
	_, _, err := db.Execute(sql, values...)
	if err != nil {
		log.Println("BaseMysql::UpdateOneBaseData failed, err:", err.Error())
		log.Println("track:", string(debug.Stack()))
		return err
	}
	return nil
}

//increase or decrease field value
func (d *BaseMysql) UpdateCountOfData(
					updateMap map[string]interface{},
					whereMap map[string]WherePara,
					table string,
					db *db.Mysql,
				) error {
	return d.UpdateCountOfDataAdv(
			updateMap,
			whereMap,
			"data",
			table,
			db,
		)
}
func (d *BaseMysql) UpdateCountOfDataAdv(
					updateMap map[string]interface{},
					whereMap map[string]WherePara,
					objField string,
					table string,
					db *db.Mysql,
				) error {
	var (
		tempStr string
		updateBuffer = bytes.NewBuffer(nil)
		whereBuffer = bytes.NewBuffer(nil)
		values = make([]interface{}, 0)
		objDefaultVal interface{}
	)

	//basic check
	if updateMap == nil || whereMap == nil ||
		table == "" || db == nil {
		return errors.New("invalid parameter")
	}

	if len(updateMap) <= 0 || len(whereMap) <= 0 {
		return errors.New("update map is nil")
	}

	if objField == "" {
		objField = "data"
	}

	//format update field sql
	tempStr = fmt.Sprintf("json_set(%s ", objField)
	updateBuffer.WriteString(tempStr)
	for field, val := range updateMap {
		switch val.(type) {
		case float64:
			objDefaultVal = 0.0
		default:
			objDefaultVal = 0
		}
		tempStr = fmt.Sprintf(", '$.%s', IFNULL(%s->'$.%s', %v), '$.%s', " +
					"GREATEST(IFNULL(json_extract(data, '$.%s'), 0) + ?, 0)",
					field, objField, field, objDefaultVal, field, field)
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
	sql := fmt.Sprintf("UPDATE %s SET %s = %s %s",
		table,
		objField,
		updateBuffer.String(),
		whereBuffer.String(),
	)

	//save into db
	_, _, err := db.Execute(sql, values...)
	if err != nil {
		log.Println("BaseMysql::UpdateCountOfOneData failed, err:", err.Error())
		log.Println("track:", string(debug.Stack()))
		return err
	}
	return nil
}

//update data
func (d *BaseMysql) UpdateData(
				updateMap map[string]interface{},
				ObjArrMap map[string][]interface{},
				whereMap map[string]WherePara,
				table string,
				db *db.Mysql,
			) error {
	return d.UpdateDataAdv(
			updateMap,
			ObjArrMap,
			whereMap,
			"data",
			table,
			db,
		)
}

func (d *BaseMysql) UpdateDataAdv(
				updateMap map[string]interface{},
				objArrMap map[string][]interface{},
				whereMap map[string]WherePara,
				objField string,
				table string,
				db *db.Mysql,
			) error {
	var (
		tempStr string
		updateBuffer = bytes.NewBuffer(nil)
		whereBuffer = bytes.NewBuffer(nil)
		values = make([]interface{}, 0)
		objectValSlice []interface{}
		objDefaultVal interface{}
		subSql string
		//isHashMap bool
		genMap map[string]interface{}
		genSlice []interface{}
		isOk bool
	)

	//basic check
	if updateMap == nil || whereMap == nil ||
	   table == "" || db == nil {
		return errors.New("invalid parameter")
	}

	if len(updateMap) <= 0 || len(whereMap) <= 0 {
		return errors.New("update map is nil")
	}

	if objField == "" {
		objField = "data"
	}

	//format update field sql
	tempStr = fmt.Sprintf("json_set(%s ", objField)
	updateBuffer.WriteString(tempStr)
	for field, val := range updateMap {
		//reset object value slice
		//isHashMap = false
		subSql = ""
		objectValSlice = objectValSlice[:0]

		//fmt.Println("field:", field, ", val:", val, ", type:", reflect.TypeOf(val))

		//check value kind
		//if hash map, need convert to json object kind
		switch val.(type) {
		case float64:
			objDefaultVal = 0.0
		case int64:
			objDefaultVal = 0
		case int:
			objDefaultVal = 0
		case bool:
			objDefaultVal = false
		case string:
			objDefaultVal = "''"
		case []interface{}:
			{
				objDefaultVal = "JSON_ARRAY()"
				genSlice, isOk = val.([]interface{})
				if isOk {
					subSql, objectValSlice = d.GenJsonArray(genSlice)
				}
			}
		case map[string]interface{}:
			{
				objDefaultVal = "JSON_OBJECT()"
				genMap, isOk = val.(map[string]interface{})
				if isOk {
					subSql, objectValSlice = d.GenJsonObject(genMap)
				}
			}
		default:
			{
				objDefaultVal = "JSON_OBJECT()"
			}
		}

		//format sub sql
		if subSql != "" {
			tempStr = fmt.Sprintf(", '$.%s', IFNULL(%s->'$.%s', %v)" +
				",'$.%s', %s", field, objField, field,
				objDefaultVal, field, subSql)
			values = append(values, objectValSlice...)
		}else{
			tempStr = fmt.Sprintf(", '$.%s', IFNULL(%s->'$.%s', %v)" +
				",'$.%s', ?", field, objField, field,
				objDefaultVal, field)
			values = append(values, val)
		}
		updateBuffer.WriteString(tempStr)
	}
	updateBuffer.WriteString(")")

	//check object array map
	if objArrMap != nil && len(objArrMap) > 0 {
		for field, objSlice := range objArrMap {
			tempSql, tempValues := d.GenJsonArrayAppendObject(objField, field, objSlice)
			updateBuffer.WriteString(tempSql)
			values = append(values, tempValues...)
		}
	}

	//format where sql
	whereBuffer, whereValues := d.formatWhereSql(whereMap)
	if whereValues != nil {
		values = append(values, whereValues...)
	}

	//format sql
	sql := fmt.Sprintf("UPDATE %s SET %s = %s %s",
						table,
						objField,
						updateBuffer.String(),
						whereBuffer.String(),
					)

	//save into db
	_, _, err := db.Execute(sql, values...)
	if err != nil {
		log.Println("BaseMysql::UpdateOneData failed, err:", err.Error())
		log.Println("track:", string(debug.Stack()))
		return err
	}
	return nil
}


 //add new data
func (d *BaseMysql) AddData(
				jsonByte []byte,
				table string,
				db *db.Mysql,
			) error {
	//basic check
	if jsonByte == nil || db == nil {
		return errors.New("invalid parameter")
	}

	//format data map
	dataMap := map[string][]byte {
		TableFieldOfData:jsonByte,
	}

	//call base func
	return d.AddDataAdv(
				dataMap,
				table,
				db,
			)
}


//add data
//support multi json data fields
func (d *BaseMysql) AddDataAdv(
				dataMap map[string][]byte,
				table string,
				db *db.Mysql,
			) error {
	var (
		buffer = bytes.NewBuffer(nil)
		valueBuffer = bytes.NewBuffer(nil)
		values = make([]interface{}, 0)
		tempStr string
	)

	//basic check
	if dataMap == nil || db == nil {
		return errors.New("invalid parameter")
	}

	tempStr = fmt.Sprintf("INSERT INTO %s(", table)
	buffer.WriteString(tempStr)
	valueBuffer.WriteString(" VALUES(")

	i := 0
	for k, v := range dataMap {
		if i > 0 {
			buffer.WriteString(",")
			valueBuffer.WriteString(",")
		}
		tempStr = fmt.Sprintf("?")
		valueBuffer.WriteString(tempStr)

		tempStr = fmt.Sprintf("%s", k)
		buffer.WriteString(tempStr)
		values = append(values, v)
		i++
	}
	valueBuffer.WriteString(")")


	buffer.WriteString(")")
	buffer.WriteString(valueBuffer.String())

	//save into db
	_, _, err := db.Execute(buffer.String(), values...)
	if err != nil {
		log.Println("BaseMysql::AddDataAdv failed, err:", err.Error())
		log.Println("track:", string(debug.Stack()))
		return err
	}

	return nil
}

//add data with on duplicate update
//if isInc opt, just increase field value
func (d *BaseMysql) AddDataWithDuplicate(
						jsonByte []byte,
						updateMap map[string]interface{},
						isInc bool,
						table string,
						db *db.Mysql,
					) error {
	var (
		tempStr string
		updateBuffer = bytes.NewBuffer(nil)
		values = make([]interface{}, 0)
	)

	//basic check
	if jsonByte == nil || db == nil || updateMap == nil {
		return errors.New("invalid parameter")
	}

	//init update buffer
	tempStr = fmt.Sprintf("data = json_set(data, ")
	updateBuffer.WriteString(tempStr)

	values = append(values, jsonByte)
	i := 0
	for field, v := range updateMap {
		if i > 0 {
			updateBuffer.WriteString(", ")
		}
		if isInc {
			tempStr = fmt.Sprintf("'$.%s', GREATEST(json_extract(data, '$.%s') + ?, 0)",
				field, field)
		}else{
			tempStr = fmt.Sprintf("'$.%s', ?", field)
		}
		values = append(values, v)
		updateBuffer.WriteString(tempStr)
		i++
	}

	//fill update buffer
	updateBuffer.WriteString(")")

	//format sql
	sql := fmt.Sprintf("INSERT INTO %s(data)  VALUES(?) ON DUPLICATE KEY UPDATE %s",
		table, updateBuffer.String(),
	)

	//save into db
	_, _, err := db.Execute(sql, values...)
	if err != nil {
		log.Println("BaseMysql::AddDataWithDuplicate failed, err:", err.Error())
		log.Println("track:", string(debug.Stack()))
		return err
	}

	return nil
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

func (d *BaseMysql) GetByteData(
					recordMap map[string]interface{},
				) []byte {
	return d.GetByteDataAdv(TableFieldOfData, recordMap)
}

func (d *BaseMysql) GetByteDataAdv(
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

////////////////
//private func
////////////////

func (d *BaseMysql) getIntegerVal(
						field string,
						recordMap map[string]interface{},
					) int64 {
	v, ok := recordMap[field]
	if !ok {
		return 0
	}
	v2, ok := v.([]uint8)
	if !ok {
		v3, ok := v.(int64)
		if ok {
			return v3
		}
		return 0
	}
	intVal, _ := strconv.ParseInt(string(v2), 10, 64)
	return intVal
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
				tempSlice := make([]interface{}, 0)
				switch wherePara.Val.(type) {
				case []interface{}:
					tempSlice, _ = wherePara.Val.([]interface{})
				case []int32:
					for _, v := range wherePara.Val.([]int32) {
						tempSlice = append(tempSlice, v)
					}
				case []int:
					for _, v := range wherePara.Val.([]int) {
						tempSlice = append(tempSlice, v)
					}
				case []int64:
					for _, v := range wherePara.Val.([]int64) {
						tempSlice = append(tempSlice, v)
					}
				case []string:
					for _, v := range wherePara.Val.([]string) {
						tempSlice = append(tempSlice, v)
					}
				}
				if tempSlice != nil {
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
		case WhereKindOfAssigned:
			{
				//like field >= value, etc.
				tempStr = fmt.Sprintf("%s %s ?", field, wherePara.Condition)
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