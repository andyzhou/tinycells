package util

import (
	"errors"
	"io/ioutil"
	"os"
)

//check file stat and last modify time
func (u *Util) GetFileModifyTime(filePath string) int64 {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return 0
	}
	modifyTime := fileInfo.ModTime().Unix()
	return modifyTime
}

//get file info
func (u *Util) GetFileInfo(filePath string) os.FileInfo {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil
	}
	return fileInfo
}

//read byte file
func (u *Util) ReadBinFile(filePath string, needRemoves ...bool) ([]byte, error) {
	//check
	if filePath == "" {
		return nil, errors.New("invalid file path")
	}
	//try read file
	byteData, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	needRemove := false
	if needRemoves != nil && len(needRemoves) > 0 {
		needRemove = needRemoves[0]
		if needRemove {
			os.Remove(filePath)
		}
	}
	return byteData, nil
}

//check or create dir
func (u *Util) CheckOrCreateDir(dir string) error {
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
