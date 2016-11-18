package log

import (
	"sync"
)

const maxCap = 2 << 16

var bytesPool = sync.Pool{
	New: func() interface{} {
		return []byte{}
	},
}

func getBytes() (b []byte) {
	ifc := bytesPool.Get()
	if ifc != nil {
		b = ifc.([]byte)
	}
	return
}

func putBytes(b []byte) {
	if cap(b) <= maxCap {
		b = b[:0]
		bytesPool.Put(b)
	}
}
