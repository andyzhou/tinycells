package util

import (
	"errors"
	"strings"
	"sync"
)

/*
 * DFA algorithm for words filter
 * - used for filter bad words
 * - replace matched word as `*`
 */

const (
	InvalidWords = " ,~,!,@,#,$,%,^,&,*,(,),_,-,+,=,?,<,>,.,—,，,。,/,\\,|,《,》,？,;,:,：,',‘,；,“,"
)

//face info
type DFA struct {
	sensitiveWord map[string]interface{}
	set map[string]interface{}
	invalidWords []string
	invalidWord map[string]interface{}
	sync.RWMutex
}

//construct
func NewDFA() *DFA {
	this := &DFA{
		sensitiveWord: map[string]interface{}{},
		set: map[string]interface{}{},
		invalidWords: []string{},
		invalidWord: map[string]interface{}{},
	}
	this.interInit()
	return this
}

//change words
func (f *DFA) ChangeSensitiveWords(txt string) (bool, string){
	str := []rune(txt)
	nowMap := f.sensitiveWord
	start := -1
	tag := -1
	hasFound := false
	for i := 0; i < len(str); i++ {
		if _, ok:= f.invalidWord[(string(str[i]))]; ok {
			//if it is invalid word, skip it.
			continue
		}
		if thisMap, ok :=nowMap[string(str[i])].(map[string]interface{}); ok {
			//record first char of word
			tag++
			if  tag == 0 {
				start = i

			}
			//check it is last char of word
			if isEnd, _ := thisMap["isEnd"].(bool);isEnd {
				//replace first to last char as `*`
				for y := start; y < i+1; y++ {
					str[y] = 42
					hasFound = true
				}
				//reset
				nowMap = f.sensitiveWord
				start = -1
				tag = -1

			}else{
				//not last, replace map to nowMap
				nowMap = nowMap[string(str[i])].(map[string]interface{})
			}
		}else{
			//if not full match, end found.
			//check from second char
			if start != -1 {
				i = start + 1
			}
			//reset
			nowMap = f.sensitiveWord
			start = -1
			tag = -1
		}
	}
	return hasFound, string(str)
}

//reset filter words
func (f *DFA) ResetFilterWords() {
	f.Lock()
	defer f.Unlock()
	f.sensitiveWord = map[string]interface{}{}
}

//add filter words
func (f *DFA) AddFilterWords(words ...string) error {
	//check
	if words == nil || len(words) <= 0 {
		return errors.New("invalid parameter")
	}
	//add on by one
	f.Lock()
	defer f.Unlock()
	for _, v := range words {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		f.set[v] = nil
	}
	//sync sensitive map
	f.addSensitiveToMap(f.set)
	return nil
}

//add invalid words
func (f *DFA) AddInvalidWords(words ...string) error {
	//check
	if words == nil || len(words) <= 0 {
		return errors.New("invalid parameter")
	}
	//add on by one
	f.Lock()
	defer f.Unlock()
	for _, v := range words {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		f.invalidWord[v] = nil
	}
	return nil
}

/////////////////////
//private func
/////////////////////

//add sensitive words
func (f *DFA) addSensitiveToMap(set map[string]interface{}){
	for key := range set {
		str := []rune(key)
		nowMap := f.sensitiveWord
		for i := 0; i < len(str); i++ {
			if _,ok := nowMap[string(str[i])]; !ok {
				//if key not exists，
				thisMap := make(map[string]interface{})
				thisMap["isEnd"] = false
				nowMap[string(str[i])] = thisMap
				nowMap = thisMap
			}else {
				nowMap = nowMap[string(str[i])].(map[string]interface{})
			}
			if i == len(str)-1 {
				nowMap["isEnd"] = true
			}
		}
	}
}

//load default invalid word
func (f *DFA) loadDefaultInvalidWord() {
	//split words
	words := strings.Split(InvalidWords,",")
	if words == nil || len(words) <= 0 {
		return
	}
	//load into invalid word
	for _, v := range words {
		f.invalidWord[v] = nil
	}
}

//inter init
func (f *DFA) interInit() {
	f.loadDefaultInvalidWord()
}