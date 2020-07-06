package main

import (
	"os"

	"github.com/Kailun2047/mydocker/cgroup"
	"github.com/Kailun2047/mydocker/container"
	"github.com/Kailun2047/mydocker/subsystems"
	log "github.com/sirupsen/logrus"
)

func Run(commands []string, tty bool, config *subsystems.ResourceConfig) {
	cmd := container.NewParentProcess(commands, tty)
	if err := cmd.Start(); err != nil {
		log.Errorf(err.Error())
	}
	cgroupManager := cgroup.NewCgroupManager("mydocker-cgroup")
	cgroupManager.Set(config)
	cgroupManager.Apply(cmd.Process.Pid)
	defer cgroupManager.Destroy()
	cmd.Wait()
	os.Exit(-1)
}
