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

func (f *Flag) GetFlagBool(nameTag string, c *cli.Context) bool {
	orgVal := f.GetFlagByName(nameTag, c)
	v, _ := orgVal.(bool)
	return v
}

func (f *Flag) GetFlagIntSlice(nameTag string, c *cli.Context) []int {
	orgVal := f.GetFlagByName(nameTag, c)
	v, _ := orgVal.([]int)
	return v
}

func (f *Flag) GetFlagInt(nameTag string, c *cli.Context) int {
	orgVal := f.GetFlagByName(nameTag, c)
	v, _ := orgVal.(int)
	return v
}

func (f *Flag) GetFlagStringSlice(nameTag string, c *cli.Context) []string {
	orgVal := f.GetFlagByName(nameTag, c)
	v, _ := orgVal.([]string)
	return v
}

func (f *Flag) GetFlagString(nameTag string, c *cli.Context) string {
	orgVal := f.GetFlagByName(nameTag, c)
	v, _ := orgVal.(string)
	return v
}

func (f *Flag) GetFlagByName(nameTag string, c *cli.Context) interface{}{
	return c.Value(nameTag)
}

//register new flag
func (f *Flag) RegisterNewFlag(nameTag string, kind int, usages ...string) error {
	//check
	if nameTag == "" || kind < FlagKindOfString {
		return errors.New("invalid parameter")
	}
	usage := ""
	if usages != nil && len(usages) > 0 {
		usage = usages[0]
	}

	//init by kind
	switch kind {
	case FlagKindOfInt:
		f.flags = append(f.flags, &cli.IntFlag{Name: nameTag, Usage: usage})
	case FlagKindOfBool:
		f.flags = append(f.flags, &cli.BoolFlag{Name: nameTag, Usage: usage})
	case FlagKindIntSlice:
		f.flags = append(f.flags, &cli.IntSliceFlag{Name: nameTag, Usage: usage})
	case FlagKindStringSlice:
		f.flags = append(f.flags, &cli.StringSliceFlag{Name: nameTag, Usage: usage})
	case FlagKindOfString:
		fallthrough
	default:
		f.flags = append(f.flags, &cli.StringFlag{Name: nameTag, Usage: usage})
	}
	return nil
}