package subsystems

type ResourceConfig struct {
	CpuSet   string
	CpuShare string
	Memory   string
}

type Subsystem interface {
	Name() string
	Set(cgroupPath string, config *ResourceConfig) error
	Apply(cgroupPath string, pid int) error
	Remove(cgroupPath string) error
}

var SubsystemIns = []Subsystem{
	&MemorySubsystem{},
}
