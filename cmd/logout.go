package main

import (
	"fmt"
	"os"

	"github.com/hashicorp/errwrap"
	"github.com/mitchellh/cli"
)

type logout struct {
	ui cli.Ui
}

func logoutFactory(ui cli.Ui) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return &logout{ui}, nil
	}
}

func (cmd *logout) Help() string     { return cmd.Synopsis() }
func (cmd *logout) Synopsis() string { return "" }

func (cmd *logout) Run(args []string) int {
	sessp, err := sessionPath()
	if err != nil {
		return exit(cmd.ui, errwrap.Wrapf("failed determine session file location: {{err}}", err))
	}

	err = os.Remove(sessp)
	if err != nil {
		if os.IsNotExist(err) {
			return exit(cmd.ui, fmt.Errorf("no session file found, nothing to remove"))
		}

		return exit(cmd.ui, errwrap.Wrapf("failed to remove session file: {{err}}", err))
	}

	cmd.ui.Info("Logged out successfully!")
	return 0
}
