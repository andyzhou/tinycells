package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"math"
	"net/url"
	"strings"
)

/*
 * gin web face
 */

//inter macro define
const (
	HttpProtocol = "://"
)

type Base struct {
}

//download data as file
func (f *Base) DownloadAsFile(downloadName string, data []byte, ctx gin.Context) error {
	//check
	if downloadName == "" || data == nil {
		return errors.New("invalid parameter")
	}

	//setup header
	ctx.Header("Content-type", "application/octet-stream")
	ctx.Header("Content-Disposition", "attachment; filename= " + downloadName)

	//write data into download file
	_, err := ctx.Writer.Write(data)
	return err
}

//calculate total pages
func (f *Base) CalTotalPages(total, size int) int {
	return int(math.Ceil(float64(total) / float64(size)))
}

//get json request body
func (f *Base) GetJsonRequest(c gin.Context, obj interface{}) error {
	//try read body
	jsonByte, err := f.GetRequestBody(c)
	if err != nil {
		return err
	}
	//try decode json data
	err = json.Unmarshal(jsonByte, obj)
	return err
}

//get request body
func (f *Base) GetRequestBody(c gin.Context) ([]byte, error) {
	return ioutil.ReadAll(c.Request.Body)
}

//get request para
func (f *Base) GetPara(name string, c *gin.Context) string {
	//decode request url
	//decodedReqUrl, _ := url.PathUnescape(c.Request.URL.RawQuery)
	values, _ := url.ParseQuery(c.Request.URL.RawQuery)

	//get act from url
	if values != nil {
		paraVal := values.Get(name)
		if paraVal != "" {
			return paraVal
		}
	}

	//get act from query, post.
	paraVal := c.Query(name)
	if paraVal == "" {
		//get from post
		paraVal = c.PostForm(name)
	}
	return paraVal
}

//get refer domain
func (f *Base) GetReferDomain(referUrl string) string {
	var (
		referDomain string
	)
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

//get request uri
func (f *Base) GetReqUri(ctx *gin.Context) string {
	var (
		reqUriFinal string
	)
	reqUri := ctx.Request.URL.RawQuery
	reqUriNew, err := url.QueryUnescape(reqUri)
	if err != nil {
		return reqUriFinal
	}
	reqUriFinal = reqUriNew
	return reqUriFinal
}

//get client ip
func (f *Base) GetClientIp(ctx *gin.Context) string {
	clientIp := ctx.Request.RemoteAddr
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