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
	println("request for " + filename)

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
	// If file does not end with .html, append it.
	if !strings.HasSuffix(filename, ".html") {
		filename = filename + ".html"
	}
	filename = strings.Replace(filename, ".html", ".md", 1)

	// Convert from markdown to html.
	bytes, err = ioutil.ReadFile(filename)
	if err != nil {
		return c.String(http.StatusNotFound, "File not found:"+filename)
	}

	params := blackfriday.HTMLRendererParameters{
		CSS: "static/main.css",
		Flags: blackfriday.CompletePage |
			blackfriday.SmartypantsQuotesNBSP |
			blackfriday.SmartypantsDashes |
			blackfriday.SmartypantsLatexDashes,
	}
	renderer := blackfriday.NewHTMLRenderer(params)

	output := blackfriday.Run(bytes, blackfriday.WithRenderer(renderer))

	outstr := string(output)

	// Add meta directive for better mobile rendering.
	// We should add a patch to blackfriday to inject it as part of complete-page-rendering.
	outstr = strings.ReplaceAll(outstr,
		`<meta charset="utf-8">`,
		`<meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1" />`)

	c.Response().Header().Add("Content-Type", "text/html; charset=UTF-8")
	return c.String(http.StatusOK, outstr)
}
