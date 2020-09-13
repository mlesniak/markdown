package handler

import (
	"bytes"
	"github.com/labstack/echo/v4"
	"github.com/mlesniak/markdown/internal/cache"
	"github.com/russross/blackfriday/v2"
	"io/ioutil"
	"os"
	"strings"
)

// TODO Fix internal flag: this is ugly for such a general function
func (h *Handler) RenderFile(log echo.Logger, internal bool, filename string, data []byte) (string, bool) {
	// Are we allowed to display this file?
	if !internal && !isPublic(data) {
		// We use the same error message to prevent
		// guessing non-accessible filenames.
		log.Infof("File not public accessible: %s", filename)
		return "", true
	}

	// Perform various pre-processing steps on the markdown.
	markdown := processRawMarkdown(h.Tags, filename, data)
	titleLine := computeTitle(markdown)

	// Convert from (processed) markdown to html.
	renderer := blackfriday.NewHTMLRenderer(blackfriday.HTMLRendererParameters{})
	html := string(blackfriday.Run([]byte(markdown), blackfriday.WithRenderer(renderer)))

	// Inject rendered html into template and fill variables.
	// If we'll have more variables we'd use proper templating.
	bsTemplate, err := ioutil.ReadFile("template.html")
	if err != nil {
		log.Warn("Template not found. This should never happen.")
		return "", true
	}
	html = strings.ReplaceAll(string(bsTemplate), "${content}", html)
	html = strings.ReplaceAll(html, "${title}", titleLine)
	html = strings.ReplaceAll(html, "${build}", buildInformation())

	// Add to cache.
	if !internal {
		h.Cache.Add(cache.Entry{
			Name: filename,
			Data: []byte(html),
		})
	}

	return html, false
}

func buildInformation() string {
	buildInformation := os.Getenv("COMMIT")
	buildInformation = strings.Trim(buildInformation, " \n")
	if buildInformation == "" {
		buildInformation = "not available"
	}
	return buildInformation
}

// isPublic checks if a file is allowed to be displayed: Since we are only
// downloading markdown files, we enforce that all files must contain the tag
// `publishTag` to be able to download it.
func isPublic(bs []byte) bool {
	return bytes.Contains(bs, []byte(publishTag))
}
