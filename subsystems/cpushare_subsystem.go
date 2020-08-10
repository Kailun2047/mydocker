package subsystems

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
)

type CpuShareSubsystem struct{}

func (s *CpuShareSubsystem) Name() string {
	return "cpu"
}

func (s *CpuShareSubsystem) Set(cgroupPath string, config *ResourceConfig) error {
	if len(config.CpuShare) == 0 {
		return nil
	}
	absCgroupPath, err := GetCgroupPath(cgroupPath, s.Name(), true)
	if err != nil {
		return err
	}
	cpushares := config.CpuShare
	if err = ioutil.WriteFile(path.Join(absCgroupPath, "cpu.shares"), []byte(cpushares), 0644); err != nil {
		return fmt.Errorf("Failed to write cpu share limit: [%v]", err)
	}
	return nil
}

func (s *CpuShareSubsystem) Apply(cgroupPath string, pid int) error {
	absCgroupPath, err := GetCgroupPath(cgroupPath, s.Name(), true)
	if err != nil {
		return err
	}
	if err = ioutil.WriteFile(path.Join(absCgroupPath, "tasks"), []byte(strconv.Itoa(pid)), 0644); err != nil {
		return fmt.Errorf("Failed to add process [%d] to cgroup [%s]: [%v]", pid, absCgroupPath, err)
	}
	return nil
}

func (s *CpuShareSubsystem) Remove(cgroupPath string) error {
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
