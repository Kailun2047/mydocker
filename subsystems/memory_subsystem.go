package subsystems

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
)

type MemorySubsystem struct{}

func (s *MemorySubsystem) Name() string {
	return "memory"
}

func (s *MemorySubsystem) Set(cgroupPath string, config *ResourceConfig) error {
	if len(config.Memory) == 0 {
		return nil
	}
	absCgroupPath, err := GetCgroupPath(cgroupPath, s.Name(), true)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(path.Join(absCgroupPath, "memory.limit_in_bytes"), []byte(config.Memory), 0644)
	if err != nil {
		return fmt.Errorf("Failed to write memory limit: [%v]", err)
	}
	return nil
}

func (s *MemorySubsystem) Apply(cgroupPath string, pid int) error {
	absCgroupPath, err := GetCgroupPath(cgroupPath, s.Name(), true)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(path.Join(absCgroupPath, "tasks"), []byte(strconv.Itoa(pid)), 0644)
	if err != nil {
		return fmt.Errorf("Failed to add pid [%d] to cgroup [%s]: [%v]", pid, absCgroupPath, err)
	}
	return nil
}

func (s *MemorySubsystem) Remove(cgroupPath string) error {
	absCgroupPath, err := GetCgroupPath(cgroupPath, s.Name(), false)
	if err != nil {
		return err
	}
	err = os.RemoveAll(absCgroupPath)
	if err != nil {
		return fmt.Errorf("Failed to remove cgroup [%s]", absCgroupPath)
	}
	return nil
}
