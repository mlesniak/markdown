package utils

import (
	"os"
	"strings"
)

func BuildInformation() string {
	buildInformation := os.Getenv("COMMIT")
	buildInformation = strings.Trim(buildInformation, " \n")
	if buildInformation == "" {
		buildInformation = "not available"
	}
	return buildInformation
}
