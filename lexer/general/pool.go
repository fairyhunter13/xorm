package general

import (
	"strings"
	"sync"
)

var (
	stringsBuilderPool = &sync.Pool{
		New: func() interface{} {
			return new(strings.Builder)
		},
	}
)

// GetBuilder return the string builder inside this pool.
func GetBuilder() *strings.Builder {
	builder, ok := stringsBuilderPool.Get().(*strings.Builder)
	if !ok {
		builder = new(strings.Builder)
	}
	builder.Reset()
	return builder
}

// PutBuilder return the strings builder inside the pool.
func PutBuilder(builder *strings.Builder) {
	stringsBuilderPool.Put(builder)
}
