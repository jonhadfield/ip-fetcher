package main

import (
	"os"
	"strings"
)

const envEnabledValue = "true"

func isEnvEnabled(key string) bool {
	return strings.EqualFold(os.Getenv(key), envEnabledValue)
}
