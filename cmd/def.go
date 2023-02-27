package cmd

import "github.com/urfave/cli/v2"

type (
	StartFunc func(c *cli.Context) error
)
