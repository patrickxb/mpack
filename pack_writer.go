package mpack

import (
	"io"
	"os"
	"encoding/binary"
	"reflect"
	"fmt"
)

type PackWriter struct {
	writer io.Writer
}

func NewPackWriter(writer io.Writer) *PackWriter {
	result := new(PackWriter)
	result.writer = writer
	return result
}

func (pw PackWriter) writeCode(code byte) (int, os.Error) {
	return pw.writer.Write([]byte{code})
}

func (pw PackWriter) writeByte(b byte) (int, os.Error) {
	return pw.writer.Write([]byte{b})
}

func (pw PackWriter) writeBinary(n interface{}) os.Error {
	return binary.Write(pw.writer, binary.BigEndian, n)
}

func (pw PackWriter) writeBlock(code byte, n interface{}, bytes int) (int, os.Error) {
	total, err := pw.writeCode(code)
	if err != nil {
		return 0, err
	}
	err = pw.writeBinary(n)
	if err != nil {
		return total, err
	}
	total += bytes
	return total, nil
}

func (pw PackWriter) packUint8(n uint8) (int, os.Error) {
	if n < 128 {
		return pw.writeByte(n)
	}
	return pw.writer.Write([]byte{type_uint8, byte(n)})
}

func (pw PackWriter) packByte(n byte) (int, os.Error) {
	return pw.packUint8(n)
}

func (pw PackWriter) packUint16(n uint16) (int, os.Error) {
	if n < 0x100 {
		return pw.packUint8(uint8(n))
	}
	return pw.writeBlock(type_uint16, n, 2)
}

func (pw PackWriter) packUint32(n uint32) (int, os.Error) {
	if n < 65536 {
		return pw.packUint16(uint16(n))
	}
	return pw.writeBlock(type_uint32, n, 4)
}

func (pw PackWriter) packUint64(n uint64) (int, os.Error) {
	if n < 4294967296 {
		return pw.packUint32(uint32(n))
	}
	return pw.writeBlock(type_uint64, n, 8)
}

func (pw PackWriter) packInt8(n int8) (int, os.Error) {
	if n > 0 {
		return pw.writer.Write([]byte{byte(n)})
	} else if n >= -32 {
		return pw.writer.Write([]byte{byte(n)})
	}
	return pw.writer.Write([]byte{type_int8, byte(n)})
}

func (pw PackWriter) packInt16(n int16) (int, os.Error) {
	if n >= -128 && n <= 127 {
		return pw.packInt8(int8(n))
	}
	return pw.writeBlock(type_int16, n, 2)
}

func (pw PackWriter) packInt32(n int32) (int, os.Error) {
	if n >= -32768 && n <= 32767 {
		return pw.packInt16(int16(n))
	}
	return pw.writeBlock(type_int32, n, 4)
}

func (pw PackWriter) packInt64(n int64) (int, os.Error) {
	if n >= -2147483648 && n <= 2147483647 {
		return pw.packInt32(int32(n))
	}
	return pw.writeBlock(type_int64, n, 8)
}

func (pw PackWriter) packNil() (int, os.Error) {
	return pw.writeCode(type_nil)
}

func (pw PackWriter) packBool(v bool) (int, os.Error) {
	b := type_true
	if v == false {
		b = type_false
	}
	return pw.writeByte(b)
}

func (pw PackWriter) packFloat32(n float32) (int, os.Error) {
	return pw.writeBlock(type_float, n, 4)
}

func (pw PackWriter) packFloat64(n float64) (int, os.Error) {
	return pw.writeBlock(type_double, n, 8)
}

func (pw PackWriter) packBytes(b []byte) (int, os.Error) {
	if len(b) < 32 {
		pw.writeByte(type_fix_raw | uint8(len(b)))
		pw.writer.Write(b)
		return 1 + len(b), nil
	} else if len(b) < 65536 {
		pw.writeCode(type_raw16)
		var length uint16 = uint16(len(b))
		pw.writeBinary(length)
		pw.writer.Write(b)
		return 3 + len(b), nil
	} else if uint32(len(b)) <= uint32(4294967295) {
		pw.writeCode(type_raw32)
		length := uint32(len(b))
		pw.writeBinary(length)
		pw.writer.Write(b)
		return 5 + len(b), nil
	}
	return 0, nil
}

func (pw PackWriter) packString(s string) (int, os.Error) {
	return pw.packBytes([]byte(s))
}

