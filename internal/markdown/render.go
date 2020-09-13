package markdown

import (
	"bytes"
	"errors"
	"github.com/labstack/echo/v4"
	"github.com/russross/blackfriday/v2"
	"io/ioutil"
	"os"
	"strings"
)

func ToHTML(log echo.Logger, filename string, data []byte) (string, error) {
	// Are we allowed to display this file?
	if !isPublic(data) {
		// We use the same error message to prevent
		// guessing non-accessible filenames.
		log.Infof("File not public accessible: %s", filename)
		return "", errors.New("file not public")
	}

	// Perform various pre-processing steps on the markdown.
	markdown := processRawMarkdown(data)
	titleLine := computeTitle(markdown)

	// Convert from (processed) markdown to html.
	renderer := blackfriday.NewHTMLRenderer(blackfriday.HTMLRendererParameters{})
	html := string(blackfriday.Run([]byte(markdown), blackfriday.WithRenderer(renderer)))

	// Inject rendered html into template and fill variables.
	// If we'll have more variables we'd use proper templating.
	bsTemplate, err := ioutil.ReadFile("template.html")
	if err != nil {
		log.Warn("Template not found. This should never happen.")
		return "", errors.New("template not found")
	}
	html = strings.ReplaceAll(string(bsTemplate), "${content}", html)
	html = strings.ReplaceAll(html, "${title}", titleLine)
	html = strings.ReplaceAll(html, "${build}", buildInformation())

	return html, nil
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

// computeTitle uses the first line in markdown as title if available and feasible.
// Otherwise, default title is used.
func computeTitle(markdown string) string {
	titleLine := defaultTitle
	lines := strings.SplitN(markdown, "\n", 2)
	if len(lines) > 0 {
		titleLine = lines[0]
		// titleLine = strings.ReplaceAll(titleLine, "#", "")
		titleLine = strings.Trim(titleLine, " #")
	}
	return titleLine
}
