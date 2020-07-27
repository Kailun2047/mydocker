package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/Kailun2047/mydocker/cgroup"
	"github.com/Kailun2047/mydocker/container"
	"github.com/Kailun2047/mydocker/subsystems"
	log "github.com/sirupsen/logrus"
)

func sendInitCommands(writePipe *os.File, commands []string) error {
	_, err := io.WriteString(writePipe, strings.Join(commands, " "))
	if err != nil {
		return fmt.Errorf("Failed to write pipe: [%v]", err)
	}
	return nil
}

func Run(commands []string, tty bool, config *subsystems.ResourceConfig) {
	cmd, writePipe := container.NewParentProcess(tty)
	if err := cmd.Start(); err != nil {
		log.Errorf(err.Error())
	}
	sendInitCommands(writePipe, commands)
	cgroupManager := cgroup.NewCgroupManager("mydocker-cgroup")
	cgroupManager.Set(config)
	cgroupManager.Apply(cmd.Process.Pid)
	defer cgroupManager.Destroy()
	cmd.Wait()
	os.Exit(-1)
}
