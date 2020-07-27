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

// Subsystem instances. Here we deal with 3 of them: memory, cpu set and cpu share.
// Limit on memory is specified with memory in bytes; limit on cpu set is specified
// in comma separated values with dash indicating a range (e.g. 0-2,4); cpu share is
// specified with a value larger than 2 and will be evaluated relatively to the value
// set for other processes.
var SubsystemIns = []Subsystem{
	&MemorySubsystem{},
	&CpuSetSubsystem{},
	&CpuShareSubsystem{},
}
