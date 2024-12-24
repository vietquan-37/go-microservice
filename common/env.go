package commons

import (
	"syscall"
)

func EnvString(key string, fallback string) string {
	if val, ok := syscall.Getenv(key); ok {
		return val
	}
	return fallback
}
