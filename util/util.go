package util

import (
	"crypto/md5"
	"encoding/hex"
	"math/rand"
)

//face info
type Util struct {
}

//construct
func NewUtil() *Util {
	this := &Util{}
	return this
}

//gen rand value between min and max
func (u *Util) RandMinMax(min, max int64) int64 {
	return u.RandWithBelowMinMax(min, 0, max-min)
}

func (u *Util) RandWithBelowMinMax(below, min, max int64) int64 {
	max++
	offset := max - min
	if offset <= 0 {
		return below + min
	}
	return below + rand.Int63n(offset) + min
}

//gen md5 string
func (u *Util) GenMd5(orgString string) string {
	if len(orgString) <= 0 {
		return ""
	}
	m := md5.New()
	m.Write([]byte(orgString))
	return hex.EncodeToString(m.Sum(nil))
}