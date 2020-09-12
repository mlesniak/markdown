package handler

import (
	"github.com/labstack/echo/v4"
	"os"
	"strings"
)

func BuildVersionHeader() func(next echo.HandlerFunc) echo.HandlerFunc {
	commit := BuildInformation()

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Add("X-Version", commit)
			return next(c)
		}
	}
}

// BuildInformation returns a string with commit hash and build time.
func BuildInformation() string {
	commit := os.Getenv("COMMIT")
	commit = strings.Trim(commit, " \n")
	if commit == "" {
		commit = "not available"
	}

	return commit
}
