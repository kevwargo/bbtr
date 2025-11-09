package btrfs

import (
	"fmt"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"syscall"
	"time"

	"github.com/kevwargo/bootsnap/internal/log"
	"golang.org/x/sys/unix"
)

const SnapshotFormat = "20060102-150405"

type Pool struct {
	AllSnapshotNames []string
	Subvols          []Subvol

	mountpoint           string
	needUnmount          bool
	needRemoveMountpoint bool
}

type Subvol struct {
	Name          string
	Path          string
	SnapshotPaths map[string]string
}

func Open(dev, mountpoint string) (_ *Pool, err error) {
	p := Pool{mountpoint: mountpoint}

	defer func() {
		if err != nil {
			p.Close()
		}
	}()

	if err = p.mount(dev); err != nil {
		return nil, err
	}

	var (
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

func (p *Pool) Close() {
	if p.needUnmount {
		cmd := exec.Command("umount", p.mountpoint)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		log.Println(cmd)

		if err := cmd.Run(); err != nil {
			log.Printf("unmounting %s: %s", p.mountpoint, err.Error())
		}
	}

	if p.needRemoveMountpoint {
		if err := os.Remove(p.mountpoint); err != nil {
			log.Printf("removing %s: %s", p.mountpoint, err.Error())
		} else {
			log.Printf("removed %s", p.mountpoint)
		}
	}
}

func (p *Pool) Table() map[string][]string {
	table := make(map[string][]string)

	for _, subvol := range p.Subvols {
		for snapshot := range subvol.SnapshotPaths {
			table[subvol.Name] = append(table[subvol.Name], snapshot)
		}
	}

	return table
}

func isSubvol(path string) bool {
	stat := statDir(path)
	if stat == nil {
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

func (p *Pool) mount(dev string) error {
	if isSubvol(p.mountpoint) {
		return nil
	}

	if statDir(p.mountpoint) == nil {
		if err := os.MkdirAll(p.mountpoint, 0o755); err != nil {
			return fmt.Errorf("creating mountpoint %q: %w", p.mountpoint, err)
		}

		p.needRemoveMountpoint = true
	}

	cmd := exec.Command("mount", "-t", "btrfs", dev, p.mountpoint)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	log.Println(cmd)

	if err := cmd.Run(); err != nil {
		return err
	}

	p.needUnmount = true

	return nil
}

func statDir(path string) os.FileInfo {
	stat, err := os.Stat(path)
	if err != nil || !stat.IsDir() {
		return nil
	}

	return stat
}

const btrfsSubvolInode = 256
