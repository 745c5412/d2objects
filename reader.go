package d2objects

import "io"

import "encoding/binary"

// Reader is the minimal interface required to deserialiaze Dofus 2 i18n files
type Reader interface {
	io.ReadSeeker
	ReadBoolean() (bool, error)
	ReadInt8() (int8, error)
	ReadUInt8() (uint8, error)
	ReadInt16() (int16, error)
	ReadUInt16() (uint16, error)
	ReadInt32() (int32, error)
	ReadUInt32() (uint32, error)
	ReadFloat() (float32, error)
	ReadDouble() (float64, error)
	ReadString() (string, error)

	Position() (int64, error)
	Goto(int64) error
}

type reader struct {
	r io.ReadSeeker
}

// NewReader creates a Reader
func NewReader(r io.ReadSeeker) Reader {
	return &reader{r}
}

func (r *reader) read(x interface{}) error {
	return binary.Read(r.r, binary.BigEndian, x)
}

func (r *reader) Read(p []byte) (n int, err error) {
	return r.r.Read(p)
}

func (r *reader) Seek(offset int64, whence int) (int64, error) {
	return r.r.Seek(offset, whence)
}

func (r *reader) Position() (int64, error) {
	p, err := r.Seek(0, io.SeekCurrent)
	if err != nil {
		return 0, err
	}
	return p, nil
}

func (r *reader) Goto(p int64) error {
	_, err := r.Seek(p, io.SeekStart)
	return err
}

func (r *reader) ReadBoolean() (bool, error) {
	b, err := r.ReadUInt8()
	if err != nil {
		return false, err
	}
	return b != 0, nil
}

func (r *reader) ReadInt8() (int8, error) {
	b, err := r.ReadUInt8()
	if err != nil {
		return 0, err
	}
	return int8(b), nil
}

func (r *reader) ReadUInt8() (uint8, error) {
	var b uint8
	if err := r.read(&b); err != nil {
		return 0, err
	}
	return b, nil
}

func (r *reader) ReadInt16() (int16, error) {
	var b int16
	if err := r.read(&b); err != nil {
		return 0, err
	}
	return b, nil
}

func (r *reader) ReadUInt16() (uint16, error) {
	var b uint16
	if err := r.read(&b); err != nil {
		return 0, err
	}
	return b, nil
}

func (r *reader) ReadInt32() (int32, error) {
	var b int32
	if err := r.read(&b); err != nil {
		return 0, err
	}
	return b, nil
}

func (r *reader) ReadUInt32() (uint32, error) {
	var b uint32
	if err := r.read(&b); err != nil {
		return 0, err
	}
	return b, nil
}

func (r *reader) ReadFloat() (float32, error) {
	var b float32
	if err := r.read(&b); err != nil {
		return 0, err
	}
	return b, nil
}

func (r *reader) ReadDouble() (float64, error) {
	var b float64
	if err := r.read(&b); err != nil {
		return 0, err
	}
	return b, nil
}

func (r *reader) ReadString() (string, error) {
	len, err := r.ReadUInt16()
	if err != nil {
		return "", err
	}

	buf := make([]byte, len)
	_, err = io.ReadFull(r.r, buf)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}
