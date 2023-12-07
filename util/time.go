package util

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

//macro define
const (
	OneMinSec = 60
	OneHourSec = OneMinSec * 60
	OneDaySec = OneHourSec * 24
	OneMonthSec = OneDaySec * 30
	OneYearSec = OneMonthSec * 12
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

//convert time string format to int format
func (u *Util) TimeStr2Seconds(timeStr string) int {
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

//convert seconds to time string format
func (u *Util) Seconds2TimeStr(seconds int) string {
	var (
		hourStr, minuteStr, secondStr string
	)

	if seconds <= 0 {
		return ""
	}

	hourInt := seconds / 3600
	minuteInt := (seconds - hourInt * 3600) / 60
	secondInt := seconds - hourInt * 3600 - minuteInt * 60

	if hourInt > 0 {
		if hourInt > 9 {
			hourStr = fmt.Sprintf("%d:", hourInt)
		}else{
			hourStr = fmt.Sprintf("0%d:", hourInt)
		}
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
	timeStr := fmt.Sprintf("%s%s:%s", hourStr, minuteStr, secondStr)
	return timeStr
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

//convert timestamp like 'Oct 10, 2020' format
func (u *Util) TimeStampToDayStr(timeStamp int64, monthSizes ...int) string {
	var (
		monthSize int
	)
	date := u.TimeStamp2Date(timeStamp)
	if date == "" {
		return  ""
	}
	tempSlice := strings.Split(date, "-")
	if tempSlice == nil || len(tempSlice) < 3 {
		return ""
	}
	if monthSizes != nil && len(monthSizes) > 0 {
		monthSize = monthSizes[0]
	}

	//get key info
	year := tempSlice[0]
	month, _ := strconv.Atoi(tempSlice[1])
	day := tempSlice[2]

	//get assigned size month info
	monthInfo := time.Month(month).String()
	if monthSize > 0 && monthSize <= len(monthInfo) {
		monthInfo = monthInfo[:monthSize]
	}
	return fmt.Sprintf("%s %s, %s", monthInfo, day, year)
}

//convert diff seconds as string, like 'xx year | xx day | xx hours ago' format
func (u *Util) DiffTimeStampToStr(timeStamp int64) string {
	if timeStamp <= 0 {
		return ""
	}
	now := time.Now().Unix()
	diffSeconds := now - timeStamp
	if diffSeconds <= 0 {
		return ""
	}

	//calculate years
	years := int(math.Floor(float64(diffSeconds) / float64(OneYearSec)))
	months := int(math.Floor(float64(diffSeconds) / float64(OneMonthSec)))
	days := int(math.Floor(float64(diffSeconds) / float64(OneDaySec)))
	hours := int(math.Floor(float64(diffSeconds) / float64(OneHourSec)))
	minutes := int(math.Floor(float64(diffSeconds) / float64(OneMinSec)))
	if years > 0 {
		return fmt.Sprintf("%v years", years)
	}
	if months > 0 {
		return fmt.Sprintf("%v months", months)
	}
	if days > 0 {
		return fmt.Sprintf("%v days", days)
	}
	if hours > 0 {
		return fmt.Sprintf("%v hours", hours)
	}
	if minutes > 0 {
		return fmt.Sprintf("%v minutes", minutes)
	}
	return ""
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