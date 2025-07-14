package pool

import (
	"bytes"
	"sync"
)

var RLPBuffers = &sync.Pool{
	New: func() interface{} {
		return bytes.NewBuffer(make([]byte, 0, 2500)) // TODO revisit this value
	},
}

func GetRLPBuffer() *bytes.Buffer {
	return RLPBuffers.Get().(*bytes.Buffer)
}

func PutRLPBuffer(b *bytes.Buffer) {
	b.Reset()
	RLPBuffers.Put(b)
}
