package btrfs

import (
	"fmt"
	"log"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"syscall"
	"time"

	"golang.org/x/sys/unix"
)

const SnapshotFormat = "20060102-150405"

type Pool struct {
	AllSnapshotNames []string
	Subvols          []Subvol
}

type Subvol struct {
	Name          string
	Path          string
	SnapshotPaths map[string]string
}

func Mount(dev, mountpoint string) (*Pool, error) {
	if err := mount(dev, mountpoint); err != nil {
		return nil, err
	}

	var (
		p                Pool
		allSnapshotNames = make(map[string]struct{})
		rootSnapDir      = filepath.Join(mountpoint, "s")
	)

	entries, err := os.ReadDir(rootSnapDir)
	if err != nil {
		return nil, fmt.Errorf("listing dir %q: %w", rootSnapDir, err)
	}

	for _, e := range entries {
		subvol := Subvol{
			Name:          e.Name(),
			Path:          filepath.Join(mountpoint, "@"+e.Name()),
			SnapshotPaths: make(map[string]string),
		}

		snapDir := filepath.Join(rootSnapDir, subvol.Name)
		snapEntries, err := os.ReadDir(snapDir)
		if err != nil {
			return nil, fmt.Errorf("listing dir %q: %w", snapDir, err)
		}

		for _, se := range snapEntries {
			snapName := se.Name()
			snapPath := filepath.Join(snapDir, snapName)

			if !isSubvol(snapPath) {
				continue
			}

			if _, err := time.Parse(SnapshotFormat, snapName); err != nil {
				continue
			}

			subvol.SnapshotPaths[snapName] = snapPath
			allSnapshotNames[snapName] = struct{}{}
		}

		p.Subvols = append(p.Subvols, subvol)
	}

	p.AllSnapshotNames = slices.Collect(maps.Keys(allSnapshotNames))
	slices.Sort(p.AllSnapshotNames)

	return &p, nil
}

func isSubvol(path string) bool {
	stat, err := os.Stat(path)
	if err != nil {
		return false
	}

	if !stat.IsDir() {
		return false
	}

	if sys, _ := stat.Sys().(*syscall.Stat_t); sys == nil || sys.Ino != btrfsSubvolInode {
		return false
	}

	var statfs unix.Statfs_t
	if unix.Statfs(path, &statfs) != nil {
		return false
	}

	return statfs.Type == unix.BTRFS_SUPER_MAGIC
}

func mount(dev, mountpoint string) error {
	if isSubvol(mountpoint) {
		return nil
	}

	if err := os.MkdirAll(mountpoint, 0o755); err != nil {
		return fmt.Errorf("creating mountpoint %q: %w", mountpoint, err)
	}

	cmd := exec.Command("mount", "-t", "btrfs", dev, mountpoint)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	log.Println(cmd)

	return cmd.Run()
}

const btrfsSubvolInode = 256
