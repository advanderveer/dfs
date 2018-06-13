package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/errwrap"
	"github.com/mitchellh/cli"
)

type login struct {
	ui cli.Ui
}

func loginFactory(ui cli.Ui) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return &login{ui}, nil
	}
}

func (cmd *login) Help() string     { return cmd.Synopsis() }
func (cmd *login) Synopsis() string { return "" }

func (cmd *login) Run(args []string) int {
	var (
		resp    *http.Response
		payload []byte
	)

RETRY:
	for {
		username, err := cmd.ui.Ask("username (or email):")
		if err != nil {
			return exit(cmd.ui, errwrap.Wrapf("failed to ask for username: {{err}}", err))
		}

		password, err := cmd.ui.AskSecret("password (hidden):")
		if err != nil {
			return exit(cmd.ui, errwrap.Wrapf("failed to ask for password: {{err}}", err))
		}

		if username == "" || password == "" {
			cmd.ui.Info("Please provide your username and password, or use Ctrl+C to cancel.")
			continue
		}

		if payload, err = json.Marshal(struct {
			Username  string `json:"username"`
			Password  string `json:"password"`
			GrantType string `json:"grant_type"`
			Audience  string `json:"audience"`
			ClientID  string `json:"client_id"`
			Scope     string `json:"scope"`
		}{
			Username:  username,
			Password:  password,
			GrantType: "password",
			Audience:  "api.turndisk.com",
			ClientID:  "7VCcJU3IVivsPfirCbJGJZxEBYKP004p",
			Scope:     "openid",
		}); err != nil {
			return exit(cmd.ui, errwrap.Wrapf("failed to marshal credentials: {{err}}", err))
		}

		resp, err = http.Post("https://datajob.eu.auth0.com/oauth/token", "application/json", bytes.NewReader(payload))
		if err != nil {
			return exit(cmd.ui, errwrap.Wrapf("failed to post credentials: {{err}}", err))
		}

		switch resp.StatusCode {
		case http.StatusForbidden:
			cmd.ui.Info("Invalid username or password, try again or use Ctrl+C to cancel.")
		case http.StatusOK:
			break RETRY
		default:
			cmd.ui.Info(fmt.Sprintf("unexpected response from authentication server: %v, try again or use ctrl+c to cancel.", resp.Status))
		}
	}

	sess := &session{}
	dec := json.NewDecoder(resp.Body)
	err := dec.Decode(sess)
	if err != nil {
		return exit(cmd.ui, errwrap.Wrapf("failed to decode authentication session: {{err}}", err))
	}

	sess.Expires = time.Now().Add(time.Second * time.Duration(sess.ExpiresInSec))
	//@TODO validate id signature

	sessp, err := sessionPath()
	if err != nil {
		return exit(cmd.ui, errwrap.Wrapf("failed determine session location: {{err}}", err))
	}

	err = os.MkdirAll(filepath.Dir(sessp), 0700)
	if err != nil {
		return exit(cmd.ui, errwrap.Wrapf("failed create session directory: {{err}}", err))
	}

	f, err := os.Create(sessp)
	if err != nil {
		return exit(cmd.ui, errwrap.Wrapf("failed create session file: {{err}}", err))
	}

	defer f.Close()
	enc := json.NewEncoder(f)
	err = enc.Encode(sess)
	if err != nil {
		return exit(cmd.ui, errwrap.Wrapf("failed to write session: {{err}}", err))
	}

	cmd.ui.Info("Logged in successfully!")
	return 0
}
