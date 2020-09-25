package main

import (
	"fmt"

	"github.com/meshplus/bitxid"
	"github.com/urfave/cli/v2"
)

func versionCMD() *cli.Command {
	return &cli.Command{
		Name:   "version",
		Usage:  "BitXID version",
		Action: version,
	}
}

func version(ctx *cli.Context) error {
	printVersion()

	return nil
}

func printVersion() {
	fmt.Printf("BitXID version: %s-%s-%s\n", bitxid.CurrentVersion, bitxid.CurrentBranch, bitxid.CurrentCommit)
	fmt.Printf("App build date: %s\n", bitxid.BuildDate)
	fmt.Printf("System version: %s\n", bitxid.Platform)
	fmt.Printf("Golang version: %s\n", bitxid.GoVersion)
}
