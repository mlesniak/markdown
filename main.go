package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"sort"
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
	referer := c.Request().Header.Get("Referer")
	println("referer: " + referer)

	filename := c.Param("name")
	println("request for " + filename)

	// Default handler for root element.
	if filename == "" {
		filename = "index.html"
	}

	var bs []byte

	// If a file with this name exists, simply deliver it.
	if !strings.HasSuffix(filename, ".md") {
		_, err := ioutil.ReadFile(filename)
		if err == nil {
			return c.File(filename)
		}
	}

	// Markdown target name.
	// If file does not end with .html, append it.
	if !strings.HasSuffix(filename, ".html") {
		filename = filename + ".html"
	}
	filename = strings.Replace(filename, ".html", ".md", 1)

	// Special case handler for ToC.
	if filename == "toc.md" {
		ignored := make(map[string]bool)
		ignored["template.html"] = true
		ignored["smashup.html"] = true

		// Generate markdown linked list
		dirs, err := ioutil.ReadDir(".")
		if err != nil {
			return c.String(http.StatusNotFound, "TOC not found:"+filename)
		}
		var b strings.Builder
		b.WriteString("<h1>List of all articles</h1>\n\n")

		names := make([]string, 0)
		for _, v := range dirs {
			if v.IsDir() {
				continue
			}
			if !strings.HasSuffix(v.Name(), ".md") {
				continue
			}
			names = append(names, v.Name())
		}
		sort.Strings(names)

		for _, name := range names {
			if _, found := ignored[name]; found {
				continue
			}
			href := strings.ReplaceAll(name, ".md", ".html")
			tbs, err := ioutil.ReadFile(name)
			if err != nil {
				println("Unable to read file for title:" + name)
				continue
			}
			title := strings.Trim(strings.Split(string(tbs), "\n")[0][1:], " ")
			b.WriteString(fmt.Sprintf("- [%v](%v)\n", title, href))
		}
		bs = []byte(b.String())
	} else {
		// Convert from markdown to html.
		xs, err := ioutil.ReadFile(filename)
		bs = xs
		if err != nil {
			return c.String(http.StatusNotFound, "File not found:"+filename)
		}
	}

	params := blackfriday.HTMLRendererParameters{
		CSS: "static/main.css",
	}
	renderer := blackfriday.NewHTMLRenderer(params)
	output := blackfriday.Run(bs, blackfriday.WithRenderer(renderer))
	outstr := string(output)

	c.Response().Header().Add("Content-Type", "text/html; charset=UTF-8")

	btempl, err := ioutil.ReadFile("template.html")
	if err != nil {
		return c.String(http.StatusNotFound, "Template not found:"+filename)
	}
	outstr = strings.ReplaceAll(string(btempl), "${content}", outstr)
	regex, err := regexp.Compile(`<h1>(.*)</h1>`)
	if err != nil {
		panic(err)
	}
	submatch := regex.FindStringSubmatch(outstr)
	outstr = strings.ReplaceAll(outstr, "${title}", submatch[1])

	return c.String(http.StatusOK, outstr)
}
