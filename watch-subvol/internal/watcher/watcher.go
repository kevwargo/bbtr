package watcher

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/kevwargo/btrscr/internal/btrfs"
	"github.com/kevwargo/btrscr/internal/btrfs/stream"
)

func Watch(subvol string) error {
	s, err := btrfs.OpenSubvol(subvol)
	if err != nil {
		return err
	}

	snapBefore, err := s.BackupNow()
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "Snapshot %s created. Press Enter to check diff...", snapBefore)
	os.Stdin.Read([]byte{0})

	snapAfter, err := s.BackupNow()
	if err != nil {
		return err
	}

	return Diff(snapBefore, snapAfter)
}

func Diff(snapBefore, snapAfter string) (finalErr error) {
	sendCmd := exec.Command("btrfs", "send", "-p", snapBefore, "--no-data", snapAfter)

	var sendErrBuf bytes.Buffer
	sendCmd.Stderr = &sendErrBuf

	sendReader, err := sendCmd.StdoutPipe()
	if err != nil {
		return err
	}

	if err := sendCmd.Start(); err != nil {
		return err
	}

	defer func() {
		errs := []error{finalErr}

		errs = append(errs, sendReader.Close())
		if we := sendCmd.Wait(); we != nil {
			errs = append(errs, fmt.Errorf("%q %w: %s", sendCmd.String(), we, sendErrBuf.String()))
		}

		finalErr = errors.Join(errs...)
	}()

	printed := make(map[string]struct{})
	for cmd, err := range stream.Parse(sendReader) {
		if err != nil {
			return err
		}

		if cmd.Type != stream.C_UPDATE_EXTENT {
			continue
		}

		if a := cmd.FindAttr(stream.A_PATH); a != nil {
			path := string(a.Value)
			if _, ok := printed[path]; !ok {
				fmt.Println(string(a.Value))
				printed[path] = struct{}{}
			}
		}
	}

	return nil
}
