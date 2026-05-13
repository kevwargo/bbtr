package btrfs

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"golang.org/x/sys/unix"
)

const SnapshotFormat = "20060102-150405Z"

type Pool struct {
	Subvols []Subvol

	mountpoint           string
	needUnmount          bool
	needRemoveMountpoint bool
}

type Subvol struct {
	Name          string
	Path          string
	SnapshotPaths map[string]string

	snapDir string
}

func OpenSubvol(subvolMountpoint string) (*Subvol, error) {
	var err error
	subvolMountpoint, err = filepath.Abs(subvolMountpoint)
	if err != nil {
		return nil, err
	}

	mountinfos, err := readMountinfos()
	if err != nil {
		return nil, err
	}

	var subvolMountinfo *mountinfo
	for _, mi := range mountinfos {
		if mi.Mountpoint == subvolMountpoint {
			if mi.FSType != "btrfs" {
				return nil, fmt.Errorf("%s is not btrfs subvolume", subvolMountpoint)
			}
			if mi.FSRoot == "/" {
				return nil, fmt.Errorf("%s is a btrfs root", subvolMountpoint)
			}

			subvolMountinfo = &mi
			break
		}
	}

	if subvolMountinfo == nil {
		return nil, fmt.Errorf("%s is not a valid mountpoint", subvolMountpoint)
	}
	if !strings.HasPrefix(subvolMountinfo.FSRoot, "/@") {
		return nil, fmt.Errorf("%s is not a conventional subvolume", subvolMountpoint)
	}
	subvolName := subvolMountinfo.FSRoot[2:]

	var pool string
	for _, mi := range mountinfos {
		if mi.Dev == subvolMountinfo.Dev && mi.FSRoot == "/" {
			pool = mi.Mountpoint
			break
		}
	}

	if pool == "" {
		return nil, fmt.Errorf("root of %s is not mounted", subvolMountpoint)
	}

	return &Subvol{
		Path:    pool + subvolMountinfo.FSRoot,
		Name:    subvolName,
		snapDir: filepath.Join(pool, "s", subvolName),
	}, nil
}

func OpenPool(dev, mountpoint string) (*Pool, error) {
	var (
		p       = Pool{mountpoint: mountpoint}
		empty   = true
		success bool
	)

	defer func() {
		if !success {
			p.Close()
		}
	}()

	if err := p.mount(dev); err != nil {
		return nil, err
	}

	snapDir := filepath.Join(mountpoint, "s")

	entries, err := os.ReadDir(snapDir)
	if err != nil {
		return nil, fmt.Errorf("listing dir %q: %w", snapDir, err)
	}

	for _, e := range entries {
		subvol := Subvol{
			Name:          e.Name(),
			Path:          filepath.Join(mountpoint, "@"+e.Name()),
			SnapshotPaths: make(map[string]string),

			snapDir: filepath.Join(snapDir, e.Name()),
		}

		snapEntries, err := os.ReadDir(subvol.snapDir)
		if err != nil {
			return nil, fmt.Errorf("listing dir %q: %w", subvol.snapDir, err)
		}

		for _, se := range snapEntries {
			snapName := se.Name()
			snapPath := filepath.Join(subvol.snapDir, snapName)

			if !isSubvol(snapPath) {
				continue
			}

			if _, err := time.Parse(SnapshotFormat, snapName); err != nil {
				continue
			}

			subvol.SnapshotPaths[snapName] = snapPath
			empty = false
		}

		p.Subvols = append(p.Subvols, subvol)
	}

	if empty {
		return nil, fmt.Errorf("mountpoint %s does not contain valid snapshots", mountpoint)
	}

	success = true

	return &p, nil
}

func (p *Pool) Close() {
	if p.needUnmount {
		if err := runCmd("umount", p.mountpoint); err != nil {
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

func (s Subvol) Backup(snapName string) (string, error) {
	path := filepath.Join(s.snapDir, snapName)
	if err := runCmd("btrfs", "subvolume", "snapshot", "-r", s.Path, path); err != nil {
		return "", err
	}

	return path, nil
}

func (s Subvol) BackupNow() (string, error) {
	return s.Backup(time.Now().UTC().Format(SnapshotFormat))
}

func (s Subvol) Restore(snapshot string) error {
	path, ok := s.SnapshotPaths[snapshot]
	if !ok {
		log.Printf("Ignoring restore request for non-existent snapshot %q for %s", snapshot, s.Name)

		return nil
	}

	if err := runCmd("btrfs", "subvolume", "delete", "--commit-after", s.Path); err != nil {
		return err
	}

	return runCmd("btrfs", "subvolume", "snapshot", path, s.Path)
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
	if err := unix.Statfs(path, &statfs); err != nil {
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

	if err := runCmd("mount", "-t", "btrfs", dev, p.mountpoint); err != nil {
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

func runCmd(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = log.Writer()
	cmd.Stderr = log.Writer()

	log.Println(cmd)

	return cmd.Run()
}

const btrfsSubvolInode = 256
