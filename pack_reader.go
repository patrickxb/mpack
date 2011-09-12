package mpack

import (
        "encoding/binary"
        "fmt"
        "io"
        "log"
        "os"
        "reflect"
)

/*
type FullReader interface {
	io.ByteReader
	io.Reader
}
*/

type PackReader struct {
        reader io.Reader
}

func NewPackReader(r io.Reader) *PackReader {
        result := new(PackReader)
        result.reader = r
        return result
}

func (pr PackReader) ReadByte() (byte, os.Error) {
        // return pr.reader.ReadByte()
        data := [1]byte{}
        n, err := pr.reader.Read(data[0:])
        if err != nil {
                log.Printf("packreader ReadByte got error (%d bytes): %s", n, err)
                return 0, err
        }
        if n != 1 {
                return 0, os.NewError("didn't read just one byte")
        }
        return data[0], nil
}

func (pr PackReader) ReadBinary(result interface{}) os.Error {
        return binary.Read(pr.reader, binary.BigEndian, result)
}

func (pr PackReader) unpackRaw(length uint32, prefixBytes int) (interface{}, int, os.Error) {
        if length == 0 {
                // lenght == 0 => nil...
                return nil, 0, nil
        }
        numRead := prefixBytes
        data := make([]byte, length)
        n, err := pr.reader.Read(data)
        numRead += n
        if err != nil {
                return nil, numRead, err
        }
        return data, numRead, nil
}

func (pr PackReader) unpackArray(length uint32, prefixBytes int) (interface{}, int, os.Error) {
        numRead := prefixBytes
        data := make([]interface{}, length)
        for i := uint32(0); i < length; i++ {
                elt, n, err := pr.unpack()
                numRead += n
                if err != nil {
                        return nil, numRead, err
                }
                data[i] = elt
        }
        return data, numRead, nil
}

func (pr PackReader) unpackMap(length uint32, prefixBytes int) (interface{}, int, os.Error) {
        numRead := prefixBytes

        m := make(map[interface{}]interface{})

        for i := uint32(0); i < length; i++ {
                key, n, err := pr.unpack()
                numRead += n
                if err != nil {
                        return nil, numRead, err
                }

                val, n, err := pr.unpack()
                numRead += n
                if err != nil {
                        return nil, numRead, err
                }

                if reflect.TypeOf(key).String() == "[]uint8" {
                        m[string(key.([]uint8))] = val
                } else {
                        m[key] = val
                }
        }

        return m, numRead, nil
}

func (pr PackReader) unpack() (interface{}, int, os.Error) {
        numRead := 0
        b, err := pr.ReadByte()
        if err != nil {
                return nil, numRead, err
        }
        numRead += 1

        // how is this possible?
        if b < 0 {
                return nil, numRead, nil
        }

        if b < positive_fix_max {
                return b, numRead, nil
        }
        if b >= negative_fix_min && b <= negative_fix_max {
                return (b & negative_fix_mask) - negative_fix_offset, numRead, nil
        }

        if b >= type_fix_raw && b <= type_fix_raw_max {
                return pr.unpackRaw(uint32(b&fix_raw_count_mask), numRead)
        }

        if b >= type_fix_array_min && b <= type_fix_array_max {
                return pr.unpackArray(uint32(b&fix_array_count_mask), numRead)
        }

        if b >= type_fix_map_min && b <= type_fix_map_max {
                return pr.unpackMap(uint32(b&fix_map_count_mask), numRead)
        }

        switch b {
        case type_nil:
                return nil, numRead, nil
        case type_false:
                return false, numRead, nil
        case type_true:
                return true, numRead, nil
        case type_uint8:
                c, err := pr.ReadByte()
                if err != nil {
                        return nil, numRead + 1, err
                }
                return uint8(c), numRead + 1, nil
        case type_uint16:
                var result uint16
                err := pr.ReadBinary(&result)
                if err != nil {
                        return nil, numRead + 2, err
                }
                return result, numRead + 2, nil
        case type_uint32:
                var result uint32
                err := pr.ReadBinary(&result)
                if err != nil {
                        return nil, numRead + 4, err
                }
                return result, numRead + 4, nil
        case type_uint64:
                var result uint64
                err := pr.ReadBinary(&result)
                if err != nil {
                        return nil, numRead + 8, err
                }
                return result, numRead + 8, nil
        case type_int16:
                var result int16
                err := pr.ReadBinary(&result)
                if err != nil {
                        return nil, numRead + 2, err
                }
                return result, numRead + 2, nil
        case type_int32:
                var result int32
                err := pr.ReadBinary(&result)
                if err != nil {
                        return nil, numRead + 4, err
                }
                return result, numRead + 4, nil
        case type_int64:
                var result int64
                err := pr.ReadBinary(&result)
                if err != nil {
                        return nil, numRead + 8, err
                }
                return result, numRead + 8, nil
        case type_float:
                var result float32
                err := pr.ReadBinary(&result)
                if err != nil {
                        return nil, numRead + 4, err
                }
                return result, numRead + 4, nil
        case type_double:
                var result float64
                err := pr.ReadBinary(&result)
                if err != nil {
                        return nil, numRead + 8, err
                }
                return result, numRead + 8, nil
        case type_raw16:
                var length uint16
                err := pr.ReadBinary(&length)
                numRead += 2
                if err != nil {
                        return nil, numRead, err
                }
                return pr.unpackRaw(uint32(length), numRead)
        case type_raw32:
                var length uint32
                err := pr.ReadBinary(&length)
                numRead += 4
                if err != nil {
                        return nil, numRead, err
                }
                return pr.unpackRaw(length, numRead)
        case type_array16:
                var length uint16
                err := pr.ReadBinary(&length)
                numRead += 2
                if err != nil {
                        return nil, numRead, err
                }
                return pr.unpackArray(uint32(length), numRead)
        case type_array32:
                var length uint32
                err := pr.ReadBinary(&length)
                numRead += 2
                if err != nil {
                        return nil, numRead, err
                }
                return pr.unpackArray(length, numRead)
        case type_map16:
                var length uint16
                err := pr.ReadBinary(&length)
                numRead += 2
                if err != nil {
                        return nil, numRead, err
                }
                return pr.unpackMap(uint32(length), numRead)
        case type_map32:
                var length uint32
                err := pr.ReadBinary(&length)
                numRead += 4
                if err != nil {
                        return nil, numRead, err
                }
                return pr.unpackMap(length, numRead)
        default:
                fmt.Printf("unhandled type prefix: %x\n", b)
        }

        return b, numRead, nil
}
