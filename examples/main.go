package main

import (
	"github.com/andyzhou/tinycells"
	"log"
)

func main() {
	testLogger()
}

//test logger
func testLogger()  {
	tc := tinycells.GetTC()
	logger := tc.GetLogger()
	config := logger.BuildDefaultConfig()
	err := logger.SetConfig(config)
	logger.SS().Infof("test logger")
	log.Printf("err:%v", err)
}