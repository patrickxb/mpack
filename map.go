package mpack

import (
        "log"
        "reflect"
)

type Map struct {
        raw map[interface{}]interface{}
}

func NewMap(generic interface{}) *Map {
        result := new(Map)
        result.raw = generic.(map[interface{}]interface{})
        return result
}

func isInt(v reflect.Value) bool {
        k := v.Kind()
        return k == reflect.Int || k == reflect.Int8 || k == reflect.Int16 || k == reflect.Int32 || k == reflect.Int64
}

func isUint(v reflect.Value) bool {
        k := v.Kind()
        return k == reflect.Uint || k == reflect.Uint8 || k == reflect.Uint16 || k == reflect.Uint32 || k == reflect.Uint64
}

func isFloat(v reflect.Value) bool {
        k := v.Kind()
        return k == reflect.Float32 || k == reflect.Float64
}

func (m Map) IntIndex(key interface{}) (int64, bool) {
        index, present := m.raw[key]
        if !present {
                return 0, false
        }
        v := reflect.ValueOf(index)
        if isInt(v) {
                return v.Int(), true
        }
        if isUint(v) {
                return int64(v.Uint()), true
        }

        return 0, false
}

func (m Map) IntPlainIndex(key interface{}) (int, bool) {
        index, present := m.raw[key]
        if !present {
                return 0, false
        }
        v := reflect.ValueOf(index)
        if isInt(v) {
                return int(v.Int()), true
        }
        if isUint(v) {
                return int(v.Uint()), true
        }

        return 0, false
}

func (m Map) Int32Index(key interface{}) (int32, bool) {
        index, present := m.raw[key]
        if !present {
                return 0, false
        }
        v := reflect.ValueOf(index)
        if isInt(v) {
                return int32(v.Int()), true
        }
        if isUint(v) {
                return int32(v.Uint()), true
        }

        return 0, false
}

func (m Map) UintIndex(key interface{}) (uint64, bool) {
        index, present := m.raw[key]
        if !present {
                return 0, false
        }
        v := reflect.ValueOf(index)
        if isUint(v) {
                return v.Uint(), true
        }
        if isInt(v) {
                return uint64(v.Int()), true
        }
        return 0, false
}

func (m Map) UintSmallIndex(key interface{}) (uint, bool) {
        index, present := m.raw[key]
        if !present {
                return 0, false
        }
        v := reflect.ValueOf(index)
        if isUint(v) {
                return uint(v.Uint()), true
        }
        if isInt(v) {
                return uint(v.Int()), true
        }
        return 0, false
}

func (m Map) FloatIndex(key interface{}) (float64, bool) {
        index, present := m.raw[key]
        if !present {
                return 0, false
        }
        v := reflect.ValueOf(index)
        if isFloat(v) {
                return v.Float(), true
        }
        if isInt(v) {
                return float64(v.Int()), true
        }
        if isUint(v) {
                return float64(v.Uint()), true
        }

        return 0, false
}

func (m Map) StringIndex(key interface{}) (string, bool) {
        index, present := m.raw[key]
        if !present {
                return "", false
        }
        v := reflect.ValueOf(index)
        if v.IsValid() == false {
                return "", true
        }
        if v.IsNil() {
                return "", true
        }
        return string(index.([]uint8)), true
}

func (m Map) ArrayIndex(key interface{}) (*Array, bool) {
        index, present := m.raw[key]
        if !present {
                return nil, false
        }
        return NewArray(index), true
}

func (m Map) MapIndex(key interface{}) (*Map, bool) {
        index, present := m.raw[key]
        if !present {
                return nil, false
        }
        v := reflect.ValueOf(index)
        if v.IsValid() == false {
                return nil, true
        }
        if v.IsNil() {
                return nil, true
        }
        return NewMap(index), true
}

func (m Map) DumpKeys() {
        for k := range m.raw {
                log.Printf(k.(string))
        }
}
