package util

import (
	"math"
	"strconv"
	"strings"
)

/*
 * number format face
 * - format number as *.K[M|B|T]
 */

//face info
type Number struct {
}

//construct
func NewNumber() *Number {
	this := &Number{}
	return this
}

func (f *Number) NearestThousandFormat(num float64) string {
	if math.Abs(num) < 999.5 {
		xNum := f.FormatNumber(num)
		xNumStr := xNum[:len(xNum)-3]
		return xNumStr
	}

	xNum := f.FormatNumber(num)
	// first, remove the .00 then convert to slice
	xNumStr := xNum[:len(xNum)-3]
	xNumCleaned := strings.Replace(xNumStr, ",", " ", -1)
	xNumSlice := strings.Fields(xNumCleaned)
	count := len(xNumSlice) - 2
	unit := [4]string{"K", "M", "B", "T"}
	xPart := unit[count]

	afterDecimal := ""
	if xNumSlice[1][0] != 0 {
		afterDecimal = "." + string(xNumSlice[1][0])
	}
	final := xNumSlice[0] + afterDecimal + xPart
	return final
}

func (f *Number) FormatNumber(input float64) string {
	x := f.RoundInt(input)
	xFormatted := f.NumberFormat(float64(x), 2, ".", ",")
	return xFormatted
}

// credit to https://github.com/DeyV/gotools/blob/master/numbers.go
func (f *Number) RoundPre(x float64, pre int) float64 {
	if math.IsNaN(x) || math.IsInf(x, 0) {
		return x
	}

	sign := 1.0
	if x < 0 {
		sign = -1
		x *= -1
	}

	var rounder float64
	pow := math.Pow(10, float64(pre))
	interred := x * pow
	_, frac := math.Modf(interred)

	if frac >= 0.5 {
		rounder = math.Ceil(interred)
	} else {
		rounder = math.Floor(interred)
	}

	return rounder / pow * sign
}

func (f *Number) NumberFormat(
		number float64,
		decimals int,
		decPoint, thousandsSep string) string {
	if math.IsNaN(number) || math.IsInf(number, 0) {
		number = 0
	}

	var ret string
	var negative bool

	if number < 0 {
		number *= -1
		negative = true
	}

	d, fact := math.Modf(number)
	if decimals <= 0 {
		fact = 0
	} else {
		pow := math.Pow(10, float64(decimals))
		fact = f.RoundPre(fact*pow, 0)
	}

	if thousandsSep == "" {
		ret = strconv.FormatFloat(d, 'f', 0, 64)
	} else if d >= 1 {
		var x float64
		for d >= 1 {
			d, x = math.Modf(d / 1000)
			x = x * 1000
			ret = strconv.FormatFloat(x, 'f', 0, 64) + ret
			if d >= 1 {
				ret = thousandsSep + ret
			}
		}
	} else {
		ret = "0"
	}

	facts := strconv.FormatFloat(fact, 'f', 0, 64)

	// "0" pad left
	for i := len(facts); i < decimals; i++ {
		facts = "0" + facts
	}

	ret += decPoint + facts
	if negative {
		ret = "-" + ret
	}
	return ret
}

func (f *Number) RoundInt(input float64) int {
	var result float64

	if input < 0 {
		result = math.Ceil(input - 0.5)
	} else {
		result = math.Floor(input + 0.5)
	}

	// only interested in integer, ignore fractional
	i, _ := math.Modf(result)
	return int(i)
}