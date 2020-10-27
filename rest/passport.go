package rest

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"github.com/andyzhou/tinycells/tc"
	"io"
	"log"
)

/*
 * passport face for xi'an
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 * used for load user basic info pass user id and token
 */

//passport face
type PassportFace struct {
	tc.BaseJson
}

//construct
func NewPassportFace() *PassportFace {
	this := &PassportFace{
	}
	return this
}

/////////
//api
/////////

//un compress data
func (f *PassportFace) UnCompressData(resp []byte) (bool, map[string]interface{}) {
	var (
		outBytes bytes.Buffer
	)

	//base64 decode
	respStr := string(resp)
	byteData, err := base64.StdEncoding.DecodeString(respStr)
	if err != nil {
		return false, nil
	}

	//zip un compress
	byteReader := bytes.NewReader(byteData)
	gzipReader, err := gzip.NewReader(byteReader)
	if err != nil {
		return false, nil
	}

	_, err = io.Copy(&outBytes, gzipReader)
	if err != nil {
		return false, nil
	}

	//json decode
	result := make(map[string]interface{})
	err = json.Unmarshal(outBytes.Bytes(), &result)
	if err != nil {
		return false, nil
	}

	return true, result
}

//compress data
func (f *PassportFace) CompressData(req map[string]interface{}) (bool, []byte) {
	var (
		in bytes.Buffer
	)
	//basic check
	if req == nil || len(req) <= 0 {
		return false, nil
	}

	//json encode
	jsonByte, err := json.Marshal(req)
	if err != nil {
		return false, nil
	}

	//zip compress
	zipWriter := gzip.NewWriter(&in)
	_, err = zipWriter.Write(jsonByte)
	if err != nil {
		zipWriter.Close()
		return false, nil
	}
	err = zipWriter.Close()
	if err != nil {
		log.Println("err:", err.Error())
	}

	//base64 encode
	out := make([]byte, 2048)
	base64.StdEncoding.Encode(out, in.Bytes())

	return true, out
}
