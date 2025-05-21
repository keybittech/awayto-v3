package api

import (
	"time"
)

const (
	maxCacheBuffer        = 1 << 12
	cacheKeySuffixModTime = ":mod"
)

var (
	duration180, _ = time.ParseDuration("180s")
)
