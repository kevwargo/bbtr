package bootsnap

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"golang.org/x/sys/unix"
)

func run(mountpoint string, kernelParams []string) error {
	rootDev, err := findRootDev(kernelParams)
	if err != nil {
		return err
	}

	if err = mount(rootDev, mountpoint); err != nil {
		return err
	}

	return nil
}

func findRootDev(kernelParams []string) (string, error) {
	for _, param := range kernelParams {
		if val, ok := strings.CutPrefix(param, rootParam); ok && val != "" {
			return val, nil
		}
	}

	return "", fmt.Errorf("%q* kernel param not found", rootParam)
}

func mount(rootDev, mountpoint string) error {
	if isBtrfsSubvol(mountpoint) {
		return nil
	}

	if err := os.MkdirAll(mountpoint, 0o755); err != nil {
		return fmt.Errorf("creating mountpoint %q: %w", mountpoint, err)
	}

	cmd := exec.Command("mount", "-t", "btrfs", rootDev, mountpoint)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	log.Println(cmd)

	return cmd.Run()
}

func isBtrfsSubvol(path string) bool {
	stat, err := os.Stat(path)
	if err != nil {
		return false
	}

	var statfs unix.Statfs_t
	if unix.Statfs(path, &statfs) != nil {
		return false
	}

	sys, ok := stat.Sys().(*syscall.Stat_t)

	return ok && sys != nil && sys.Ino == btrfsSubvolInode && statfs.Type == unix.BTRFS_SUPER_MAGIC
}

type volume struct {
	path      string
	snapshots []string
}

const (
	rootParam        = "root="
	btrfsSubvolInode = 256
)
