package tc

import (
	"crypto/md5"
	"fmt"
	"strconv"
)

/*
 * Short url interface
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 * base on "github.com/kaimixu/shorturl"
 */

//internal macro define
const (
	CharsetAlphanumeric = iota
	CharsetRandomAlphanumeric
)

//short url info
type ShortUrl struct {
}
 
 //construct
func NewShortUrl() *ShortUrl {
	this := &ShortUrl{
	}
	return this
}


// 生成6-8字符的短链接，参数t表示字符集类型，回调函数(cb)用于检测短链接是否重复
// 起初生成6位的短链接，当四组6位短链接都重复时，再生成8位的短链接
func (s *ShortUrl) Generator(t int, url string, cb func(url, keyword string) bool) string {
	var (
		shortUrl string
	)

	if url == "" || cb == nil {
		return shortUrl
	}

	//get charset
	charset := s.getCharset(t)
	hexMd5 := fmt.Sprintf("%x", md5.Sum([]byte(url)))
	sections := len(hexMd5)/8

	//generate keyword
	keyword := s.generator6(charset, url, hexMd5, sections, cb)
	if keyword == "" {
		keyword = s.generator8(charset, url, hexMd5, sections, cb)
		if keyword == "" {
			return ""
		}
	}

	//if strings.HasSuffix(domain, "/") {
	//	shorturl = domain + keyword
	//}else {
	//	shorturl = domain + "/" + keyword
	//}

	return keyword
}

//////////////
//private func
//////////////


// 生成6字符短key
// 根据url生成32字符的签名，将其分成4段，每段8位字符
// 循环处理4段8位字符，将每段转换成16进制与0x3FFFFFFF进行逻辑与操作，得到30位的无符号数
// 将30位数分成6段，依次得到6个0-61的数字索引查字符集表获得6位字符串
func (s *ShortUrl) generator6(charset ,url, hexMd5 string,  sectionNum int, cb func(url, keyword string) bool) string {
	for i := 0; i < sectionNum; i++ {
		sectionHex := hexMd5[i*8:8+i*8]
		bits, _ := strconv.ParseUint(sectionHex, 16, 32)
		bits = bits & 0x3FFFFFFF

		keyword := ""
		for j := 0; j < 6; j++ {
			idx := bits & 0x3D
			keyword = keyword + string(charset[idx])
			bits = bits >> 5
		}

		if cb(url, keyword) {
			return keyword
		}
	}

	return ""
}

// 生成8字符短key
func (s *ShortUrl) generator8(charset, url, hexMd5 string,  sectionNum int, cb func(url, keyword string) bool) string {
	for i := 0; i < sectionNum; i++ {
		sectionHex := hexMd5[i*8:i*8+8]
		bits, _ := strconv.ParseUint(sectionHex, 16, 32)
		bits = bits & 0xFFFFFFFF

		keyword := ""
		for j := 0; j < 8; j++ {
			idx := bits & 0x3D
			keyword = keyword + string(charset[idx])
			bits = bits >> 4
		}

		if cb(url, keyword) {
			return keyword
		}
	}

	return ""
}

func (s *ShortUrl) getCharset(t int) string {
	switch t {
	case CharsetAlphanumeric:
		return "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	case CharsetRandomAlphanumeric:
		return "A0a12B3b4CDc56Ede7FGf8Hg9IhJKiLjkMNlmOPnQRopqrSstTuvUVwxWXyYzZ"
	default:
		//panic("invalid charset type t:" + strconv.Itoa(t))
		return ""
	}
}