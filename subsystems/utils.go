package subsystems

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strings"
)

// Get the mount point of input subsystem for current process.
func GetSubsystemMountPoint(subsystem string) (string, error) {
	f, err := os.Open("/proc/self/mountinfo")
	if err != nil {
		return "", fmt.Errorf("Failed to read mountinfo: [%v]", err)
	}
	// Example line of mountinfo:
	// 44 31 0:39 / /sys/fs/cgroup/memory rw,nosuid,nodev,noexec,relatime shared:23 - cgroup cgroup rw,memory
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fields := strings.Split(scanner.Text(), " ")
		subsysInfo := strings.Split(fields[len(fields)-1], ",")
		if subsystem == subsysInfo[len(subsysInfo)-1] {
			return fields[4], nil
		}
	}
	return "", fmt.Errorf("[%s] subsystem mount point not found in mountinfo; scanning error: [%v]", subsystem, scanner.Err())
}

// Get the absolute path of the cgroup under the specified subsystem for cgroup cgroupPath.
// If autoCreate is set to true, a new cgroup will be created if not already exists.
func GetCgroupPath(cgroupPath, subsystem string, autoCreate bool) (string, error) {
	subsystemMountPoint, err := GetSubsystemMountPoint(subsystem)
	if err != nil {
		return "", err
	}
	absCgroupPath := path.Join(subsystemMountPoint, cgroupPath)
	if _, err := os.Stat(absCgroupPath); err != nil {
		if os.IsNotExist(err) && autoCreate {
			err := os.Mkdir(absCgroupPath, 0755)
			if err != nil {
				return "", fmt.Errorf("Failed to create cgroup path %s", absCgroupPath)
			}
		} else {
			return "", fmt.Errorf("Can't find cgroup path %s", absCgroupPath)
		}
	}
	return absCgroupPath, nil
}
