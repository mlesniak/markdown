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

const (
	defaultTitle = "mlesniak.com"
	rootFilename = "202009010520 index"
)

func main() {
	e := echo.New()

	// Serve static files.
	e.Static("/static", "static")

	// Serve dynamic files.
	e.GET("/", handle)
	e.GET("/:name", handle)

	// Start server.
	e.HideBanner = true
	e.HidePort = true
	e.Logger.Fatal(e.Start(":8080"))
}

// handle is the default handler for all non-static content.
func handle(c echo.Context) error {
	// Prepare filename.
	filename := c.Param("name")
	filename = fixFilename(filename)

	// Read file from dropbox storage.
	bs, err := readFromDropbox(filename)
	if err != nil {
		return c.String(http.StatusNotFound, "File not found:"+filename)
	}

	// Compute title from html.
	titleLine := computeTitle(bs)

	// Convert to html.
	renderer := blackfriday.NewHTMLRenderer(blackfriday.HTMLRendererParameters{})
	html := string(blackfriday.Run(bs, blackfriday.WithRenderer(renderer)))

	// Remove all tags.
	regex := regexp.MustCompile(`[\s]?#\w+`)
	matches := regex.FindAllString(html, -1)
	for _, mtch := range matches {
		if mtch != "" {
			html = strings.ReplaceAll(html, mtch, "")
		}
	}

	// Inject rendered html into template and fill variables.
	// If we'll have more variables we'd use proper templating.
	bsTemplate, err := ioutil.ReadFile("template.html")
	if err != nil {
		return c.String(http.StatusInternalServerError, "Template not found. This should never happen.")
	}
	html = strings.ReplaceAll(string(bsTemplate), "${content}", html)
	html = strings.ReplaceAll(html, "${title}", titleLine)

	// Handle wiki-Links
	regex, err = regexp.Compile(`\[\[(.*?)\]\]`)
	if err != nil {
		panic(err)
	}
	submatches := regex.FindAllStringSubmatch(html, -1)
	for _, matches := range submatches {
		if len(matches) < 2 {
			continue
		}
		m := matches[1]
		link := strings.SplitN(m, " ", 2)[1]
		link = fmt.Sprintf(`<a href="%s">%s</a>`, matches[1], link)
		html = strings.ReplaceAll(html, matches[0], link)
	}

	c.Response().Header().Add("Content-Type", "text/html; charset=UTF-8")
	return c.String(http.StatusOK, html)
}

// computeTitle uses the first line in markdown as title if available and feasible.
// Otherwise, default title is used.
func computeTitle(bs []byte) string {
	titleLine := defaultTitle
	lines := strings.SplitN(string(bs), "\n", 2)
	if len(lines) > 0 {
		titleLine = lines[0]
		// titleLine = strings.ReplaceAll(titleLine, "#", "")
		titleLine = strings.Trim(titleLine, " #")
	}
	return titleLine
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
