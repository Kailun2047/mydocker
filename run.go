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
	writePipe.Close()
	return nil
}

func Run(commands []string, tty bool, config *subsystems.ResourceConfig) {
	cmd, writePipe := container.NewParentProcess(tty)
	if err := cmd.Start(); err != nil {
		log.Errorf("Command failed to start: [%v]", err.Error())
		return
	}
	sendInitCommands(writePipe, commands)
	cgroupManager := cgroup.NewCgroupManager("mydocker-cgroup")
	err := cgroupManager.Set(config)
	if err != nil {
		log.Errorf("Failed to set resource configurations: [%v]", err)
		return
	}
	err = cgroupManager.Apply(cmd.Process.Pid)
	if err != nil {
		log.Errorf("Failed to apply resource configurations: [%v]", err)
	}
	defer cgroupManager.Destroy()
	err = cmd.Wait()
	if err != nil {
		log.Errorf("Failed to execute command [%v]: [%v]", cmd.Args, err)
	}
	os.Exit(0)
}
