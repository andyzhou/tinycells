package cmd

import (
	"errors"
	"github.com/urfave/cli"
)

//face info
type Flag struct {
	flags []cli.Flag
}

//construct
func NewFlag() *Flag {
	this := &Flag{
		flags: []cli.Flag{},
	}
	return this
}

//get flags
func (f *Flag) GetFlags() []cli.Flag {
	return f.flags
}

//register new flag
func (f *Flag) RegisterNewFlag(name string, kind int, usages ...string) error {
	//check
	if name == "" || kind < FlagKindOfString {
		return errors.New("invalid parameter")
	}
	usage := ""
	if usages != nil && len(usages) > 0 {
		usage = usages[0]
	}

	//init by kind
	switch kind {
	case FlagKindOfInt:
		f.flags = append(f.flags, &cli.IntFlag{Name: name, Usage: usage})
	case FlagKindOfBool:
		f.flags = append(f.flags, &cli.BoolFlag{Name: name, Usage: usage})
	case FlagKindOfString:
		fallthrough
	default:
		f.flags = append(f.flags, &cli.StringFlag{Name: name, Usage: usage})
	}
	return nil
}