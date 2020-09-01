package main

import (
	"bytes"
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
	// Default title if the title can not be extracted from the markdown file.
	defaultTitle = "mlesniak.com"

	// Name of the root file if no filename is specified.
	rootFilename = "202009010520 index"

	// Tag name to define markdown files which are allowed to be published.
	publishTag = "#public"
)

// main is the entry point :-).
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
	bs, err := readFromDropbox(filename)
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

// processRawMarkdown performs various conversion steps which are not supported by
// the markdown processor. In addition, it uses the first line of the file to compute
// a potential title.
func processRawMarkdown(rawMarkdown []byte) string {
	markdown := string(rawMarkdown)

	// Remove all tags.
	regex := regexp.MustCompile(`[\s]?#\w+`)
	matches := regex.FindAllString(markdown, -1)
	for _, match := range matches {
		if match != "" {
			markdown = strings.ReplaceAll(markdown, match, "")
		}
	}

	// Handle wiki-Links.
	regex = regexp.MustCompile(`\[\[(.*?)\]\]`)
	submatches := regex.FindAllStringSubmatch(markdown, -1)
	for _, matches := range submatches {
		if len(matches) < 2 {
			continue
		}
		fileLinkName := matches[1]
		wikiLink := matches[0]
		displayedName := strings.SplitN(fileLinkName, " ", 2)[1]
		markdownLink := fmt.Sprintf(`[%s](%s)`, displayedName, fileLinkName)
		markdown = strings.ReplaceAll(markdown, wikiLink, markdownLink)
	}

	return markdown
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

// readFromDropbox downloads the requested file from dropbox.
// TODO Dropbox client service
func readFromDropbox(filename string) ([]byte, error) {
	// Will be configured in the service later on.
	token := os.Getenv("TOKEN")

	// Create request.
	client := http.Client{}
	request, err := http.NewRequest("POST", "https://content.dropboxapi.com/2/files/download", nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %s", err)
	}
	argument := fmt.Sprintf(`{"path": "/notes/%s"}`, filename)
	request.Header.Add("Authorization", "Bearer "+token)
	request.Header.Add("Dropbox-API-Arg", argument)

	// Execute request.
	resp, err := client.Do(request)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// Read file content.
	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read file from response: %s", err)
	}

	return bs, err
}
