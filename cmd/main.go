package main

import (
	"os"

	"github.com/mitchellh/cli"
)

func main() {
	ui := &cli.BasicUi{
		Reader:      os.Stdin,
		Writer:      os.Stdout,
		ErrorWriter: os.Stderr,
	}

	c := cli.NewCLI(appName, appVersion)
	c.Args = os.Args[1:]
	c.Commands = map[string]cli.CommandFactory{
		"login":      loginFactory(ui),
		"logout":     logoutFactory(ui),
		"disk mount": diskMountFactory(ui),
	}

	status, err := c.Run()
	if err != nil {
		ui.Error(err.Error())
	}

	os.Exit(status)
}
