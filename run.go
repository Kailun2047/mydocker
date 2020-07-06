package main

import (
	"os"

	"github.com/Kailun2047/mydocker/container"
	log "github.com/sirupsen/logrus"
)

func Run(command string, tty bool) {
	cmd := container.NewParentProcess(command, tty)
	if err := cmd.Start(); err != nil {
		log.Errorf(err.Error())
	}
	cmd.Wait()
	os.Exit(-1)
}
