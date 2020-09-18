package markdown

import (
	"os"
	"strings"
)

func buildInformation() string {
	buildInformation := os.Getenv("COMMIT")
	buildInformation = strings.Trim(buildInformation, " \n")
	if buildInformation == "" {
		buildInformation = "not available"
	}
	return buildInformation
}
