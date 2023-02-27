package util

import (
	"math"
	"math/rand"
)

//round float value
func (u *Util) Round(f float64, n int) float64 {
	n10 := math.Pow10(n)
	return math.Trunc((f+0.5/n10)*n10) / n10
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
