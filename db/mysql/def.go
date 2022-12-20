package mysql

const (
	LazyCommandChanSize = 1024
	ConnCheckRate = 20 //xxx seconds
	DBPoolMin = 1
	DBPoolMax = 64
)

//internal db server status
const (
	DBStateActive = 1
	DBStateDown = 0
)