func (pw PackWriter) packInt64Array(a []int64) (int, os.Error) {
	numBytes := 0
	if len(a) < 16 {
		n, err := pw.writeCode(type_fix_array_min | uint8(len(a)))
		if err != nil {
			return numBytes, err
		}
		numBytes += n
	} else if len(a) < 65536 {
		n, err := pw.writeCode(type_array16)
		if err != nil {
			return numBytes, err
		}
		numBytes += n
		length := uint16(len(a))
		err = pw.writeBinary(length)
		if err != nil {
			return numBytes, err
		}
		numBytes += 2
	} else if uint32(len(a)) <= uint32(4294967295) {
		n, err := pw.writeCode(type_array32)
		if err != nil {
			return numBytes, err
		}
		numBytes += n
		length := uint32(len(a))
		err = pw.writeBinary(length)
		if err != nil {
			return numBytes, err
		}
		numBytes += 4
	}

	for i := 0; i < len(a); i++ {
		n, err := pw.packInt64(a[i])
		if err != nil {
			return numBytes, err
		}
		numBytes += n
	}
	return numBytes, nil
}

func (pw PackWriter) packArray(a reflect.Value) (int, os.Error) {
	numBytes := 0
	if a.Len() < 16 {
		n, err := pw.writeCode(type_fix_array_min | uint8(a.Len()))
		if err != nil {
			return numBytes, err
		}
		numBytes += n
	} else if a.Len() < 65536 {
		n, err := pw.writeCode(type_array16)
		if err != nil {
			return numBytes, err
		}
		numBytes += n
		length := uint16(a.Len())
		err = pw.writeBinary(length)
		if err != nil {
			return numBytes, err
		}
		numBytes += 2
	} else if uint32(a.Len()) <= uint32(4294967295) {
		n, err := pw.writeCode(type_array32)
		if err != nil {
			return numBytes, err
		}
		numBytes += n
		length := uint32(a.Len())
		err = pw.writeBinary(length)
		if err != nil {
			return numBytes, err
		}
		numBytes += 4
	}
	for i := 0; i < a.Len(); i++ {
		elt := a.Index(i)
		n, err := pw.pack(elt.Interface())
		if err != nil {
			return numBytes, err
		}
		numBytes += n
	}
	return numBytes, nil
}

func (pw PackWriter) packMap(m reflect.Value) (int, os.Error) {
	numBytes := 0

	if m.Len() < 16 {
		n, err := pw.writeCode(type_fix_map_min | uint8(m.Len()))
		if err != nil {
			return numBytes, err
		}
		numBytes += n
	} else if m.Len() < 65536 {
		n, err := pw.writeCode(type_map16)
		if err != nil {
			return numBytes, err
		}
		numBytes += n
		length := uint16(m.Len())
		err = pw.writeBinary(length)
		if err != nil {
			return numBytes, err
		}
		numBytes += 2
	} else if uint32(m.Len()) <= uint32(4294967295) {
		n, err := pw.writeCode(type_map32)
		if err != nil {
			return numBytes, err
		}
		numBytes += n
		length := uint32(m.Len())
		err = pw.writeBinary(length)
		if err != nil {
			return numBytes, err
		}
		numBytes += 4
	}

	keys := m.MapKeys()
	for i := 0; i < len(keys); i++ {
		n, err := pw.pack(keys[i].Interface())
		if err != nil {
			return numBytes, err
		}
		numBytes += n
		n, err = pw.pack(m.MapIndex(keys[i]).Interface())
		if err != nil {
			return numBytes, err
		}
		numBytes += n
	}

	return numBytes, nil
}

func (pw PackWriter) pack(value interface{}) (int, os.Error) {
	if value == nil {
		return pw.packNil()
	}
	switch tvalue := value.(type) {
	case int8:
		return pw.packInt8(tvalue)
	case int16:
		return pw.packInt16(tvalue)
	case int32:
		return pw.packInt32(tvalue)
	case int64:
		return pw.packInt64(tvalue)
	case int:
		return pw.packInt64(int64(tvalue))
	case uint8:
		return pw.packUint8(tvalue)
	case uint16:
		return pw.packUint16(tvalue)
	case uint32:
		return pw.packUint32(tvalue)
	case uint64:
		return pw.packUint64(tvalue)
	case uint:
		return pw.packUint64(uint64(tvalue))
	case bool:
		return pw.packBool(tvalue)
	case float32:
		return pw.packFloat32(tvalue)
	case float64:
		return pw.packFloat64(tvalue)
	case []byte:
		return pw.packBytes(tvalue)
	case []int64:
		return pw.packInt64Array(tvalue)
	case string:
		return pw.packString(tvalue)
	}

	// see if it is an array...
	rvalue := reflect.ValueOf(value)
	if rvalue.Kind() == reflect.Array || rvalue.Kind() == reflect.Slice {
		return pw.packArray(rvalue)
	}

	// is it a map?
	if rvalue.Kind() == reflect.Map {
		return pw.packMap(rvalue)
	}

	fmt.Printf("unknown type: %s\n", value)
	return 0, nil
}
