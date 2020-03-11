package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli"
)

// Exit statuses
const (
	ExitOK  = 0
	ExitErr = 1
)

var (
	registryDockerHubFlag = cli.StringFlag{
		Name:     "registry-docker-hub",
		Required: true,
	}
	registryFilterFlag = cli.StringFlag{
		Name: "registry-filter",
	}
	repositoryGitFlag = cli.StringFlag{
		Name:     "repository-git",
		Required: true,
	}
	repositoryBranchFlag = cli.StringFlag{
		Name:  "repository-branch",
		Value: "master",
	}
	repositoryPathFlag = cli.StringFlag{
		Name: "repository-path",
	}
)

func newApp() *cli.App {
	app := cli.NewApp()
	app.Usage = ""
	app.HideVersion = true

	app.Flags = []cli.Flag{
		registryDockerHubFlag,
		registryFilterFlag,
		repositoryGitFlag,
		repositoryBranchFlag,
		repositoryPathFlag,
	}
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
