package main

import (
	"log"
	"os"

	"strucim/proxy/commands"

	"github.com/joho/godotenv"
	"github.com/urfave/cli/v2"
)

func mainOrig() {

	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file", err)
	}

	// CLI App takes care of parsing commands, flags and printing help etc.
	app := &cli.App{
		Name:  "struCIM",
		Usage: "struCIM cloud interface for part identification",
	}

	// global flags i.e. proxy --verbose
	app.Flags = []cli.Flag{
		&cli.BoolFlag{
			Name:  "verbose, v",
			Usage: "Enable detailed output",
		},
	}

	app.Commands = []*cli.Command{
		{
			Name:    "identify",
			Aliases: []string{"id"},
			Usage:   "Identify a part from point cloud csv `FILE`",
			Action:  commands.Identify,
		},
		{
			Name:   "heartbeat",
			Usage:  "Tests connectivity and authentication with the cloud",
			Action: commands.Heartbeat,
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
