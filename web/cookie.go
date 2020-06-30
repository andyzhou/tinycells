package web

import (
	"github.com/gorilla/securecookie"
	"github.com/kataras/iris"
	"time"
)

//internal macro defines
const (
	CookieHashKey = "04DgUStWPtd0aJmiR+L2z5xwbpPr/hmH" //"the-big-and-secret-fash-key-here"//
	CookieBlockKey = "5oR820YTUcjpBFAYlgbte8jEL9o2oHI2" //"lot-secret-of-characters-big-too"//
	CookieExpireSeconds = 86400 //xxx seconds
)

type Cookie struct {
	app *iris.Application
	secureCookie *securecookie.SecureCookie `secure cookie instance`
	expireTime int
}

//construct
func NewCookie() *Cookie {
	//self init
	this := &Cookie{
		expireTime:CookieExpireSeconds,
	}

	//inter init
	this.interInit()
	return this
}


///////
//api
//////

//set core data
func (f *Cookie) SetCoreData(app *iris.Application) bool {
	if app == nil {
		return false
	}
	f.app = app
	return true
}

//delete cookie
func (f *Cookie) DelCookie(key string, ctx iris.Context) bool {
	if key == "" {
		return false
	}
	ctx.RemoveCookie(key)
	return true
}

//get cookie
func (f *Cookie) GetCookie(key string, ctx iris.Context) string {
	if key == "" {
		return ""
	}
	return ctx.GetCookie(key, iris.CookieDecode(f.secureCookie.Decode))
}

//set cookie
func (f *Cookie) SetCookie(
					key, value string,
					totalSeconds int,
					ctx iris.Context,
				) bool {
	if key == "" || value == "" || totalSeconds <= 0 {
		return false
	}
	expireTime := time.Duration(totalSeconds) * time.Second
	ctx.SetCookieKV(
		key,
		value,
		iris.CookieEncode(f.secureCookie.Encode),
		iris.CookieExpires(expireTime),
	)
	return true
}


//////////////
//private func
//////////////

func (f *Cookie) interInit() {
	//init security key
	hashKey := []byte(CookieHashKey)
	blockKey := []byte(CookieBlockKey)

	//init cookie
	f.secureCookie = securecookie.New(hashKey, blockKey)
}
