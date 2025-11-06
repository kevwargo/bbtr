package bootsnap

import (
	"fmt"
	"os"
	"strings"

	"github.com/kevwargo/bootsnap/internal/btrfs"
)

func Execute() error {
	params, err := readParams()
	if err != nil {
		return err
	}

	if _, ok := params[paramEnabled]; !ok {
		return nil
	}

	rootDev := params[paramRoot]
	if rootDev == nil {
		return fmt.Errorf("kernel parameter %q not found", paramRoot)
	}

	mountpoint := defaultMountpoint
	if mntpt := params[paramMountpoint]; mntpt != nil {
		mountpoint = *mntpt
	}

	pool, err := btrfs.Open(*rootDev, mountpoint)
	if err != nil {
		return err
	}
	defer pool.Close()

	return runMenu(pool)
}

func readParams() (map[string]*string, error) {
	raw := os.Args[1:]
	if len(raw) == 0 {
		data, err := os.ReadFile(procCmdline)
		if err != nil {
			return nil, err
		}

		raw = strings.Split(strings.TrimSpace(string(data)), " ")
	}

	params := make(map[string]*string)

	for _, param := range raw {
		parts := strings.SplitN(param, "=", 2)

		if len(parts) == 1 {
			params[param] = nil
		} else {
			params[parts[0]] = &parts[1]
		}
	}

	return params, nil
}

const (
	name              = "bootsnap"
	defaultMountpoint = "/btrfs-pool"
	paramRoot         = "root"
	paramEnabled      = name + ".on"
	paramMountpoint   = name + ".mountpoint"
	procCmdline       = "/proc/cmdline"
)
