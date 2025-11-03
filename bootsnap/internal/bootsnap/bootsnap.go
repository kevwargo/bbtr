package bootsnap

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/kevwargo/bootsnap/internal/btrfs"
	"github.com/spf13/cobra"
)

func Execute() error {
	var (
		force      bool
		mountpoint string
	)

	cmd := &cobra.Command{
		Use:           name,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(_ *cobra.Command, _ []string) error {
			return execute(mountpoint, force)
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force bootsnap menu even without the appropriate kernel param")
	cmd.Flags().StringVarP(&mountpoint, "mountpoint", "m", defaultMountpoint, "Mountpoint for BTRFS pool")

	return cmd.Execute()
}

func readKernelParams() ([]string, error) {
	data, err := os.ReadFile("/proc/cmdline")
	if err != nil {
		return nil, err
	}

	return strings.Split(string(data), " "), nil
}

func execute(mountpoint string, force bool) error {
	kernelParams, err := readKernelParams()
	if err != nil {
		return err
	}

	if !force && !slices.Contains(kernelParams, name) {
		return nil
	}

	rootDev, err := findRootDev(kernelParams)
	if err != nil {
		return err
	}

	pool, err := btrfs.Open(rootDev, mountpoint)
	if err != nil {
		return err
	}
	defer pool.Close()

	return runMenu(pool)
}

func findRootDev(kernelParams []string) (string, error) {
	for _, param := range kernelParams {
		if val, ok := strings.CutPrefix(param, rootParam); ok && val != "" {
			return val, nil
		}
	}

	return "", fmt.Errorf("%q* kernel param not found", rootParam)
}

const (
	name              = "bootsnap"
	defaultMountpoint = "/btrfs-pool"
	rootParam         = "root="
)
