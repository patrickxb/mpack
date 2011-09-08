package mpack

const (
	type_nil             byte = 0xc0
	type_false           byte = 0xc2
	type_true            byte = 0xc3
	type_float           byte = 0xca
	type_double          byte = 0xcb
	type_uint8           byte = 0xcc
	type_uint16          byte = 0xcd
	type_uint32          byte = 0xce
	type_uint64          byte = 0xcf
	type_int8            byte = 0xd0
	type_int16           byte = 0xd1
	type_int32           byte = 0xd2
	type_int64           byte = 0xd3
	type_raw16           byte = 0xda
	type_raw32           byte = 0xdb
	type_array16         byte = 0xdc
	type_array32         byte = 0xdd
	type_map16           byte = 0xde
	type_map32           byte = 0xdf
	type_fix_raw         byte = 0xa0
	type_fix_raw_max     byte = 0xbf
	type_fix_array_min   byte = 0x90
	type_fix_array_max   byte = 0x9f
	type_fix_map_min     byte = 0x80
	type_fix_map_max     byte = 0x8f
	fix_array_count_mask byte = 0xf
	fix_map_count_mask   byte = 0xf
	fix_raw_count_mask   byte = 0x1f
	negative_fix_min     byte = 0xe0
	negative_fix_max     byte = 0xff
	negative_fix_mask    byte = 0x1f
	negative_fix_offset  byte = 0x20
	positive_fix_max     byte = 0x7f
	rpc_request          byte = 0x00
	rpc_response         byte = 0x01
	rpc_notify           byte = 0x02
)
