package main

import (
	"github.com/andyzhou/tinycells"
	"testing"
)

func TestLogger(t *testing.T) {
	tc := tinycells.GetTC()
	logger := tc.GetLogger()
	config := logger.BuildDefaultConfig()
	err := logger.SetConfig(config)
	logger.SS().Infof("test logger")
	t.Logf("load config result err:%v", err)
}
