package d2objects

import (
	"errors"
	"fmt"
	"io"
)

// A File is the holder of the Class Definition and contains
// every object
type File struct {
	r           Reader
	startOffset int64
	indexes     map[int32]int64
	classes     map[int32]*ClassDefinition
}

type ClassDefinition struct {
	namespace string
	name      string
	fields    []*FieldDefinition
}

// These are the possible GameDataType
const (
	_                     = iota
	GameDataTypeInt int32 = -iota
	GameDataTypeBoolean
	GameDataTypeString
	GameDataTypeNumber
	GameDataTypeI18n
	GameDataTypeUint
	GameDataTypeVector = -99
)

type readFn func(Reader, *File, int) (interface{}, error)
type innerType struct {
	name     string
	dataType int32
	fn       readFn
}

// A FieldDefinition is the definition of a single field.
// It consits of a name, a datatype and maybe inner types (for Vectors)
type FieldDefinition struct {
	name     string
	dataType int32

	fn readFn

	// this has to be a slice of pointers because of the "manipulations" we do in getReadMethod()
	innerTypes []*innerType
}

// ErrInvalidD2oHeader means that the d2o header is malformed
var ErrInvalidD2oHeader = errors.New("invalid d2o header")

// ParseFile parses a d2o File
func ParseFile(r Reader) (*File, error) {
	f := &File{
		r:       r,
		indexes: map[int32]int64{},
		classes: map[int32]*ClassDefinition{},
	}
	if err := f.parse(); err != nil {
		return nil, fmt.Errorf("d2objects: %v", err)
	}
	return f, nil
}

