package cgroup

import (
	"github.com/Kailun2047/mydocker/subsystems"
)

type CgroupManager struct {
	CgroupPath string
	Config     *subsystems.ResourceConfig
}

func (m *CgroupManager) Set(config *subsystems.ResourceConfig) error {
	m.Config = config
	for _, subsys := range subsystems.SubsystemIns {
		err := subsys.Set(m.CgroupPath, config)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *CgroupManager) Apply(pid int) error {
	for _, subsys := range subsystems.SubsystemIns {
		err := subsys.Apply(m.CgroupPath, pid)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *CgroupManager) Destroy() error {
	for _, subsys := range subsystems.SubsystemIns {
		err := subsys.Remove(m.CgroupPath)
		if err != nil {
			return err
		}
	}
	return nil
}

func NewCgroupManager(cgroupPath string) *CgroupManager {
	return &CgroupManager{
		CgroupPath: cgroupPath,
	}
}
