package main

import (
	"fmt"
	"os"

	cli "github.com/urfave/cli/v2"
)

// Exit statuses
const (
	ExitOK  = 0
	ExitErr = 1
)

var (
	portFlag = &cli.UintFlag{
		Name:    "port",
		Aliases: []string{"p"},
		Value:   8000,
	}
	intervalFlag = &cli.Int64Flag{
		Name:    "interval",
		Aliases: []string{"i"},
		Value:   60,
	}
	tokenFlag = &cli.StringFlag{
		Name:    "token",
		Aliases: []string{"t"},
	}
)

func newApp() *cli.App {
	app := cli.NewApp()
	app.Usage = ""
	app.HideVersion = true

	app.Flags = []cli.Flag{portFlag, intervalFlag, tokenFlag}
	app.Action = execute

	return app
}

func main() {
	app := newApp()
	os.Exit(printOnError(app.Run(os.Args)))
}

func printOnError(err error) int {
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		return ExitErr
	}
	return ExitOK
}
