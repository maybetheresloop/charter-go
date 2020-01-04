package main

import (
	"os"

	"github.com/BurntSushi/toml"
	"github.com/maybetheresloop/charter-go"
	"github.com/urfave/cli"
)

const (
	DefaultAddr          string = ":5678"
	DefaultDir           string = "/"
	DefaultPortRangeFrom uint16 = 40001
	DefaultPortRangeTo   uint16 = 40009
)

func defaultConfig() *charter.Config {
	defaultDir, err := os.UserHomeDir()
	if err != nil {
		defaultDir = DefaultDir
	}

	return &charter.Config{
		Addr:       DefaultAddr,
		DefaultDir: defaultDir,
		PassivePortRange: charter.PassivePortRange{
			From: DefaultPortRangeFrom,
			To:   DefaultPortRangeTo,
		},
	}
}

func populateFromCliContext(conf *charter.Config, ctx *cli.Context) {
	if ctx.IsSet("no-anonymous") {
		conf.NoAnonymous = true
	}

	if ctx.IsSet("anonymous-only") {
		conf.AnonymousOnly = true
	}
}

func populateFromFile(conf *charter.Config, filename string) error {
	_, err := toml.DecodeFile(filename, conf)
	return err
}
