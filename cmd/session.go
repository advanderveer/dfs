package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/hashicorp/errwrap"
	"github.com/mitchellh/cli"
	homedir "github.com/mitchellh/go-homedir"
)

//return the place on the users home directory were the session is stored
func sessionPath() (string, error) {
	hdir, err := homedir.Dir()
	if err != nil {
		return "", nil
	}

	dir := filepath.Join(hdir, "."+appName, "session.json")
	return dir, nil
}

//session type is used for persisting a session
type session struct {
	AccessToken  string    `json:"access_token"`
	IDToken      string    `json:"id_token"`
	Scope        string    `json:"scope"`
	ExpiresInSec int64     `json:"expires_in"`
	Expires      time.Time `json:"expires"`
	TokenType    string    `json:"token_type"`
}

func (sess *session) validatedClaims() (c jwt.MapClaims, err error) {
	var token *jwt.Token
	if token, err = jwt.Parse(sess.IDToken, func(token *jwt.Token) (interface{}, error) {
		checkAud := token.Claims.(jwt.MapClaims).VerifyAudience(appID, false)
		if !checkAud {
			return token, fmt.Errorf("invalid audience, claims: %v", token.Claims)
		}

		iss := appIssuer
		checkIss := token.Claims.(jwt.MapClaims).VerifyIssuer(iss, false)
		if !checkIss {
			return token, fmt.Errorf("invalid issuer, claims: %v", token.Claims)
		}

		cert, err := getPemCert(token)
		if err != nil {
			return nil, fmt.Errorf("failed to get signing certificate: %v", err)
		}

		return jwt.ParseRSAPublicKeyFromPEM([]byte(cert))
	}); err != nil {
		return nil, errwrap.Wrapf("failed to parse session token: {{err}}", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token, error: %v", err)
	}

	c, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("session token doesn't have map claims")
	}

	return c, nil
}

//get the certificate from well known endpoint, source: https://auth0.com/docs/quickstart/backend/golang/01-authorization
func getPemCert(token *jwt.Token) (string, error) {
	type JSONWebKeys struct {
		Kty string   `json:"kty"`
		Kid string   `json:"kid"`
		Use string   `json:"use"`
		N   string   `json:"n"`
		E   string   `json:"e"`
		X5c []string `json:"x5c"`
	}

	type Jwks struct {
		Keys []JSONWebKeys `json:"keys"`
	}

	cert := ""
	resp, err := http.Get(fmt.Sprintf("%s.well-known/jwks.json", appIssuer))
	if err != nil {
		return cert, err
	}

	defer resp.Body.Close()
	var jwks = Jwks{}
	err = json.NewDecoder(resp.Body).Decode(&jwks)

	if err != nil {
		return cert, err
	}

	for k := range jwks.Keys {
		if token.Header["kid"] == jwks.Keys[k].Kid {
			cert = "-----BEGIN CERTIFICATE-----\n" + jwks.Keys[k].X5c[0] + "\n-----END CERTIFICATE-----"
		}
	}

	if cert == "" {
		err := errors.New("Unable to find appropriate key")
		return cert, err
	}

	return cert, nil
}

//Continue session will attempt to load the session from a file on the users computer
//and issue a login command if this is not present or has expired
func continueSession(ui cli.Ui, loginf cli.CommandFactory) (sess *session, claims jwt.MapClaims, err error) {
	sessp, err := sessionPath()
	if err != nil {
		return nil, nil, err
	}

	login, err := loginf()
	if err != nil {
		return nil, nil, errwrap.Wrapf("failed to create login command: {{err}}", err)
	}

	for {
		f, err := os.Open(sessp)
		if err != nil {
			if !os.IsNotExist(err) {
				return nil, nil, errwrap.Wrapf("failed to open session file: {{err}}", err)
			}

			ui.Info("No existing session found, please log in to continue.")
			errc := login.Run([]string{})
			if errc != 0 {
				return nil, nil, fmt.Errorf("login failed: exit code %d", errc)
			}

			continue //retry open
		}

		defer f.Close()
		sess = &session{}
		dec := json.NewDecoder(f)
		err = dec.Decode(sess)
		if err != nil {
			return nil, nil, errwrap.Wrapf("failed to decode session: {{err}}", err)
		}

		if sess.Expires.Before(time.Now().Add(time.Minute * -10)) {
			ui.Info("Previous session expired, please log in again to continue.")
			errc := login.Run([]string{})
			if errc != 0 {
				return nil, nil, fmt.Errorf("login failed: exit code %d", errc)
			}

			continue //retry open
		}

		c, err := sess.validatedClaims()
		if err != nil {
			return nil, nil, err
		}

		return sess, c, nil
	}
}
