package container

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"syscall"

	log "github.com/sirupsen/logrus"
)

var (
	ContainerRootUrl   = "/root/"
	ContainerMntUrl    = "/root/mnt/"
	busyboxTarFileName = "busybox.tar"
	busyboxDir         = "busybox"
	writeLayerDir      = "writeLayer"
	workDir            = "work"
)

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

func PathExist(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func CreateReadOnlyLayer(rootUrl string) {
	readOnlyLayerUrl := path.Join(rootUrl, busyboxDir) + "/"
	exist, err := PathExist(readOnlyLayerUrl)
	if err != nil {
		log.Errorf("Can't decide if read-only layer [%s] exists or not: [%v]", readOnlyLayerUrl, err)
		return
	}
	if exist {
		log.Infof("Read-only layer [%s] already exists", readOnlyLayerUrl)
		return
	}
	if err := os.MkdirAll(readOnlyLayerUrl, 0777); err != nil {
		log.Errorf("Failed to create read-only layer dir [%s]: [%v]", readOnlyLayerUrl, err)
	}
	busyboxTarUrl := path.Join(rootUrl, busyboxTarFileName)
	cmd := exec.Command("tar", "-xvf", busyboxTarUrl, "-C", readOnlyLayerUrl)
	log.Infof("Creating read-only layer [%s] ...", readOnlyLayerUrl)
	combinedOutput, err := cmd.CombinedOutput()
	if err != nil {
		log.Errorf("Failed to create read-only layer [%s]; command output: [%s]", readOnlyLayerUrl, string(combinedOutput))
	}
}

func CreateWriteLayer(rootUrl string) {
	writeLayerUrl := path.Join(rootUrl, writeLayerDir)
	cmd := exec.Command("mkdir", "-p", writeLayerUrl)
	if err := cmd.Run(); err != nil {
		log.Errorf("Failed to created write layer [%s]", writeLayerUrl)
	}
}

func CreateWorkDir(rootUrl string) {
	workDirUrl := path.Join(rootUrl, workDir)
	cmd := exec.Command("mkdir", "-p", workDirUrl)
	if err := cmd.Run(); err != nil {
		log.Errorf("Failed to created work directory [%s]", workDirUrl)
	}
}

func CreateMountPoint(rootUrl, mntUrl string) {
	if err := os.MkdirAll(mntUrl, 0777); err != nil {
		log.Errorf("Failed to create mount point [%s]: [%v]", mntUrl, err)
		return
	}
	readLayerUrl := path.Join(rootUrl, busyboxDir)
	writeLayerUrl := path.Join(rootUrl, writeLayerDir)
	workDirUrl := path.Join(rootUrl, workDir)
	overlayfsDirs := "lowerdir=" + readLayerUrl + ",upperdir=" + writeLayerUrl + ",workdir=" + workDirUrl
	cmd := exec.Command("mount", "-t", "overlay", "-o", overlayfsDirs, "none", mntUrl)
	combinedOutput, err := cmd.CombinedOutput()
	if err != nil {
		log.Errorf("Failed to mount overlayfs at [%s]; command output: [%s]", mntUrl, string(combinedOutput))
	}
}

func DeleteMountPoint(rootUrl, mntUrl string) {
	cmd := exec.Command("umount", mntUrl)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("Failed to un-mount overlayfs at [%s]: [%v]", mntUrl, err)
		return
	}
	if err := os.RemoveAll(mntUrl); err != nil {
		log.Errorf("Failed to remove container's root filesystem")
	}
}

func DeleteWriteLayer(rootUrl string) {
	if err := os.RemoveAll(path.Join(rootUrl, writeLayerDir)); err != nil {
		log.Errorf("Failed to remove container's writable layer")
	}
}

func NewWorkspace(rootUrl, mntUrl string) {
	CreateReadOnlyLayer(rootUrl)
	CreateWriteLayer(rootUrl)
	CreateWorkDir(rootUrl)
	CreateMountPoint(rootUrl, mntUrl)
}

func DeleteWorkspace(rootUrl, mntUrl string) {
	DeleteMountPoint(rootUrl, mntUrl)
	DeleteWriteLayer(rootUrl)
}
