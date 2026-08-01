package cmd

import (
	"fmt"

	"github.com/kevwargo/bbtr/internal/btrfs"
	"github.com/spf13/cobra"
)

func Execute() error {
	var cfg config

	cmd := &cobra.Command{
		Use:           "bbtr [flags] SUBVOL",
		SilenceErrors: true,
		SilenceUsage:  true,
		Args:          cobra.MinimumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) (err error) {
			for _, subvolPath := range args {
				cfg.subvol, err = btrfs.OpenSubvol(subvolPath)
				if err != nil {
					return err
				}

				if err = run(cfg); err != nil {
					return err
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&cfg.dest, "dest-dir", "d", "", "A directory to store encrypted and compressed snapshots")
	cmd.Flags().BoolVarP(&cfg.dryRun, "dry-run", "n", false, "Print commands instead of executing them")

	cmd.MarkFlagRequired("dest-dir")

	return cmd.Execute()
}

type config struct {
	subvol *btrfs.Subvol
	dest   string
	dryRun bool
}

func run(cfg config) error {
	fmt.Printf("%s (%s)\n", cfg.subvol.Name, cfg.subvol.Path)
	for name, snap := range cfg.subvol.SnapshotPaths {
		fmt.Printf(" snap %s (%s)\n", name, snap)
	}

	return nil
}
