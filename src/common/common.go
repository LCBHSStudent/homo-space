package common

import (
	"crypto/md5"
	"encoding/hex"
	"strings"
	"time"
)

type HomoInfo struct {
	ID              int64
	Name            string
	Description     string
	Level           int
	HP              int
	ATN             int
	INT             int
	DEF             int
	RES             int
	SPD             int
	LUK             int
	Skills          [6]int
	Status          int
}

var SkillList = [1]string{"Unknown"}

func remove(slice []interface{}, elem interface{}) []interface{} {
	if len(slice) == 0 {
		return slice
	}
	for i, v := range slice {
		if v == elem {
			slice = append(slice[:i], slice[i+1:]...)
			return remove(slice,elem)
		}
	}
	return slice
}

var HashSalt string = "28c1fdd170a5204386cb1313c7077b34f83e4aaf4aa829ce78c231e05b0bae2c"

func GetCDK() string {
	var result  string
	var curTime string = getCurFormatTime()
	
	data        := []byte(curTime+HashSalt)
	encodedData := md5.Sum(data)
	
	for _, b := range encodedData {
		temp    := make([]byte, 1)
		temp[0] = byte(int(b) & 0xff)
		hexString := hex.EncodeToString(temp)
		if len(hexString) < 2 {
			hexString = "0" + hexString
		}
		result = result+hexString
	}
	
	return result
}

func getCurFormatTime() string {
	rawTime := time.Now().Format("2006-01-02 15:04:05")
	var curTime string
	
	for index, timeMsg := range strings.Split(rawTime, " ") {
		if index == 0 {
			curTime = timeMsg
		} else if index == 1 {
			curTime = curTime + "T" + timeMsg + "+08:00"
		}
	}
	return curTime
}

func JudgeWeather2LvUp(exp int, level int) bool {
	return true
}

func merge() {

}

func MergeSort(slice []int64, cmp func(int64, int64)) {

}