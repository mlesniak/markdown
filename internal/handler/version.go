package handler

import (
	"github.com/labstack/echo/v4"
	"github.com/mlesniak/markdown/internal/utils"
)

func BuildVersionHeader() func(next echo.HandlerFunc) echo.HandlerFunc {
	commit := utils.BuildInformation()

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Add("X-Version", commit)
			return next(c)
		}
	}
}
