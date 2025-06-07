package api

import (
	"time"
)

const (
	maxCacheBuffer = 1 << 12
)

var (
	duration180s, _   = time.ParseDuration("180s")
	duration86400s, _ = time.ParseDuration("86400s")
)
