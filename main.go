package main

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "mydocker",
		Usage: "simple docker runtime implementation",
		Commands: []*cli.Command{
			runCommand,
			initCommand,
		},
		Before: func(ctx *cli.Context) error {
			log.SetFormatter(&log.JSONFormatter{})
			log.SetOutput(os.Stdout)
			return nil
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatalf(err.Error())
	}
}
