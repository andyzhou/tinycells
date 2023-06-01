package mysql

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"reflect"
	"runtime/debug"
	"strconv"
)

/*
 * mysql json format data face
 * - only for json format data
 * - use as anonymous class
 */

//where para
type WherePara struct {
	Kind int
	Condition string //used for `WhereKindOfAssigned`, for example ">", "<=", etc.
	Val interface{}
}

//face info
type JsonData struct {}

//get max value for assigned field
//field should be integer kind
func (f *JsonData) GetMaxVal(
			jsonField string,
			whereMap map[string]WherePara,
			objField string,
			table string,
			db *Connect,
		) (int64, error) {
	var (
		values = make([]interface{}, 0)
		max int64
	)

	//check
	if jsonField == "" || table == "" || db == nil {
		return max, errors.New("invalid parameter")
	}
	if objField == "" {
		objField = TableFieldOfData
	}

	//format where sql
	whereBuffer, whereValues := f.formatWhereSql(whereMap)
	if whereValues != nil {
		values = append(values, whereValues...)
	}

	//format sql
	sql := fmt.Sprintf("SELECT max(json_extract(%v, '$.%s')) as max FROM %s %s",
		objField, jsonField, table, whereBuffer.String(),
	)

	//query one record
	recordMap, err := db.GetRow(sql, values...)
	if err != nil {
		return max, err
	}
	max = f.getIntegerVal(TableFieldOfMax, recordMap)
	return max, nil
}

//sum assigned field count
//field should be integer kind
func (f *JsonData) SumCount(
			jsonField string,
			whereMap map[string]WherePara,
			objField string,
			table string,
			db *Connect,
		) (int64, error) {
	var (
		values = make([]interface{}, 0)
		total int64
	)

	//check
	if jsonField == "" || table == "" || db == nil {
		return total, errors.New("invalid parameter")
	}
	if objField == "" {
		objField = TableFieldOfData
	}

	//format where sql
	whereBuffer, whereValues := f.formatWhereSql(whereMap)
	if whereValues != nil {
		values = append(values, whereValues...)
	}

	//format sql
	sql := fmt.Sprintf("SELECT sum(json_extract(%v, '$.%s')) as total FROM %s %s",
		objField, jsonField, table, whereBuffer.String(),
	)

	//query one record
	recordMap, err := db.GetRow(sql, values...)
	if err != nil {
		return total, err
	}
	total = f.getIntegerVal(TableFieldOfTotal, recordMap)
	return total, nil
}

//get total num
func (f *JsonData) GetTotalNum(
			whereMap map[string]WherePara,
			table string,
			db *Connect,
		) (int64, error) {
	var (
		values = make([]interface{}, 0)
		total int64
	)

	//basic check
	if table == "" || db == nil {
		return total, errors.New("invalid parameter")
	}

	//format where sql
	whereBuffer, whereValues := f.formatWhereSql(whereMap)
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
		return total, err
	}

	total = f.getIntegerVal(TableFieldOfTotal, recordMap)
	return total, nil
}

