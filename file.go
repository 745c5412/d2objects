package d2objects

import (
	"errors"
	"io"
)

// A File is the holder of the Class Definition and contains
// every object
type File struct {
	r           Reader
	startOffset int64
}

// ErrInvalidD2oHeader means that the d2o header is malformed
var ErrInvalidD2oHeader = errors.New("invalid d2o header")

// ParseFile parses a d2o File
func ParseFile(r Reader) (*File, error) {
	f := &File{}
	if err := f.parse(); err != nil {
		return nil, err
	}
	return f, nil
}

func (f *File) parse() error {
	startOffset, err := f.parseHeader()
	if err != nil {
		return err
	}
	f.startOffset = startOffset + 7

	return nil
}

func (f *File) parseHeader() (int64, error) {
	var header [3]byte
	if _, err := io.ReadFull(f.r, header[:]); err != nil {
		return 0, err
	}
	if string(header[:]) != "D2O" {
		if err := f.r.Goto(0); err != nil {
			return 0, err
		}
		if s, err := f.r.ReadString(); err != nil {
			return 0, err
		} else if s != "AKSD" {
			return 0, ErrInvalidD2oHeader
		}
		if _, err := f.r.ReadInt16(); err != nil {
			return 0, err
		}
		offset, err := f.r.ReadInt32()
		if err != nil {
			return 0, err
		}
		if _, err = f.r.Seek(int64(offset), io.SeekCurrent); err != nil {
			return 0, err
		}
		pos, err := f.r.Position()
		if err != nil {
			return 0, err
		}
		if _, err := io.ReadFull(f.r, header[:]); err != nil {
			return 0, err
		} else if string(header[:]) != "D2O" {
			return 0, ErrInvalidD2oHeader
		}
		return pos, nil
	}
	return 0, nil
}
