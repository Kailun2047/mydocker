package container

import (
	"os"
	"os/exec"
	"syscall"

	log "github.com/sirupsen/logrus"
)

func NewParentProcess(commands []string, tty bool) *exec.Cmd {
	log.Infof("Creating parent process for command [%v]", commands)
	args := []string{"init"}
	args = append(args, commands...)
	cmd := exec.Command("/proc/self/exe", args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWNS | syscall.CLONE_NEWPID | syscall.CLONE_NEWIPC | syscall.CLONE_NEWNET | syscall.CLONE_NEWUTS,
	}
	if tty {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	return cmd
}

func RunContainerInitProcess(commands []string, args []string) error {
	log.Infof("Initializing container process for command [%v]", commands)
	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")
	argv := commands
	err := syscall.Exec(commands[0], argv, os.Environ())
	if err != nil {
		log.Errorf(err.Error())
	}
	return nil
}
