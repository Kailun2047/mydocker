package container

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
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
	// Start container process in a new root file system.
	cmd.Dir = "/root/busybox"
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

func pivotRoot(rootDir string) error {
	// Bind mount rootDir to a new mount point so that the new root and old root can be in different filesystems.
	err := syscall.Mount(rootDir, rootDir, "bind", uintptr(syscall.MS_BIND|syscall.MS_REC), "")
	if err != nil {
		return fmt.Errorf("Could not mount root file system: [%v]", err)
	}
	oldRootDir := ".pivot_root"
	pivotRootDir := path.Join(rootDir, oldRootDir)
	err = os.Mkdir(pivotRootDir, 0777)
	if err != nil {
		return fmt.Errorf("Failed to create directory for old root fs: [%v]", err)
	}
	err = syscall.PivotRoot(rootDir, pivotRootDir)
	if err != nil {
		return fmt.Errorf("Failed to pivot root: [%v]", err)
	}
	// Change to new root dir, unmount old rootfs and remove old root dir.
	err = os.Chdir("/")
	if err != nil {
		return fmt.Errorf("Could not change to new root directory")
	}
	pivotRootDir = path.Join("/", oldRootDir)
	err = syscall.Unmount(pivotRootDir, syscall.MNT_DETACH)
	if err != nil {
		return fmt.Errorf("Could not unmount old root file system: [%v]", err)
	}
	return os.Remove(pivotRootDir)
}

func setupMount() error {
	curDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("Failed to get pwd: [%v]", err)
	}
	log.Infof("Changed to directory [%v]", curDir)
	err = syscall.Mount("", "/", "", uintptr(syscall.MS_PRIVATE|syscall.MS_REC), "")
	err = pivotRoot(curDir)
	if err != nil {
		return err
	}
	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	err = syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")
	if err != nil {
		return fmt.Errorf("Failed to mount proc: [%v]", err)
	}
	return syscall.Mount("tmpfs", "/dev", "tmpfs", uintptr(syscall.MS_NOSUID|syscall.MS_STRICTATIME), "mode=755")
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
