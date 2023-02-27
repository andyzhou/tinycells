package util

import (
	"math/rand"
	"time"
)

//face info
type Util struct {
}

//construct
func NewUtil() *Util {
	this := &Util{}
	return this
}

//get rand number
func (u *Util) GetRandomVal(maxVal int) int {
	randSand := rand.NewSource(time.Now().UnixNano())
	r := rand.New(randSand)
	return r.Intn(maxVal)
}
func (u *Util) GetRealRandomVal(maxVal int) int {
	return int(rand.Float64() * 1000) % (maxVal)
}