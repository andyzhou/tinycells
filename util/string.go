package util

import (
	"crypto/md5"
	"encoding/hex"
	"unicode"
)

//convert slice to string
func (u *Util) Slice2Str(orgSlice []string) string {
	var result string
	if len(orgSlice) <= 0 {
		return result
	}
	for _, v := range orgSlice {
		result += v
	}
	return result
}

//sub string, support utf8 string
func (u *Util) SubString(source string, start int, length int) string {
	rs := []rune(source)
	len := len(rs)
	if start < 0 {
		start = 0
	}
	if start >= len {
		start = len
	}
	end := start + length
	if end > len {
		end = len
	}
	return string(rs[start:end])
}

//lower first character
func (u *Util) LcFirst(str string) string {
	for i, v := range str {
		return string(unicode.ToLower(v)) + str[i+1:]
	}
	return ""
}

//upper first character
func (u *Util) UcFirst(str string) string {
	for i, v := range str {
		return string(unicode.ToUpper(v)) + str[i+1:]
	}
	return ""
}

//verify string is english, numeric or combination
func (u *Util) VerifyEnglishNumeric(input string) bool {
	if input == "" {
		return false
	}
	for _, v := range input {
		if !unicode.IsLetter(v) && !unicode.IsNumber(v) {
			return false
		}
	}
	return true
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
