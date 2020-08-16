package main

import (
	"io/ioutil"
	"net/http"
	"regexp"
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

	// Special case handler for ToC.
	if filename == "toc.html" {

	} else {
		// Convert from markdown to html.
		bytes, err = ioutil.ReadFile(filename)
		if err != nil {
			return c.String(http.StatusNotFound, "File not found:"+filename)
		}
	}

	params := blackfriday.HTMLRendererParameters{
		CSS: "static/main.css",
		Flags: // blackfriday.CompletePage |
		// TODO Won't work like this
		blackfriday.SmartypantsQuotesNBSP |
			blackfriday.SmartypantsDashes |
			blackfriday.SmartypantsLatexDashes,
	}
	renderer := blackfriday.NewHTMLRenderer(params)

	output := blackfriday.Run(bytes, blackfriday.WithRenderer(renderer))

	outstr := string(output)

	// Add meta directive for better mobile rendering.
	// We should add a patch to blackfriday to inject it as part of complete-page-rendering.
	// outstr = strings.ReplaceAll(outstr,
	// 	`<meta charset="utf-8">`,
	// 	`<meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1" />`)

	c.Response().Header().Add("Content-Type", "text/html; charset=UTF-8")

	btempl, err := ioutil.ReadFile("template.html")
	if err != nil {
		return c.String(http.StatusNotFound, "Template not found:"+filename)
	}
	outstr = strings.ReplaceAll(string(btempl), "${content}", outstr)
	regex, err := regexp.Compile("<h1>(.*)</h1>")
	if err != nil {
		panic(err)
	}
	submatch := regex.FindStringSubmatch(outstr)
	outstr = strings.ReplaceAll(outstr, "${title}", submatch[1])

	return c.String(http.StatusOK, outstr)
}

// TODO Collect referer
// TODO TOC
