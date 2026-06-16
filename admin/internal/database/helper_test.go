package database

import (
	"os"
	"strconv"
)

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" { return v }
	return def
}
func envInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		n, _ := strconv.Atoi(v)
		return n
	}
	return def
}
