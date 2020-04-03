package rest

import (
	"github.com/andyzhou/tinycells/tc"
	"sync"
	"bytes"
	"strings"
	"sort"
)

/*
 * Sign interface
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 * used for generate signature
 */

//sort field info
type SortField struct {
	Field string
	Value string
}

//sign conf
type SignConf struct {
	Switcher bool
	SignKey string
	SkipReqPara []string
}

//sign interface info
type SignFace struct {
	signConf *SignConf
	skipFields map[string]bool
	tc.Utils
	sync.RWMutex
}

//construct
func NewSignFace() *SignFace {
	//self init
	this := &SignFace{
		signConf:new(SignConf),
		skipFields:make(map[string]bool),
	}
	return this
}

func NewSignConf() *SignConf {
	this := &SignConf{
		SkipReqPara:make([]string, 0),
	}
	return this
}

///////
//api
//////

//generate sign
func (s *SignFace) GenSign(fields map[string]string) (bool, string) {
	var (
		sign string
	)

	//basic check
	if fields == nil || len(fields) <= 0 {
		return false, sign
	}

	//create new sorter
	sorter := s.createSorter(fields)
	if sorter == nil {
		return false, sign
	}

	//generate final sign
	sign = s.generateSign(sorter)

	return true, sign
}

//add skip fields
func (s *SignFace) AddSkipFields(fields []string) bool {
	if fields == nil || len(fields) <= 0 {
		return false
	}
	s.Lock()
	defer s.Unlock()
	for _, field := range fields {
		s.skipFields[field] = true
	}
	return true
}

//set sign conf
func (s *SignFace) SetConf(conf *SignConf) bool {
	if conf == nil {
		return false
	}
	s.signConf = conf
	s.AddSkipFields(conf.SkipReqPara)
	return true
}

//get switcher
func (s *SignFace) GetSwitcher() bool {
	return s.signConf.Switcher
}

///////////////
//private func
///////////////

//generate final sign
func (s *SignFace) generateSign(sorter SortFields) string {
	var (
		signVal string
		byteBuff = bytes.NewBuffer(nil)
	)

	//basic check
	if sorter == nil {
		return signVal
	}

	//gen final string
	for _, v := range sorter {
		byteBuff.WriteString(v.Value)
	}
	byteBuff.WriteString(s.signConf.SignKey)

	//md5 value
	signVal = s.GenMd5(byteBuff.String())

	//convert to lower format
	return strings.ToLower(signVal)
}

//create sorter
func (s *SignFace) createSorter(fields map[string]string) SortFields {
	var (
		sorter SortFields
		isOk bool
	)

	//basic check
	if fields == nil || len(fields) <= 0 {
		return nil
	}

	//add elements
	for k, v := range fields {
		//check is skip fields
		_, isOk = s.skipFields[k]
		if isOk {
			continue
		}
		element := &SortField{
			Field:k,
			Value:v,
		}
		sorter = append(sorter, element)
	}

	//sort elements
	sort.Sort(sorter)

	return sorter
}

///////////////////////
//interface for sort
//DO NOT CHANGE THIS!
//////////////////////

//sort fields slice
type SortFields []*SortField

//Len()
func (s SortFields) Len() int {
	return len(s)
}

//sort by id ASC
func (s SortFields) Less(i, j int) bool {
	return s[i].Field < s[j].Field
}

//Swap()
func (s SortFields) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}


