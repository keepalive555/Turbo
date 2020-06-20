package util

import (
	"os"
)

const (
	TurboDebug = "TURBO_DEBUG"
)

func IsDebugg() bool {
	env := os.Getenv(TurboDebug)
	if env == "" || env == "off" {
		return false
	}
	return true
}
