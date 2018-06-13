package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/errwrap"
	"github.com/mitchellh/cli"
	homedir "github.com/mitchellh/go-homedir"
)

var (
	appName    = "turndisk"
	appVersion = "0.0.0"
)

func sessionPath() (string, error) {
	hdir, err := homedir.Dir()
	if err != nil {
		return "", nil
	}

	dir := filepath.Join(hdir, "."+appName, "session.json")
	return dir, nil
}

func exit(ui cli.Ui, err error) int {
	if err == nil {
		return 0
	}

	ui.Error(fmt.Sprintf("Error: %v", err))
	return 1
}

type session struct {
	AccessToken  string    `json:"access_token"`
	IDToken      string    `json:"id_token"`
	Scope        string    `json:"scope"`
	ExpiresInSec int64     `json:"expires_in"`
	Expires      time.Time `json:"expires"`
	TokenType    string    `json:"token_type"`
}

func continueSession(ui cli.Ui, loginf cli.CommandFactory) (sess *session, err error) {
	sessp, err := sessionPath()
	if err != nil {
		return nil, err
	}

	login, err := loginf()
	if err != nil {
		return nil, errwrap.Wrapf("failed to create login command: {{err}}", err)
	}

	for {
		f, err := os.Open(sessp)
		if err != nil {
			if !os.IsNotExist(err) {
				return nil, errwrap.Wrapf("failed to open session file: {{err}}", err)
			}

			ui.Info("No existing session found, please log in to continue.")
			errc := login.Run([]string{})
			if errc != 0 {
				return nil, fmt.Errorf("login failed: exit code %d", errc)
			}

			continue //retry open
		}

		defer f.Close()
		sess = &session{}
		dec := json.NewDecoder(f)
		err = dec.Decode(sess)
		if err != nil {
			return nil, errwrap.Wrapf("failed to decode sessin: {{err}}", err)
		}

		if sess.Expires.Before(time.Now().Add(time.Minute * -10)) {
			ui.Info("Previous session expired, please log in again to continue.")
			errc := login.Run([]string{})
			if errc != 0 {
				return nil, fmt.Errorf("login failed: exit code %d", errc)
			}

			continue //retry open
		}

		return sess, nil
	}
}
