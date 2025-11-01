package bootsnap

import (
	"os"
	"slices"
	"strings"

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
		RunE: func(cmd *cobra.Command, args []string) error {
			kernelParams, err := readKernelParams()
			if err != nil {
				return err
			}

			if force || slices.Contains(kernelParams, name) {
				return run(mountpoint, kernelParams)
			}

			return nil
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

const (
	name              = "bootsnap"
	defaultMountpoint = "/btrfs-pool"
)
