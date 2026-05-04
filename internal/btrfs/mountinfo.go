package btrfs

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type mountinfo struct {
	FSType     string
	Dev        string
	FSRoot     string
	Mountpoint string
}

func readMountinfos() ([]mountinfo, error) {
	var mountinfos []mountinfo

	f, err := os.Open("/proc/self/mountinfo")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		mi, err := parseMountinfo(sc.Text())
		if err != nil {
			return nil, err
		}

		mountinfos = append(mountinfos, mi)
	}

	return mountinfos, sc.Err()
}

func parseMountinfo(line string) (mountinfo, error) {
	fields := strings.Split(line, " ")
	mi := mountinfo{
		FSRoot:     fields[3],
		Mountpoint: fields[4],
	}
	for i := 6; i < len(fields); i++ {
		if fields[i] == "-" {
			if len(fields) <= i+2 {
				return mountinfo{}, fmt.Errorf("too few fields after separator: %q", line)
			}

			mi.FSType = fields[i+1]
			mi.Dev = fields[i+2]

			return mi, nil
		}
	}

	return mountinfo{}, fmt.Errorf("no separator: %q", line)
}
