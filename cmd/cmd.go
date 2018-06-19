package main

import (
	"fmt"

	"github.com/mitchellh/cli"
)

var (
	appIssuer  = "https://datajob.eu.auth0.com/"
	appName    = "turndisk"
	appVersion = "0.0.0"
	appID      = "7VCcJU3IVivsPfirCbJGJZxEBYKP004p"
)

func exit(ui cli.Ui, err error) int {
	if err == nil {
		return 0
	}

	ui.Error(fmt.Sprintf("Error: %v", err))
	return 1
}
