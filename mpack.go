package mpack

import (
        "os"
        "io"
        //	"time"
        //	"fmt"
)

func Pack(w io.Writer, value interface{}) (packedBytes int, err os.Error) {
        //	stime := time.Nanoseconds()
        pw := NewPackWriter(w)
        n, e := pw.pack(value)

        //	etime := time.Nanoseconds()
        //	msecs := (float64)(etime-stime) / 1000000
        //	fmt.Printf("pack time: %.3fms\n", msecs)

        return n, e
}

func Unpack(reader io.Reader) (interface{}, int, os.Error) {
        pr := NewPackReader(reader)
        return pr.unpack()
}
