package cmd

import "github.com/urfave/cli"

type (
	StartFunc func(c *cli.Context) error
)
