package util

import (
	"fmt"
	"strings"
	"time"
)

//begin of time period
func (u *Util) BeginningOfDay() time.Time {
	y, m, d := u.Now().Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func (u *Util) BeginningOfWeek() time.Time {
	t := u.BeginningOfDay()
	weekday := int(t.Weekday())
	weekStartDay := int(time.Monday)

	if weekday < weekStartDay {
		weekday = weekday + 7 - weekStartDay
	} else {
		weekday = weekday - weekStartDay
	}
	return t.AddDate(0, 0, -weekday)
}

func (u *Util) BeginningOfMonth() time.Time {
	y, m, _ := u.Now().Date()
	return time.Date(y, m, 1, 0, 0, 0, 0, time.UTC)
}

//convert string format time to unix
func (u *Util) ConvertStrTime2Unix(timeStr string) (int64, error) {
	nowT := time.Now()
	tm, err := time.Parse(TimeLayoutStr, timeStr)
	if err != nil {
		return 0, err
	}
	tm.Sub(nowT)
	return tm.UTC().Unix(), nil
}

//convert timestamp to date format
func (u *Util) TimeStamp2DateTime(timeStamp int64) string {
	return time.Unix(timeStamp, 0).UTC().Format(TimeLayoutStr)
}

//convert date time string to timestamp
func (u *Util) DateTime2Unix(dateTime string) (int64, error) {
	//remove un useful info
	dateTime = strings.Replace(dateTime, "T", " ", -1)
	dateTime = strings.Replace(dateTime, "Z", "", -1)

	//theTime, err := time.Parse(TimeLayOut, dateTime)
	theTime, err := time.ParseInLocation(TimeLayoutStr, dateTime, time.Local)
	if err != nil {
		return 0, err
	}
	return theTime.Unix(), nil
}

//convert timestamp to date format, like YYYY-MM-DD
func (u *Util) TimeStamp2Date(timeStamp int64) string {
	dateTime := time.Unix(timeStamp, 0).Format(TimeLayoutStr)
	tempSlice := strings.Split(dateTime, " ")
	if tempSlice == nil || len(tempSlice) <= 0 {
		return ""
	}
	return tempSlice[0]
}

//get current date, like YYYY-MM-DD
func (u *Util) GetCurDate() string {
	now := time.Now()
	curDate := fmt.Sprintf("%d-%d-%d", now.Year(), now.Month(), now.Day())
	return curDate
}

//sec to ms second
func (u *Util) Sec2MSec(s int64) int64 {
	return s * 1000
}
func (u *Util) MSec2Sec(ms int64) int64 {
	return ms / 1000
}

//get utc second
func (u *Util) UNIX() int64 {
	return u.Now().Unix()
}

func (u *Util) UNIX_MS() int64 {
	return u.Now().UnixNano() / 1e6
}

//get current utc time
func (u *Util) Now() time.Time {
	if controlDuration != 0 {
		return time.Now().Add(controlDuration).UTC()
	}
	return time.Now().UTC()
}

//reset server control duration
func (u *Util) ResetControlDuration() {
	controlDuration = 0
}

//set server control duration
func (u *Util) SetControlDuration(duration time.Duration) {
	controlDuration = duration
}

//change server control duration
func (u *Util) ChangeControlDuration(timeStr string) (time.Duration, error) {
	nowT := time.Now()
	changeTime, err := time.Parse(TimeLayoutStr, timeStr)
	if err != nil {
		return 0, err
	}
	controlDuration = changeTime.Sub(nowT)
	return controlDuration, nil
}