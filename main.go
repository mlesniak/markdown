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

const rootFilename = "202009010520 index"

func main() {
	e := echo.New()

	// Serve static files.
	e.Static("/static", "static")

	// Serve dynamic files.
	e.GET("/", handle)
	e.GET("/:name", handle)

	// Start server.
	e.Logger.Fatal(e.Start(":8080"))
}

// handle is the default handler for all non-static content.
func handle(c echo.Context) error {
	// Prepare filename.
	filename := c.Param("name")
	filename = fixFilename(filename)

	var bs []byte
	// Convert from markdown to html.
	xs, err := readFromDropbox(filename)
	bs = xs
	if err != nil {
		return c.String(http.StatusNotFound, "File not found:"+filename)
	}

	outstr := string(bs)

	// Remove all tags
	regex := regexp.MustCompile("[\\s]#\\w+\\S")
	mtchs := regex.FindAllString(outstr, -1)
	for _, mtch := range mtchs {
		if mtch != "" {
			outstr = strings.ReplaceAll(outstr, mtch, "")
		}
	}

	params := blackfriday.HTMLRendererParameters{
		CSS: "static/main.css",
	}
	renderer := blackfriday.NewHTMLRenderer(params)
	output := blackfriday.Run([]byte(outstr), blackfriday.WithRenderer(renderer))
	outstr = string(output)

	c.Response().Header().Add("Content-Type", "text/html; charset=UTF-8")

	btempl, err := ioutil.ReadFile("template.html")
	if err != nil {
		return c.String(http.StatusNotFound, "Template not found:"+filename)
	}
	outstr = strings.ReplaceAll(string(btempl), "${content}", outstr)
	regex, err = regexp.Compile(`<h1>(.*)</h1>`)
	if err != nil {
		panic(err)
	}
	submatch := regex.FindStringSubmatch(outstr)
	title := "mlesniak.com"
	if len(submatch) > 0 {
		title = submatch[1]
	}
	outstr = strings.ReplaceAll(outstr, "${title}", title)

	// Handle wiki-Links
	regex, err = regexp.Compile(`\[\[(.*?)\]\]`)
	if err != nil {
		panic(err)
	}
	submatches := regex.FindAllStringSubmatch(outstr, -1)
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

// fixFilename transform the requested filename, i.e. redirects to
// index page or fixes simplified filenames without suffix.
func fixFilename(filename string) string {
	// If / is requested we redirect to our index page.
	if filename == "" {
		filename = rootFilename
	}

	// In our markup, wiki links have no markdown suffix.
	// Append suffix if not yet present.
	if !strings.HasSuffix(filename, ".md") {
		filename = filename + ".md"
	}

	return filename
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
