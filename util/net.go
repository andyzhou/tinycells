package util

import (
	"io"
	"log"
	"net"
	"net/http"
	"strings"
)

//gather all ip from client
func (u *Util) GetClientAllIp(r *http.Request) []string {
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

//get current host
func (u *Util) GetCurHost() string {
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

//check tcp error
//return true need quit, false just gen error
func (u *Util) CheckTcpError(err error) bool {
	var (
		isOk bool
		netOpError *net.OpError
	)
	if err != nil {
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

//analyze client ip
func (u *Util) analyzeClientIp(address string) string {
	tempSlice := strings.Split(address, ":")
	if len(tempSlice) < 1 {
		return ""
	}
	return tempSlice[0]
}