//check and get json byte
func (f *JsonData) GetByteDataByField(
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

func (f *JsonData) GetByteData(
			recordMap map[string]interface{},
		) []byte {
	return f.GetByteDataAdv(TableFieldOfData, recordMap)
}

func (f *JsonData) GetByteDataAdv(
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


//get batch data with condition
func (f *JsonData) GetBatchData(
			whereMap map[string]WherePara,
			orderBy string,
			offset int,
			size int,
			table string,
			db *Connect,
		) ([][]byte, error) {
	recordsMap, err := f.GetBatchDataAdv(
		nil,
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
		jsonByte := f.GetByteData(recordMap)
		if jsonByte == nil {
			continue
		}
		result = append(result, jsonByte)
	}
	return result, nil
}

func (f *JsonData) GetBatchDataAdv(
			dataFields []string,
			whereMap map[string]WherePara,
			orderBy string,
			offset int,
			size int,
			table string,
			db *Connect,
		) ([]map[string]interface{}, error) {
	var (
		limitSql, orderBySql string
		values = make([]interface{}, 0)
		dataFieldBuffer = bytes.NewBuffer(nil)
	)

	//basic check
	if table == "" || db == nil {
		return nil, errors.New("invalid parameter")
	}

	//format data fields
	if dataFields == nil || len(dataFields) <= 0 {
		dataFields = []string{"data"}
	}
	i := 0
	for _, dataField := range dataFields {
		if i > 0 {
			dataFieldBuffer.WriteString(",")
		}
		dataFieldBuffer.WriteString(dataField)
		i++
	}

	//format where sql
	whereBuffer, whereValues := f.formatWhereSql(whereMap)
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
		dataFieldBuffer.String(),
		table,
		whereBuffer.String(),
		orderBySql,
		limitSql,
	)

	//query records
	recordsMap, err := db.GetRows(sql, values...)
	if err != nil {
		return nil, err
	}

	//check records map
	if recordsMap == nil || len(recordsMap) <= 0 {
		return nil, nil
	}
	return recordsMap, nil
}

//get batch random data
func (f *JsonData) GetBathRandomData(
			whereMap map[string]WherePara,
			size int,
			table string,
			db *Connect,
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
	whereBuffer, whereValues := f.formatWhereSql(whereMap)
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
	recordsMap, err := db.GetRows(sql, values...)
	if err != nil {
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
		jsonByte := f.GetByteData(recordMap)
		if jsonByte == nil {
			continue
		}
		result = append(result, jsonByte)
	}
	return result, nil
}

//get one data
//dataField default value is 'data'
func (f *JsonData) GetOneData(
			dataField string,
			whereMap map[string]WherePara,
			needRand bool,
			table string,
			db *Connect,
		) ([]byte, error) {
	if dataField == "" {
		dataField = TableFieldOfData
	}
	dataFields := []string{
		dataField,
	}
	byteMap, err := f.GetOneDataAdv(
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

func (f *JsonData) GetOneDataAdv(
			dataFields []string,
			whereMap map[string]WherePara,
			needRand bool,
			table string,
			db *Connect,
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
	whereBuffer, whereValues := f.formatWhereSql(whereMap)
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
		jsonByte := f.GetByteDataByField(dataField, recordMap)
		if jsonByte != nil {
			result[dataField] = jsonByte
		}
	}
	return result, nil
}

//add new data
func (f *JsonData) AddData(
			jsonByte []byte,
			table string,
			db *Connect,
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
	return f.AddDataAdv(
		dataMap,
		table,
		db,
	)
}


//delete data
func (f *JsonData) DelOneData(
			whereMap map[string]WherePara,
			table string,
			db *Connect,
		) error {
	return f.DelData(
		whereMap,
		table,
		db,
	)
}

func (f *JsonData) DelData(
			whereMap map[string]WherePara,
			table string,
			db *Connect,
		) error {
	var (
		values = make([]interface{}, 0)
	)

	//basic check
	if whereMap == nil || table == "" || db == nil {
		return errors.New("invalid parameter")
	}

	//format where sql
	whereBuffer, whereValues := f.formatWhereSql(whereMap)
	if whereValues != nil {
		values = append(values, whereValues...)
	}

	//format sql
	sql := fmt.Sprintf("DELETE FROM %s %s", table, whereBuffer.String())

	//remove from db
	_, _, err := db.Execute(sql, values...)
	return err
}

//update one base data
func (f *JsonData) UpdateBaseData(
			dataByte []byte,
			whereMap map[string]WherePara,
			table string,
			db *Connect,
		) error {
	return f.UpdateBaseDataAdv("", dataByte, whereMap, table, db)
}

func (f *JsonData) UpdateBaseDataAdv(
			dataField string,
			dataByte []byte,
			whereMap map[string]WherePara,
			table string,
			db *Connect,
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
	whereBuffer, whereValues := f.formatWhereSql(whereMap)
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
	return err
}

//increase or decrease field value
func (f *JsonData) UpdateCountOfData(
			updateMap map[string]interface{},
			whereMap map[string]WherePara,
			table string,
			db *Connect,
			isOverWrites ...bool,
		) error {
	return f.UpdateCountOfDataAdv(
		updateMap,
		whereMap,
		TableFieldOfData,
		table,
		db,
		isOverWrites...,
	)
}

func (f *JsonData) UpdateCountOfDataAdv(
			updateMap map[string]interface{},
			whereMap map[string]WherePara,
			objField string,
			table string,
			db *Connect,
			isOverWrites ...bool,
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

	//is just over write new count value?
	isOverWrite := false
	if isOverWrites != nil && len(isOverWrites) > 0 {
		isOverWrite = isOverWrites[0]
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
		if isOverWrite {
			//overwrite count new value
			tempStr = fmt.Sprintf(", '$.%s', IFNULL(%s->'$.%s', %v), '$.%s', ?",
				field, objField, field, objDefaultVal, field)
		}else{
			//inc count new value
			tempStr = fmt.Sprintf(", '$.%s', IFNULL(%s->'$.%s', %v), '$.%s', " +
				"GREATEST(IFNULL(json_extract(%s, '$.%s'), 0) + ?, 0)",
				field, objField, field, objDefaultVal, field, objField, field)
		}
		updateBuffer.WriteString(tempStr)
		values = append(values, val)
	}
	updateBuffer.WriteString(")")

	//format where sql
	whereBuffer, whereValues := f.formatWhereSql(whereMap)
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
	return err
}

//update data
func (f *JsonData) UpdateData(
			updateMap map[string]interface{},
			ObjArrMap map[string][]interface{},
			whereMap map[string]WherePara,
			table string,
			db *Connect,
		) error {
	return f.UpdateDataAdv(
		updateMap,
		ObjArrMap,
		whereMap,
		TableFieldOfData,
		table,
		db,
	)
}

//update data adv
func (f *JsonData) UpdateDataAdv(
			updateMap map[string]interface{},
			objArrMap map[string][]interface{},
			whereMap map[string]WherePara,
			objField string,
			table string,
			db *Connect,
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
		objField = TableFieldOfData
	}

	//format update field sql
	tempStr = fmt.Sprintf("json_set(%s ", objField)
	updateBuffer.WriteString(tempStr)
	for field, val := range updateMap {
		//reset object value slice
		//isHashMap = false
		subSql = ""
		objectValSlice = objectValSlice[:0]

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
					subSql, objectValSlice = f.GenJsonArray(genSlice)
				}
			}
		case map[string]interface{}:
			{
				objDefaultVal = "JSON_OBJECT()"
				genMap, isOk = val.(map[string]interface{})
				if isOk {
					subSql, objectValSlice = f.GenJsonObject(genMap)
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
			tempSql, tempValues := f.GenJsonArrayAppendObject(objField, field, objSlice)
			updateBuffer.WriteString(tempSql)
			values = append(values, tempValues...)
		}
	}

	//format where sql
	whereBuffer, whereValues := f.formatWhereSql(whereMap)
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
	return err
}

