package mpack

type Params struct {
	raw *Map
}

func NewParams(generic interface{}) *Params {
	result := new(Params)
	args := NewArray(generic)
	result.raw = NewMap(args.Item(0))
	return result
}

func (p Params) Version() uint64 {
	v, present := p.raw.UintIndex("version")
	if !present {
		return 0
	}
	return v
}

// XXX this is pretty dumb...should be able to embed Map in this type???

func (p Params) IntIndex(key interface{}) (int64, bool) {
	return p.raw.IntIndex(key)
}

func (p Params) UintIndex(key interface{}) (uint64, bool) {
	return p.raw.UintIndex(key)
}

func (p Params) FloatIndex(key interface{}) (float64, bool) {
	return p.raw.FloatIndex(key)
}

func (p Params) StringIndex(key interface{}) (string, bool) {
	return p.raw.StringIndex(key)
}

func (p Params) ArrayIndex(key interface{}) (*Array, bool) {
	return p.raw.ArrayIndex(key)
}
