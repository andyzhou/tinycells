package mysql

const (
	ConnCheckRate = 20 //xxx seconds
	DBPoolMin = 1
)

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