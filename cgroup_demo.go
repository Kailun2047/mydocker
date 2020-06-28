package main

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"syscall"
	"io/ioutil"
	"strconv"
)

var (
	cgroupMemoryHierarchyMount = "/sys/fs/cgroup/memory"
	memoryLimitConfigFile = "memory.limit_in_bytes"
	tasksFile = "tasks"
)

func main() {
	if os.Args[0] == "/proc/self/exe" {
		cmd := exec.Command("sh", "-c", "stress --vm-bytes 200m --vm-keep -m 1")
		cmd.SysProcAttr = &syscall.SysProcAttr{}
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
	// Use "/proc/self/exe" to rerun this program.
	cmd := exec.Command("/proc/self/exe")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	} else {
		// Create new cgroup with memory limit and put pid for this program into it.
		memoryLimitCgroupDir := path.Join(cgroupMemoryHierarchyMount, "memorylimittest")
		err := os.Mkdir(memoryLimitCgroupDir, 0755)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		ioutil.WriteFile(path.Join(memoryLimitCgroupDir, memoryLimitConfigFile), []byte("100m"), 0644)
		curPid := cmd.Process.Pid
		fmt.Println("Pid of current process: ", curPid)
		ioutil.WriteFile(path.Join(memoryLimitCgroupDir, tasksFile), []byte(strconv.Itoa(curPid)), 0644)
		err = cmd.Wait()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
}