package api

import (
	"time"
)

const (
	maxCacheBuffer = 1 << 12
)

var (
	duration180, _ = time.ParseDuration("180s")
)
