package examples

import (
	"github.com/andyzhou/tinycells"
	"testing"
)

func TestConfig(t *testing.T) {
	tc := tinyecells.GetTC()
	tc.SetConfig()
	err := tc.GetConfig().GetIniConf().LoadConfig("test.cfg")
	t.Logf("load config result err:%v", err)
}
