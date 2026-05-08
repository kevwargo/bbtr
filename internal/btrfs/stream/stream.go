package stream

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"iter"
)

const magic = "btrfs-stream\x00"

type header struct {
	Magic   [len(magic)]byte
	Version uint32
}

type CmdHdr struct {
	Length uint32
	Type   uint16
	CRC    uint32
}

type Command struct {
	CmdHdr
	Attrs []Attribute
}

func (c Command) FindAttr(aType int) *Attribute {
	for _, a := range c.Attrs {
		if a.Type == uint16(aType) {
			return &a
		}
	}

	return nil
}

type AttrHdr struct {
	Type   uint16
	Length uint16
}

type Attribute struct {
	AttrHdr
	Value []byte
}

func Parse(r io.Reader) iter.Seq2[*Command, error] {
	s := state{r: r}
	return s.iterate
}

type state struct {
	r io.Reader
}

func (s state) iterate(yield func(*Command, error) bool) {
	if err := s.verifyHeader(); err != nil {
		if !errors.Is(err, io.EOF) {
			yield(nil, err)
		}

		return
	}

	for {
		cmd, err := s.readCommand()
		if err != nil {
			yield(nil, err)

			return
		}

		if cmd != nil {
			if !yield(cmd, nil) {
				return
			}
		} else {
			return
		}
	}
}

func (s state) verifyHeader() error {
	var hdr header
	if err := binary.Read(s.r, binary.LittleEndian, &hdr); err != nil {
		return err
	}

	if string(hdr.Magic[:]) != magic {
		return fmt.Errorf("magic %v does not match expected %v", hdr.Magic, []byte(magic))
	}

	switch hdr.Version {
	case 1, 2:
		return nil
	default:
		return fmt.Errorf("invalid btrfs-stream version: %d, must be 1 or 2", hdr.Version)
	}
}

func (s state) readCommand() (*Command, error) {
	var cmd Command
	if err := binary.Read(s.r, binary.LittleEndian, &cmd.CmdHdr); err != nil {
		if errors.Is(err, io.EOF) {
			return nil, nil
		}

		return nil, fmt.Errorf("reading command header: %w", err)
	}

	if cmd.Length == 0 {
		return &cmd, nil
	}

	payload := make([]byte, cmd.Length)
	if _, err := io.ReadFull(s.r, payload); err != nil {
		return nil, fmt.Errorf("reading command payload: %w", err)
	}

	ar := bytes.NewBuffer(payload)
	for {
		var attr Attribute
		if err := binary.Read(ar, binary.LittleEndian, &attr.AttrHdr); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return nil, fmt.Errorf("reading attribute header: %w", err)
		}

		attr.Value = make([]byte, attr.Length)
		if _, err := io.ReadFull(ar, attr.Value); err != nil {
			return nil, fmt.Errorf("reading attribute value: %w", err)
		}

		cmd.Attrs = append(cmd.Attrs, attr)
	}

	return &cmd, nil
}
