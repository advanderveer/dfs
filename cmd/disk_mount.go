package main

import (
	"fmt"

	"github.com/hashicorp/errwrap"
	"github.com/mitchellh/cli"
)

type diskMount struct {
	ui cli.Ui
}

func diskMountFactory(ui cli.Ui) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return &diskMount{ui}, nil
	}
}

func (cmd *diskMount) Help() string     { return cmd.Synopsis() }
func (cmd *diskMount) Synopsis() string { return "" }

func (cmd *diskMount) Run(args []string) int {
	sess, err := continueSession(cmd.ui, loginFactory(cmd.ui))
	if err != nil {
		return exit(cmd.ui, errwrap.Wrapf("failed to continue session: {{err}}", err))
	}

	fmt.Println(sess.Expires)

	return 0
}
