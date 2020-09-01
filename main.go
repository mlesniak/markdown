// TODO Refactor this...
// TODO Use echo logging
package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
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

	// Default handler for root element.
	if filename == "" {
		filename = "202009010520 index"
	}

	var bs []byte

	// Deliver non-markdown files from local directory.
	if !strings.HasSuffix(filename, ".md") {
		_, err := ioutil.ReadFile(filename)
		if err == nil {
			return c.File(filename)
		}
	}

	// If file does not end with .html, append it.
	if !strings.HasSuffix(filename, ".html") {
		filename = filename + ".html"
	}
	filename = strings.Replace(filename, ".html", ".md", 1)

	// Convert from markdown to html.
	xs, err := readFromDropbox(filename)
	bs = xs
	if err != nil {
		return c.String(http.StatusNotFound, "File not found:"+filename)
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
	title := "mlesniak.com"
	if len(submatch) > 0 {
		title = submatch[1]
	}
	outstr = strings.ReplaceAll(outstr, "${title}", title)

	// Format date

	// Remove all found tags #\w+\S

	// Handle wiki-Links
	regex, err = regexp.Compile(`\[\[(.*?)\]\]`)
	if err != nil {
		panic(err)
	}
	submatches := regex.FindAllStringSubmatch(outstr, 10)
	for _, matches := range submatches {
		if len(matches) < 2 {
			continue
		}
		m := matches[1]
		link := strings.SplitN(m, " ", 2)[1]
		link = fmt.Sprintf(`<a href="%s">%s</a>`, matches[1], link)
		outstr = strings.ReplaceAll(outstr, matches[0], link)
	}

	return c.String(http.StatusOK, outstr)
}

func readFromDropbox(filename string) ([]byte, error) {
	client := http.Client{}
	request, err := http.NewRequest("POST", "https://content.dropboxapi.com/2/files/download", nil)
	if err != nil {
		panic(err)
	}
	token := os.Getenv("TOKEN")
	request.Header.Add("Authorization", "Bearer "+token)
	request.Header.Add("Dropbox-API-Arg", "{\"path\": \"/notes/"+filename+"\"}")
	resp, err := client.Do(request)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	all, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	isPublic := bytes.Contains(all, []byte("#public"))
	if !isPublic {
		println("File not public: " + filename)
		return nil, errors.New("content is not public")
	}

	return all, err

}
