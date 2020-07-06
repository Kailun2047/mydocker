package main

import (
	"os"

	"github.com/Kailun2047/mydocker/container"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

var runCommand = &cli.Command{
	Name:  "run",
	Usage: "docker run implementation",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "it",
			Usage: "interactive mode with tty",
		},
	},
	Action: func(ctx *cli.Context) error {
		if len(os.Args) < 1 {
			log.Warn("Container command arguments can't be empty")
			os.Exit(1)
		}
		args := ctx.Args()
		command := args.Get(0)
		tty := ctx.Bool("it")
		Run(command, tty)
		return nil
	},
}

var initCommand = &cli.Command{
	Name:  "init",
	Usage: "init container process",
	Action: func(ctx *cli.Context) error {
		command := ctx.Args().Get(0)
		err := container.RunContainerInitProcess(command, nil)
		return err
	},
}
