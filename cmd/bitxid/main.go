package main

import (
	"fmt"
	"os"
	"time"

	"github.com/urfave/cli/v2"
)

func main() {
	app := cli.NewApp()
	app.Name = "BitXID"
	app.Usage = "A revolutionary identity library of Web3"
	app.Compiled = time.Now()

	cli.VersionPrinter = func(c *cli.Context) {
		printVersion()
	}

	// global flags
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:  "repo",
			Usage: "BitXID repo path",
		},
	}

	app.Commands = []*cli.Command{
		versionCMD(),
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
	}
}
