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
	e.Static("/static", "static")
	e.Logger.Fatal(e.Start(":8080"))
}

func handle(c echo.Context) error {
	filename := c.Param("name")

	// Default handler for root element.
	if filename == "" {
		filename = "index.html"
	}

	// If a file with this name exists, simply deliver it.
	bytes, err := ioutil.ReadFile(filename)
	if err == nil {
		return c.File(filename)
	}

	// Markdown target name.
	filename = strings.Replace(filename, ".html", ".md", 1)

	// Convert from markdown to html.
	bytes, err = ioutil.ReadFile(filename)
	if err != nil {
		return c.String(http.StatusNotFound, "File not found:"+filename)
	}

	params := blackfriday.HTMLRendererParameters{
		CSS:   "static/main.css",
		Flags: blackfriday.CompletePage,
	}
	renderer := blackfriday.NewHTMLRenderer(params)

	output := blackfriday.Run(bytes, blackfriday.WithRenderer((renderer)))

	c.Response().Header().Add("Content-Type", "text/html; charset=UTF-8")
	return c.String(http.StatusOK, string(output))
}
