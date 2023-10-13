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

//deep copy object
func (f *Util) DeepCopy(src, dist interface{}) (err error){
	buf := bytes.Buffer{}
	if err = gob.NewEncoder(&buf).Encode(src); err != nil {
		return
	}
	return gob.NewDecoder(&buf).Decode(dist)
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