package main

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()
	e.GET("/", handle)
	e.GET("/:name", handle)
	e.Logger.Fatal(e.Start(":8080"))
}

func handle(c echo.Context) error {
	filename := c.Param("name")

	// Default handler for root element.
	if filename == "" {
		filename = "index.html"
	}

	// Markdown target name.
	filename = strings.Replace(filename, ".html", ".md", 1)

	return c.String(http.StatusOK, filename)
}
