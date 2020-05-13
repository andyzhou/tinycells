package tc

import (
	"net/http"
	"strings"
	"os"
	"net"
	"log"
	"io"
	"time"
	"math/rand"
	"crypto/md5"
	"encoding/hex"
	"unicode"
	"strconv"
	"fmt"
	"io/ioutil"
	"reflect"
	"math"
	"bytes"
)

/*
 * public util functions
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

//internal macro defines
const (
	UseLostConnectionErr = "use of closed network connection"
	TimeLayOut = "2006-01-02 15:04:05" //can't be changed!!!
)

//util info
type Utils struct {
}

//shuffle slice
func (u *Utils) ShuffleSlice(data []int) {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(data), func(i, j int) { data[i], data[j] = data[j], data[i] })
}

//create json_object sql pass json data map
func (u *Utils) GenJsonObject(
					genHashMap map[string]interface{},
				) (string, []interface{}) {
	var (
		buffer = bytes.NewBuffer(nil)
		tempStr string
		values = make([]interface{}, 0)
	)

	//basic check
	if genHashMap == nil || len(genHashMap) <= 0 {
		return buffer.String(), values
	}

	//convert into relate data
	i := 0
	buffer.WriteString("json_object(")
	for k, v := range genHashMap {
		if i > 0 {
			buffer.WriteString(" ,")
		}
		tempStr = fmt.Sprintf("'%s', ?", k)
		values = append(values, v)
		buffer.WriteString(tempStr)
		i++
	}
	buffer.WriteString(")")

	return buffer.String(), values
}

//check or create dir
func (u *Utils) CheckOrCreateDir(dir string) error {
	_, err := os.Stat(dir)
	if err == nil {
		return err
	}
	bRet := os.IsExist(err)
	if bRet {
		return nil
	}
	err = os.Mkdir(dir, 0777)
	return err
}

//round float value
func (u *Utils) Round(f float64, n int) float64 {
	n10 := math.Pow10(n)
	return math.Trunc((f+0.5/n10)*n10) / n10
}

//reset object instance
func (u *Utils) RestObject(v interface{}) {
	p := reflect.ValueOf(v).Elem()
	p.Set(reflect.Zero(p.Type()))
}

//read byte file
//return bool and byte data
func (u *Utils) ReadBinFile(filePath string, needRemove bool) (bool, []byte) {
	if filePath == "" {
		return false, nil
	}

	//try read file
	byteData, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Println("ImageResize::ReadImage failed, err:", err.Error())
		return false, nil
	}

	if needRemove {
		os.Remove(filePath)
	}

	return true, byteData
}

//convert seconds to time string format
func (u *Utils) Seconds2TimeStr(seconds int) string {
	var (
		hourStr, minuteStr, secondStr string
	)

	if seconds <= 0 {
		return ""
	}

	hourInt := seconds / 3600
	minuteInt := (seconds - hourInt * 3600) / 60
	secondInt := seconds - hourInt * 3600 - minuteInt * 60

	if hourInt > 9 {
		hourStr = fmt.Sprintf("%d", hourInt)
	}else{
		hourStr = fmt.Sprintf("0%d", hourInt)
	}

	if minuteInt > 9 {
		minuteStr = fmt.Sprintf("%d", minuteInt)
	}else{
		minuteStr = fmt.Sprintf("0%d", minuteInt)
	}

	if secondInt > 9 {
		secondStr = fmt.Sprintf("%d", secondInt)
	}else{
		secondStr = fmt.Sprintf("0%d", secondInt)
	}

	//format time string
	timeStr := fmt.Sprintf("%s:%s:%s", hourStr, minuteStr, secondStr)
	return timeStr
}

//convert time string format to int format
func (u *Utils) TimeStr2Seconds(timeStr string) int {
	var (
		timeSeconds int
		tempIntVal int
		tempFloatVal float64
	)

	//basic check
	if timeStr == "" {
		return timeSeconds
	}

	i := 1
	tempSlice := strings.Split(timeStr, ":")
	for _, info := range tempSlice {
		switch i {
		case 1://hour
			tempIntVal, _ = strconv.Atoi(info)
			timeSeconds += tempIntVal * 3600
		case 2://minute
			tempIntVal, _ = strconv.Atoi(info)
			timeSeconds += tempIntVal * 60
		case 3:	//second
			tempFloatVal, _ = strconv.ParseFloat(info, 64)
			timeSeconds += int(tempFloatVal)
		}
		i++
	}

	return timeSeconds
}


//upper first character
func (u *Utils) UcFirst(str string) string {
	for i, v := range str {
		return string(unicode.ToUpper(v)) + str[i+1:]
	}
	return ""
}

//lower first character
func (u *Utils) LcFirst(str string) string {
	for i, v := range str {
		return string(unicode.ToLower(v)) + str[i+1:]
	}
	return ""
}

//sub string
func (u *Utils) SubString(source string, start int, end int) string {
	var substring = ""
	var pos = 0
	for _, c := range source {
		if pos < start {
			pos++
			continue
		}
		if pos >= end {
			break
		}
		pos++
		substring += string(c)
	}
	return substring
}

//generate md5 string
func (u *Utils) GenMd5(orgString string) string {
	if len(orgString) <= 0 {
		return ""
	}
	m := md5.New()
	m.Write([]byte(orgString))
	return hex.EncodeToString(m.Sum(nil))
}

//get rand number
func (u *Utils) GetRandomVal(maxVal int) int {
	randSand := rand.NewSource(time.Now().UnixNano())
	r := rand.New(randSand)
	return r.Intn(maxVal)
}

func (u *Utils) GetRealRandomVal(maxVal int) int {
	return int(rand.Float64() * 1000) % (maxVal)
}

//convert date time string to timestamp
func (u *Utils) DateTime2Unix(dateTime string) int64 {
	//remove un useful info
	dateTime = strings.ReplaceAll(dateTime, "T", " ")
	dateTime = strings.ReplaceAll(dateTime, "Z", "")

	//theTime, err := time.Parse(TimeLayOut, dateTime)
	theTime, err := time.ParseInLocation(TimeLayOut, dateTime, time.Local)
	if err != nil {
		log.Println("TimeToUnixStamp, convert failed, err:", err.Error())
		return 0
	}
	return theTime.Unix()
}

//convert timestamp to date format, like YYYY-MM-DD
func (u *Utils) TimeStamp2Date(timeStamp int64) string {
	dateTime := time.Unix(timeStamp, 0).Format(TimeLayOut)
	tempSlice := strings.Split(dateTime, " ")
	if tempSlice == nil || len(tempSlice) <= 0 {
		return ""
	}
	return tempSlice[0]
}

//convert timestamp to data time string format
func (u *Utils) TimeStamp2DateTime(timeStamp int64) string {
	return time.Unix(timeStamp, 0).Format(TimeLayOut)
}

//check tcp error
//return true need quit, false just gen error
func (u *Utils) CheckTcpError(err error) bool {
	var (
		isOk bool
		netOpError *net.OpError
	)
	if err != nil {
		log.Println("service::checkTcpError, err:", err)
		netOpError, isOk = err.(*net.OpError)
		if isOk && netOpError.Err.Error() == UseLostConnectionErr {
			//use a broken connect
			return true
		}
		if err == io.EOF {
			//ignore EOF since client might send nothing for the moment
			return true
		}
		netErr, ok := err.(net.Error)
		if ok && netErr.Timeout() {
			//socket operate time out
			return true
		}
	}
	return false
}

//get current host
func (u *Utils) GetCurHost() string {
	//get local ip
	var defaultHost = "127.0.0.1"
	var ipAddress string

	addressSlice, err := net.InterfaceAddrs()
	if nil != err {
		log.Fatal("Get local IP addr failed!!!")
		return defaultHost
	}
	for _, address := range addressSlice {
		if ipNet, ok := address.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if nil != ipNet.IP.To4() {
				ipAddress = ipNet.IP.String()
				return ipAddress
			}
		}
	}
	return defaultHost
}

//add parameter into slice
func (u *Utils) AddParamInSlice(orgSlice *[]string, para string) bool {
	if para == "" {
		return false
	}
	*orgSlice = append(*orgSlice, para)
	return true
}

//convert slice to string
func (u *Utils) Slice2Str(orgSlice []string) string {
	var result string
	if len(orgSlice) <= 0 {
		return result
	}
	for _, v := range orgSlice {
		result += v
	}
	return result
}

//check file stat and last modify time
func (u *Utils) GetFileModifyTime(filePath string) int64 {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return 0
	}
	modifyTime := fileInfo.ModTime().Unix()
	return modifyTime
}

//get file info
func (u *Utils) GetFileInfo(filePath string) os.FileInfo {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil
	}
	return fileInfo
}

//gather all ip from client
func (u *Utils) GetClientAllIp(r *http.Request) []string {
	var tempStr string
	var ipSlice = make([]string, 0)

	//get original data
	clientAddress := r.RemoteAddr
	xRealIp := r.Header.Get("X-Real-IP")
	xForwardedFor := r.Header.Get("X-Forwarded-For")

	//analyze general ip
	if clientAddress != "" {
		tempStr = u.analyzeClientIp(clientAddress)
		if tempStr != "" {
			ipSlice = append(ipSlice, tempStr)
		}
	}

	//analyze x-real-ip
	if xRealIp != "" {
		tempStr = u.analyzeClientIp(clientAddress)
		if tempStr != "" {
			ipSlice = append(ipSlice, tempStr)
		}
	}

	//analyze x-forward-for
	//like:192.168.0.1,192.168.0.2
	if xForwardedFor != "" {
		tempSlice := strings.Split(xForwardedFor, ",")
		if len(tempSlice) > 0 {
			for _, tmpAddr := range tempSlice {
				tempStr = u.analyzeClientIp(tmpAddr)
				if tempStr != "" {
					ipSlice = append(ipSlice, tempStr)
				}
			}
		}
	}

	return ipSlice
}

//analyze client ip
func (u *Utils) analyzeClientIp(address string) string {
	tempSlice := strings.Split(address, ":")
	if len(tempSlice) < 1 {
		return ""
	}
	return tempSlice[0]
}
