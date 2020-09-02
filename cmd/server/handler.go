package main

import (
	"bytes"
	"github.com/labstack/echo/v4"
	"github.com/russross/blackfriday/v2"
	"io/ioutil"
	"net/http"
	"strings"
)

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
		// We use the same error message to prevent
		// guessing non-accessible filenames.
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
	return bytes.Contains(bs, []byte(publishTag))
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