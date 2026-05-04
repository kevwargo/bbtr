package watcher

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"regexp"

	"github.com/kevwargo/btrscr/internal/btrfs"
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

	recvCmd := exec.Command("btrfs", "receive", "--dump")
	recvCmd.Stdout = FilterLines(os.Stdout, "^(write|update_extent)")
	recvCmd.Stderr = os.Stderr
	stream, err := recvCmd.StdinPipe()
	if err != nil {
		return err
	}

	sendCmd := exec.Command("btrfs", "send", "-p", snapBefore, "--no-data", snapAfter)
	sendCmd.Stderr = os.Stderr
	sendCmd.Stdout = stream

	if err := sendCmd.Start(); err != nil {
		return err
	}
	log.Printf("Send command %q started (%d)", sendCmd.String(), sendCmd.Process.Pid)

	if err := recvCmd.Start(); err != nil {
		return err
	}
	log.Printf("Receive command %q started (%d)", recvCmd.String(), recvCmd.Process.Pid)

	var errs []error
	errs = append(errs, sendCmd.Wait())
	log.Println("send finished")
	errs = append(errs, stream.Close())
	log.Println("pipe closed")
	errs = append(errs, recvCmd.Wait())
	log.Println("recv finished")

	return errors.Join(errs...)
}

type lineFilter struct {
	out     io.Writer
	rx      *regexp.Regexp
	linebuf []byte
}

func FilterLines(out io.Writer, regex string) io.Writer {
	return &lineFilter{
		out: out,
		rx:  regexp.MustCompile(regex),
	}
}

func (f *lineFilter) Write(buf []byte) (int, error) {
	var total int
	f.linebuf = append(f.linebuf, buf...)
	lines := f.linebuf
	for len(lines) > 0 {
		nl := bytes.IndexByte(lines, '\n')
		if nl < 0 {
			break
		}

		line := lines[:nl+1]
		if f.rx.Match(line) {
			n, err := f.out.Write(line)
			total += n
			if err != nil {
				return total, err
			}
		}

		lines = lines[nl+1:]
	}

	copy(f.linebuf, lines)
	f.linebuf = f.linebuf[:len(lines)]

	return len(buf), nil
}