func (f *File) parse() error {
	startOffset, err := f.parseHeader()
	if err != nil {
		return err
	}
	f.startOffset = startOffset + 7
	indexOffset, err := f.r.ReadInt32()
	if err != nil {
		return err
	}
	if err = f.r.Goto(startOffset + int64(indexOffset)); err != nil {
		return err
	}
	if err = f.parseIndexes(startOffset); err != nil {
		return err
	}
	if err = f.parseClassDefinitions(); err != nil {
		return err
	}
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

func (f *File) parseIndexes(startOffset int64) error {
	indexesSize, err := f.r.ReadInt32()
	if err != nil {
		return err
	}
	for offset := int32(0); offset < indexesSize; offset += 8 {
		index, err := f.r.ReadInt32()
		if err != nil {
			return err
		}
		value, err := f.r.ReadInt32()
		if err != nil {
			return err
		}
		f.indexes[index] = int64(value) + startOffset
	}
	return nil
}

func (f *File) parseClassDefinitions() error {
	defCounts, err := f.r.ReadInt32()
	if err != nil {
		return err
	}
	for i := int32(0); i < defCounts; i++ {
		id, err := f.r.ReadInt32()
		if err != nil {
			return err
		}
		class, err := f.parseClassDefinition()
		if err != nil {
			return err
		}
		f.classes[id] = class
	}
	return nil
}

func (f *File) parseClassDefinition() (*ClassDefinition, error) {
	class := &ClassDefinition{}
	if err := class.parse(f.r); err != nil {
		return nil, err
	}
	return class, nil
}

func (c *ClassDefinition) parse(r Reader) (err error) {
	c.namespace, err = r.ReadString()
	if err != nil {
		return
	}
	c.name, err = r.ReadString()
	if err != nil {
		return
	}
	fieldCount, err := r.ReadInt32()
	if err != nil {
		return
	}
	c.fields = make([]*FieldDefinition, fieldCount)
	for i := range c.fields {
		f := &FieldDefinition{}
		if err = f.parse(r); err != nil {
			return
		}
		c.fields[i] = f
	}
	return
}

func (f *FieldDefinition) parse(r Reader) (err error) {
	f.name, err = r.ReadString()
	if err = f.parseType(r); err != nil {
		return
	}
	return nil
}

func (f *FieldDefinition) parseType(r Reader) (err error) {
	f.dataType, err = r.ReadInt32()
	if err != nil {
		return err
	}
	f.fn, f.innerTypes, err = f.getReadMethod(r, f.dataType, nil)
	return nil
}

// getReadMethod returns the read method to used depending on the dataType
// it gets pretty ugly because of vectors that have inner types.
// each inner type has to recurse to get its read method
func (f *FieldDefinition) getReadMethod(r Reader, dataType int32, innerTypes []*innerType) (readFn, []*innerType, error) {
	switch dataType {
	case GameDataTypeInt:
		return f.readInt, innerTypes, nil
	case GameDataTypeBoolean:
		return f.readBoolean, innerTypes, nil
	case GameDataTypeString:
		return f.readString, innerTypes, nil
	case GameDataTypeNumber:
		return f.readNumber, innerTypes, nil
	case GameDataTypeI18n:
		return f.readI18n, innerTypes, nil
	case GameDataTypeUint:
		return f.readUint, innerTypes, nil
	case GameDataTypeVector:
		var err error
		inner := &innerType{}
		inner.name, err = r.ReadString()
		if err != nil {
			return nil, nil, err
		}
		inner.dataType, err = r.ReadInt32()
		if err != nil {
			return nil, nil, err
		}
		innerTypes = append(innerTypes, inner)
		inner.fn, innerTypes, err = f.getReadMethod(r, inner.dataType, innerTypes)
		if err != nil {
			return nil, nil, err
		}
		return f.readVector, innerTypes, nil
	}
	if dataType > 0 {
		return f.readObject, innerTypes, nil
	}
	return nil, nil, fmt.Errorf("invalid data type %v", dataType)
}

// GetObjects retrieves all objects of the File
func (f *File) GetObjects() ([]map[string]interface{}, error) {
	count := len(f.indexes)
	if err := f.r.Goto(f.startOffset); err != nil {
		return nil, err
	}
	objects := make([]map[string]interface{}, count)
	for i := range objects {
		objectOffset, err := f.r.ReadInt32()
		if err != nil {
			return nil, err
		}
		object, err := f.classes[objectOffset].Read(f.r, f)
		if err != nil {
			return nil, err
		}
		objects[i] = object
	}
	return objects, nil
}

// GetObject retrieves a single object
func (f *File) GetObject(id int32) (map[string]interface{}, error) {
	index, found := f.indexes[id]
	if !found {
		return nil, fmt.Errorf("index for %v not found", id)
	}
	if err := f.r.Goto(index); err != nil {
		return nil, err
	}
	classDefIndex, err := f.r.ReadInt32()
	if err != nil {
		return nil, err
	}

	classDefinition, found := f.classes[classDefIndex]
	if !found {
		return nil, fmt.Errorf("class definition for %v not found (idx: %v)", id, classDefIndex)
	}
	v, err := classDefinition.Read(f.r, f)
	if err != nil {
		return nil, fmt.Errorf("%v: %v", id, err)
	}
	return v, nil
}

func (c *ClassDefinition) Read(r Reader, f *File) (map[string]interface{}, error) {
	fields := map[string]interface{}{}
	for _, fieldDef := range c.fields {
		v, err := fieldDef.Read(r, f)
		if err != nil {
			return nil, err
		}
		fields[fieldDef.name] = v
	}
	return fields, nil
}

func (f *FieldDefinition) Read(r Reader, file *File) (interface{}, error) {
	return f.fn(r, file, 0)
}

func (f *FieldDefinition) readInt(r Reader, _ *File, _ int) (interface{}, error) {
	return r.ReadInt32()
}

func (f *FieldDefinition) readBoolean(r Reader, _ *File, _ int) (interface{}, error) {
	return r.ReadBoolean()
}

func (f *FieldDefinition) readString(r Reader, _ *File, _ int) (interface{}, error) {
	return r.ReadString()
}

func (f *FieldDefinition) readNumber(r Reader, _ *File, _ int) (interface{}, error) {
	return r.ReadDouble()
}

func (f *FieldDefinition) readI18n(r Reader, _ *File, _ int) (interface{}, error) {
	return r.ReadInt32()
}

func (f *FieldDefinition) readUint(r Reader, _ *File, _ int) (interface{}, error) {
	return r.ReadUInt32()
}

func (f *FieldDefinition) readVector(r Reader, file *File, offset int) (interface{}, error) {
	count, err := r.ReadInt32()
	if err != nil {
		return nil, err
	}
	values := make([]interface{}, count)

	inner := f.innerTypes[offset]
	for i := range values {
		v, err := inner.fn(r, file, offset+1)
		if err != nil {
			return nil, err
		}
		values[i] = v
	}
	return values, nil
}

func (f *FieldDefinition) readObject(r Reader, file *File, _ int) (interface{}, error) {
	id, err := r.ReadInt32()
	if err != nil {
		return nil, err
	} else if id == -1431655766 {
		return nil, nil
	}
	return file.GetObject(id)
}
