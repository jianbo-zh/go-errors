package errors

import (
	"bytes"
	"sync"
)

var (
	innerSeparator = []byte(": ")
	groupSeparator = []byte("; ")

	// Prefix for multi-line messages
	multilinePrefix    = []byte("the following errors occurred:")
	multilineSeparator = []byte("\n -  ")
	multilineIndent    = []byte("    ")
)

// _bufferPool is a pool of bytes.Buffers.
var bufferPool = sync.Pool{
	New: func() interface{} {
		return &bytes.Buffer{}
	},
}
