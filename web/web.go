package web

import (
	"github.com/kataras/iris"
	"net/url"
	"strings"
	"fmt"
)

/*
 * public api for iris web
 */

 //inter macro define
 const (
 	HttpProtocol = "://"
 )

 type BaseWeb struct {
 }

//get refer domain
func (w *BaseWeb) GetReferDomain(ctx iris.Context) string {
	var (
		referDomain string
	)

	referUrl := ctx.Request().Referer()
	if referUrl == "" {
		return referDomain
	}

	//find first '://' pos
	protocolLen := len(HttpProtocol)
	protocolPos := strings.Index(referUrl, HttpProtocol)
	if protocolPos <= -1 {
		return referDomain
	}

	//pick domain
	tempBytes := []byte(referUrl)
	tempBytesLen := len(tempBytes)
	prefixLen := protocolPos + protocolLen
	resetUrl := tempBytes[prefixLen:tempBytesLen]
	tempSlice := strings.Split(string(resetUrl), "/")
	if tempSlice == nil || len(tempSlice) <= 0 {
		return referDomain
	}
	referDomain = fmt.Sprintf("%s%s", tempBytes[0:prefixLen], tempSlice[0])
	return referDomain
}

//get general parameter
func (w *BaseWeb) GetParameter(
					name string,
					form url.Values,
					ctx iris.Context,
				) string {
	value := form.Get(name)
	if value == "" {
		value = ctx.Params().Get(name)
	}
	return value
}

//get all values of one parameter
func (w *BaseWeb) GetParameterValues(
					name string,
					form url.Values,
					ctx iris.Context,
				) []string {
	if form == nil {
		return nil
	}
	vs := form[name]
	if len(vs) == 0 {
		return nil
	}
	return vs
}

//get http request parameters
func (w *BaseWeb) GetHttpParameters(ctx iris.Context) url.Values {
	var err error
	//get request uri
	reqUri := w.GetReqUri(ctx)
	if reqUri == "" {
		return nil
	}
	//parse form
	queryForm, err := url.ParseQuery(reqUri)
	if err != nil {
		return nil
	}
	return queryForm
}

//get request uri
func (w *BaseWeb) GetReqUri(ctx iris.Context) string {
	var (
		reqUriFinal string
	)

	reqUri := ctx.Request().URL.RawQuery
	reqUriNew, err := url.QueryUnescape(reqUri)
	if err != nil {
		return reqUriFinal
	}
	reqUriFinal = reqUriNew
	return reqUriFinal
}

//get client ip
func (w *BaseWeb) GetClientIp(ctx iris.Context) string {
	clientIp := ctx.RemoteAddr()
	xRealIp := ctx.GetHeader("X-Real-IP")
	xForwardedFor := ctx.GetHeader("X-Forwarded-For")
	if clientIp != "" {
		return clientIp
	}else{
		if xRealIp != "" {
			clientIp = xRealIp
		}else{
			clientIp = xForwardedFor
		}
	}
	return clientIp
}

