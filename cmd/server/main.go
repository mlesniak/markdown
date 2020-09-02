package main

import (
	"bytes"
	"github.com/mlesniak/markdown/pkg/dropbox"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/russross/blackfriday/v2"
)

const (
	// Default title if the title can not be extracted from the markdown file.
	defaultTitle = "mlesniak.com"

	// Name of the root file if no filename is specified.
	rootFilename = "202009010520 index"

	// Tag name to define markdown files which are allowed to be published.
	publishTag = "#public"
)

var dropboxService *dropbox.Service

func init() {
	dropboxToken := os.Getenv("TOKEN")
	if dropboxToken == "" {
		panic("No dropbox token set, aborting.")
	}
	dropboxService = dropbox.New(dropboxToken)
}

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

// handle is the default handler for all non-static content. It uses the parameter name
// to download the correct markdown file from dropbox, perform various transformations
// and convert it to html.
func handle(c echo.Context) error {
	// Prepare filename.
	filename := c.Param("name")
	filename = fixFilename(filename)

	// Read file from dropbox storage.
	bs, err := dropboxService.Read(filename)
	if err != nil {
		return c.String(http.StatusNotFound, "File not found:"+filename)
	}

	// Are we allowed to display this file?
	if !isPublicFile(bs) {
		return c.String(http.StatusNotFound, "File not found:"+filename)
	}

	// Perform various pre-processing steps on the markdown.
	markdown := processRawMarkdown(bs)
	titleLine := computeTitle(markdown)

	// Convert from (processed) markdown to html.
	renderer := blackfriday.NewHTMLRenderer(blackfriday.HTMLRendererParameters{})
	html := string(blackfriday.Run([]byte(markdown), blackfriday.WithRenderer(renderer)))

	// Inject rendered html into template and fill variables.
	// If we'll have more variables we'd use proper templating.
	bsTemplate, err := ioutil.ReadFile("template.html")
	if err != nil {
		return c.String(http.StatusInternalServerError, "Template not found. This should never happen.")
	}
	html = strings.ReplaceAll(string(bsTemplate), "${content}", html)
	html = strings.ReplaceAll(html, "${title}", titleLine)

	// Return generated HTML file with correct content type.
	c.Response().Header().Add("Content-Type", "text/html; charset=UTF-8")
	return c.String(http.StatusOK, html)
}

// isPublicFile checks if a file is allowed to be displayed: Since we are only
// downloading markdown files, we enforce that all files must contain the tag
// `publishTag` to be able to download it.
func isPublicFile(bs []byte) bool {
	isPublic := bytes.Contains(bs, []byte(publishTag))
	if !isPublic {
		return false
	}

	return true
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
