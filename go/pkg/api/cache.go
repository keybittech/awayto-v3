package api

import (
	"os"
	"time"
)

const (
	maxCacheBuffer        = 1 << 12
	cacheKeySuffixData    = ":data"
	cacheKeySuffixModTime = ":mod"
)

var (
	apiPathLen     = len(os.Getenv("API_PATH"))
	duration180, _ = time.ParseDuration("180s")
)
