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

	// url := "https://datajob.eu.auth0.com/oauth/token"
	//
	// payload := strings.NewReader("{\"grant_type\":\"password\",\"username\": \"a.vanderveer@nerdalize.com\",\"password\": \"Password1!\",\"audience\": \"api.turndisk.com\", \"client_id\": \"7VCcJU3IVivsPfirCbJGJZxEBYKP004p\", \"scope\": \"openid\"}")
	//
	// req, _ := http.NewRequest("POST", url, payload)
	//
	// req.Header.Add("content-type", "application/json")
	//
	// res, _ := http.DefaultClient.Do(req)
	//
	// defer res.Body.Close()
	// body, _ := ioutil.ReadAll(res.Body)
	//
	// fmt.Println(string(body))

}
