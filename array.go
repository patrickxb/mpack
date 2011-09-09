package mpack

import (
        "bytes"
        "reflect"
)

type Array struct {
        raw []interface{}
}

func NewArray(generic interface{}) *Array {
        result := new(Array)
        result.raw = generic.([]interface{})
        return result
}

func NewEmptyArray() *Array {
        result := new(Array)
        result.raw = make([]interface{}, 0)
        return result
}

func (a Array) Len() int {
        return len(a.raw)
}

func (a Array) Raw() []interface{} {
        return a.raw
}

func (a Array) Item(index int) interface{} {
        return a.raw[index]
}

func (a Array) IntItem(index int) int64 {
        v := reflect.ValueOf(a.raw[index])
        if isInt(v) {
                return v.Int()
        }
        if isUint(v) {
                return int64(v.Uint())
        }
        return 0
}

func (a Array) UintItem(index int) uint64 {
        v := reflect.ValueOf(a.raw[index])
        if isUint(v) {
                return v.Uint()
        }
        if isInt(v) {
                return uint64(v.Int())
        }
        return 0
}

func (a Array) Uint32Item(index int) uint32 {
        v := reflect.ValueOf(a.raw[index])
        if isUint(v) {
                return uint32(v.Uint())
        }
        if isInt(v) {
                return uint32(v.Int())
        }
        return 0
}

func (a Array) StringItem(index int) string {
        return string(a.raw[index].([]uint8))
}

func (a Array) BufferItem(index int) *bytes.Buffer {
        return bytes.NewBuffer(a.raw[index].([]byte))
}

func (a Array) ArrayItem(index int) *Array {
        return NewArray(a.raw[index])
}

func (a Array) MapItem(index int) *Map {
        return NewMap(a.raw[index])
}

func (a Array) FloatItem(index int) float64 {
        v := reflect.ValueOf(a.raw[index])
        if isFloat(v) {
                return v.Float()
        }
        if isInt(v) {
                return float64(v.Int())
        }
        if isUint(v) {
                return float64(v.Uint())
        }
        return 0.0
}

func (a *Array) Append(item interface{}) {
        a.raw = append(a.raw, item)
}
