package api

import (
	"bytes"
	"os"
	"sync"
	"time"
)

const (
	maxCacheBuffer        = 1 << 12
	cacheKeySuffixData    = ":data"
	cacheKeySuffixModTime = ":mod"
)

var (
	apiPathLen                = len(os.Getenv("API_PATH"))
	duration180, _            = time.ParseDuration("180s")
	cacheMiddlewareBufferPool = sync.Pool{
		New: func() interface{} {
			var buf bytes.Buffer
			return &buf
		},
	}
)
