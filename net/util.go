package net

import (
	"errors"
	"net"
)

/*
 * net util tool
 */

type Util struct {
}

//construct
func NewUtil() *Util {
	this := &Util{}
	return this
}

//get local ip
func (f *Util) GetLocalIp() ([]string, error) {
	addrSlice, err := net.InterfaceAddrs()
	if err != nil || addrSlice == nil {
		return nil, err
	}
	ips := make([]string, 0)
	for _, v := range addrSlice {
		if ipNet, ok := v.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				ips = append(ips, ipNet.IP.To4().String())
			}
		}
	}
	return ips, nil
}

//get out ip
func (f *Util) GetOutIp() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil || conn == nil {
		return "", err
	}
	defer conn.Close()
	addr, ok := conn.LocalAddr().(*net.UDPAddr)
	if !ok || addr == nil {
		return "", errors.New("can't get udp address")
	}
	return addr.IP.String(), nil
}