//add data
//support multi json data fields
func (f *JsonData) AddDataAdv(
			dataMap map[string][]byte,
			table string,
			db *Connect,
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
	return err
}


//add data with on duplicate update
//if isInc opt, just increase field value
func (f *JsonData) AddDataWithDuplicate(
			jsonByte []byte,
			updateMap map[string]interface{},
			isInc bool,
			objField string,
			table string,
			db *Connect,
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
	if objField == "" {
		objField = TableFieldOfData
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
			tempStr = fmt.Sprintf("'$.%s', GREATEST(json_extract(%v, '$.%s') + ?, 0)",
				field, objField, field)
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
	return err
}


//create json_array sql pass json data slice
func (f *JsonData) GenJsonArrayAppendObject(
			tabField string,
			dataField string,
			jsonSlice []interface{},
		) (string, []interface{}) {
	var (
		buffer = bytes.NewBuffer(nil)
		tempStr string
		values = make([]interface{}, 0)
	)

	//basic check
	if tabField == "" || dataField == "" ||
		jsonSlice == nil || len(jsonSlice) <= 0 {
		return buffer.String(), values
	}

	//check data field
	tempStr = fmt.Sprintf(", %s = JSON_SET(%s, '$.%s', IFNULL(%s->'$.%s',JSON_ARRAY()))",
		tabField, tabField, dataField, tabField, dataField)
	buffer.WriteString(tempStr)

	//convert into relate data
	i := 0
	tempStr = fmt.Sprintf(", %s = JSON_ARRAY_APPEND(%s, ", tabField, tabField)
	buffer.WriteString(tempStr)
	for _, v := range jsonSlice {
		if i > 0 {
			buffer.WriteString(" ,")
		}
		tempStr = fmt.Sprintf("'$.%s', CAST(? AS JSON)", dataField)
		buffer.WriteString(tempStr)
		values = append(values, v)
		i++
	}
	buffer.WriteString(")")
	return buffer.String(), values
}

//create json_object sql pass json data map
//return subSql, values
func (f *JsonData) GenJsonArray(
			valSlice []interface{},
		) (string, []interface{}) {
	var (
		buffer = bytes.NewBuffer(nil)
		values = make([]interface{}, 0)
	)
	//basic check
	if valSlice == nil || len(valSlice) <= 0 {
		return buffer.String(), values
	}

	arrayBuffer := bytes.NewBuffer(nil)
	arrayBuffer.WriteString("JSON_ARRAY(")
	for k, v2 := range valSlice {
		if k > 0 {
			arrayBuffer.WriteString(",")
		}
		arrayBuffer.WriteString("?")
		values = append(values, v2)
	}
	arrayBuffer.WriteString(")")
	return arrayBuffer.String(), values
}


//for general map
func (f *JsonData) GenJsonObject(
			genHashMap map[string]interface{},
		) (string, []interface{}) {
	var (
		buffer = bytes.NewBuffer(nil)
		tempStr string
		values = make([]interface{}, 0)
	)

	//basic check
	if genHashMap == nil || len(genHashMap) <= 0 {
		return buffer.String(), values
	}

	//convert into relate data
	i := 0
	buffer.WriteString("json_object(")
	for k, v := range genHashMap {
		if i > 0 {
			buffer.WriteString(" ,")
		}
		//check value is array kind or not
		v1, isArray := v.([]string)
		if isArray {
			//format sub sql
			arrayBuffer := bytes.NewBuffer(nil)
			arrayBuffer.WriteString("JSON_ARRAY(")
			for k, v2 := range v1 {
				if k > 0 {
					arrayBuffer.WriteString(",")
				}
				arrayBuffer.WriteString("?")
				values = append(values, v2)
			}
			arrayBuffer.WriteString(")")

			//is array format
			tempStr = fmt.Sprintf("'%s', %s", k, arrayBuffer.String())
		}else{
			tempStr = fmt.Sprintf("'%s', ?", k)
			values = append(values, v)
		}
		buffer.WriteString(tempStr)
		i++
	}
	buffer.WriteString(")")

	return buffer.String(), values
}

////////////////
//private func
////////////////

//convert field value into integer format
func (f *JsonData) getIntegerVal(
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

//format where sql
func (f *JsonData) formatWhereSql(
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
				valType := reflect.TypeOf(wherePara.Val)
				if valType.Kind() == reflect.Slice {
					refVal := reflect.ValueOf(wherePara.Val)
					for i := 0; i < refVal.Len(); i++ {
						tempSlice = append(tempSlice, refVal.Index(i).Interface())
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