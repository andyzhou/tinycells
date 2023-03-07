package cmd

import (
	"errors"
	"github.com/urfave/cli/v2"
	"os"
)

//face info
type Cmd struct {
	app *cli.App
	flag *Flag
	isRunning bool
}

//construct
func NewCmd() *Cmd {
	this := &Cmd{
		flag: NewFlag(),
	}
	return this
}

//get flag
func (f *Cmd) GetFlag() *Flag {
	return f.flag
}

//start app, step-3
func (f *Cmd) StartApp() error {
	//check
	if f.app == nil {
		return errors.New("app hadn't init")
	}
	if f.isRunning {
		return errors.New("app is running")
	}
	//start app
	err := f.app.Run(os.Args)
	if err != nil {
		return err
	}
	f.isRunning = true
	return nil
}

//init app, step-2
func (f *Cmd) InitApp(sf StartFunc, appNames ...string) error {
	//check cache
	if f.app != nil {
		return errors.New("app had init")
	}
	//init new
	appName := DefaultAppName
	if appNames != nil && len(appNames) > 0 {
		appName = appNames[0]
	}
	app := &cli.App{
		Name:  appName,
		Action: func(c *cli.Context) error {
			return sf(c)
		},
		Flags: f.flag.GetFlags(),
	}
	f.app = app
	return nil
}

//register new flag, step-1
func (f *Cmd) RegisterBoolFlag(nameTag string, usages ...string) error {
	return f.RegisterNewFlag(nameTag, FlagKindOfBool, usages...)
}
func (f *Cmd) RegisterIntFlag(nameTag string, usages ...string) error {
	return f.RegisterNewFlag(nameTag, FlagKindOfInt, usages...)
}
func (f *Cmd) RegisterStringFlag(nameTag string, usages ...string) error {
	return f.RegisterNewFlag(nameTag, FlagKindOfString, usages...)
}
func (f *Cmd) RegisterNewFlag(nameTag string, kind int, usages ...string) error {
	return f.flag.RegisterNewFlag(nameTag, kind, usages...)
}

