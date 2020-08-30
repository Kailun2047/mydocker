package main

import (
	"os"
	"strings"

	"github.com/Kailun2047/mydocker/container"
	"github.com/Kailun2047/mydocker/subsystems"
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
		&cli.StringFlag{
			Name:  "m",
			Usage: "specify memory limit for container",
		},
		&cli.StringFlag{
			Name:  "c",
			Usage: "specify cpu share for container",
		},
		&cli.StringFlag{
			Name:  "cpuset-cpus",
			Usage: "specify cpu set that can be used by container",
		},
		&cli.StringFlag{
			Name:  "cpuset-mems",
			Usage: "specify memory node set that can be used by container",
		},
	},
	Action: func(ctx *cli.Context) error {
		if len(os.Args) < 1 {
			log.Warn("Container command arguments can't be empty")
			os.Exit(1)
		}
		tty := ctx.Bool("it")
		resourceConfig := &subsystems.ResourceConfig{
			Memory:   ctx.String("m"),
			CpuShare: ctx.String("c"),
			CpuSet:   strings.Join([]string{ctx.String("cpuset-cpus"), ctx.String("cpuset-mems")}, " "),
		}
		log.Infof("Resource config: [%v]", *resourceConfig)
		args := ctx.Args()
		commands := args.Slice()
		Run(commands, tty, resourceConfig)
		return nil
	},
}

var initCommand = &cli.Command{
	Name:  "init",
	Usage: "init container process",
	Action: func(ctx *cli.Context) error {
		err := container.RunContainerInitProcess()
		return err
	},
}
