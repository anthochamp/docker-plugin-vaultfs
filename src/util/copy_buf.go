// from github.com/docker/go-plugins-helpers/sdk/pool.go (Apache 2.0 licence)

package util

import (
	"io"
	"sync"
)

const buffer32K = 32 * 1024

var buffer32KPool = &sync.Pool{New: func() interface{} { return make([]byte, buffer32K) }}

// copyBuf uses a shared buffer pool with io.CopyBuffer
func CopyBuf(w io.Writer, r io.Reader) (int64, error) {
	buf := buffer32KPool.Get().([]byte)
	written, err := io.CopyBuffer(w, r, buf)
	buffer32KPool.Put(buf)
	return written, err
}
