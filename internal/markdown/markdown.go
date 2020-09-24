package markdown

import (
	"errors"
	"github.com/labstack/echo/v4"
	"github.com/mlesniak/markdown/internal/utils"
	"github.com/russross/blackfriday/v2"
	"io/ioutil"
	"strings"
)

func ToHTML(log echo.Logger, filename string, data []byte) (string, error) {
	// Perform various pre-processing steps on the markdown.
	markdown := processRawMarkdown(data)
	titleLine := title(markdown)

	// Convert from (processed) markdown to html.
	renderer := blackfriday.NewHTMLRenderer(blackfriday.HTMLRendererParameters{})
	html := string(blackfriday.Run([]byte(markdown), blackfriday.WithRenderer(renderer)))

	// Inject rendered html into template and fill variables.
	// We are intentionally not using html.template here since we
	// do don't want escaping, etc.
	bsTemplate, err := ioutil.ReadFile("template.html")
	if err != nil {
		log.Warn("Template not found. This should never happen.")
		return "", errors.New("template not found")
	}
	html = strings.ReplaceAll(string(bsTemplate), "{{content}}", html)
	html = strings.ReplaceAll(html, "{{title}}", titleLine)
	html = strings.ReplaceAll(html, "{{build}}", utils.BuildInformation())
	html = strings.ReplaceAll(html, "{{backlinks}}", generateBacklinkHTML(filename))

	return html, nil
}

// title uses the first line in markdown as title if available and feasible.
// Otherwise, default title is used.
func title(markdown string) string {
	titleLine := defaultTitle
	lines := strings.SplitN(markdown, "\n", 2)
	if len(lines) > 0 {
		titleLine = lines[0]
		titleLine = strings.Trim(titleLine, " #")
	}
	return titleLine
}
