package main

import (
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/russross/blackfriday/v2"
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

	// Convert from markdown to html.
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return c.String(http.StatusNotFound, "File not found:" + filename)
	}

	output := blackfriday.Run(bytes)

	c.Response().Header().Add("Content-Type", "text/html; charset=UTF-8")
	return c.String(http.StatusOK, string(output))
}
