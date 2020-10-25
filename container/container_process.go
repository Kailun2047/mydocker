package container

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"syscall"

	log "github.com/sirupsen/logrus"
)

// NewParentProcess returns a process (not started) that create a container
// (i.e. a process with its own several namespaces), with commands running in it.
// The pipe is used to deal with long arguments and special chars in command.
func NewParentProcess(tty bool) (*exec.Cmd, *os.File) {
	log.Infof("Creating parent process")
	args := []string{"init"}
	readPipe, writePipe, err := os.Pipe()
	if err != nil {
		log.Errorf("Failed to create pipe: [%v]", err)
		return nil, nil
	}
	cmd := exec.Command("/proc/self/exe", args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWNS | syscall.CLONE_NEWPID | syscall.CLONE_NEWIPC | syscall.CLONE_NEWNET | syscall.CLONE_NEWUTS,
	}
	cmd.ExtraFiles = []*os.File{readPipe}
	if tty {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	// Use OverlayFS for the new root filesystem and start container process in.
	NewWorkspace(ContainerRootUrl, ContainerMntUrl)
	cmd.Dir = ContainerMntUrl
	return cmd, writePipe
}

func getInitCommands() ([]string, error) {
	// The init cmd (process) is created with read pipe as extra file, so we can use
	// file descriptor 3 (0 to 2 are used for stdin, stdout and stderr) to open it.
	readPipe := os.NewFile(3, "read_pipe")
	commandBytes, err := ioutil.ReadAll(readPipe)
	if err != nil {
		return nil, fmt.Errorf("Failed to read pipe: [%v]", err)
	}
	commands := strings.Split(string(commandBytes), " ")
	return commands, nil
}

// RunContainerInitProcess initialize container process. Specifically, it mount /proc and
// invokes execve system call through syscall.Exec() to replace the current process with
// PID 1 with the command that the container is started with. The commands are read out of
// the other end of the pipe that's created in the parent process.
func RunContainerInitProcess() error {
	log.Infof("Initializing container process")
	err := setupMount()
	if err != nil {
		return fmt.Errorf("Error when setting up mounts for container process: [%v]", err)
	}
	commands, err := getInitCommands()
	if err != nil {
		return err
	}
	if commands == nil || len(commands) == 0 {
		return fmt.Errorf("Got empty run commands")
	}
	cmdPath, err := exec.LookPath(commands[0])
	if err != nil {
		return fmt.Errorf("Failed to find command path: [%v]", err)
	}
	return syscall.Exec(cmdPath, commands, os.Environ())
}
