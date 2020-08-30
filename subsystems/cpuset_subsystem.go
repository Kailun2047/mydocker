package subsystems

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
)

type CpuSetSubsystem struct{}

func (s *CpuSetSubsystem) Name() string {
	return "cpuset"
}

func (s *CpuSetSubsystem) Set(cgroupPath string, config *ResourceConfig) error {
	absCgroupPath, err := GetCgroupPath(cgroupPath, s.Name(), true)
	if err != nil {
		return err
	}
	cpuset := strings.Split(config.CpuSet, " ")
	if len(cpuset) != 2 {
		return fmt.Errorf("Invalid cpuset configuration (want 2 values but get %d)", len(cpuset))
	}
	cpusetCpus, cpusetMems := cpuset[0], cpuset[1]
	if err = ioutil.WriteFile(path.Join(absCgroupPath, "cpuset.cpus"), []byte(cpusetCpus), 0644); err != nil {
		return fmt.Errorf("Failed to write cpu set cpu limit: [%v]", err)
	}
	if err = ioutil.WriteFile(path.Join(absCgroupPath, "cpuset.mems"), []byte(cpusetMems), 0644); err != nil {
		return fmt.Errorf("Failed to write cpu set mem limit: [%v]", err)
	}
	return nil
}

func (s *CpuSetSubsystem) Apply(cgroupPath string, pid int) error {
	absCgroupPath, err := GetCgroupPath(cgroupPath, s.Name(), true)
	if err != nil {
		return err
	}
	if err = ioutil.WriteFile(path.Join(absCgroupPath, "tasks"), []byte(strconv.Itoa(pid)), 0644); err != nil {
		return fmt.Errorf("Failed to add process [%d] to cgroup [%s]: [%v]", pid, absCgroupPath, err)
	}
	return nil
}

func (s *CpuSetSubsystem) Remove(cgroupPath string) error {
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
