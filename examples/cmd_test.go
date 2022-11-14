package examples

import (
	"github.com/andyzhou/tinycells"
	"github.com/urfave/cli"
	"testing"
)

func startApp(c *cli.Context) error {
	c.String("xx")
	return nil
}

func CmdConfig(t *testing.T) {
	tc := tinycells.GetTC()
	cmd := tc.GetCmd()
	err := cmd.RegisterStringFlag("name")
	err = cmd.InitApp(startApp)
	cmd.StartApp()
	t.Logf("err:%v", err)
}
