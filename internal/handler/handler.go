package handler

import (
	"bytes"
	"github.com/labstack/echo/v4"
	"github.com/russross/blackfriday/v2"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/mlesniak/markdown/internal/cache"
)

const (
	// Directory containing static files for website.
	staticRoot = "static/"
)

// StorageReader allows to read data from an abstract storage.
type StorageReader interface {
	// Implemented by the dropbox interface.
	Read(log echo.Logger, filename string) ([]byte, error)
}

type Handler struct {
	RootFilename  string
	StorageReader StorageReader
	Cache         *cache.Cache
}

// Handle is the default handler for all non-static content. It uses the parameter name
// to download the correct markdown file from dropbox, perform various transformations
// and convert it to html.
func (h *Handler) Handle(c echo.Context) error {
	log := c.Logger()
	filename := c.Param("name")

	// Check if filename exists in static root directory. This is secure without checking
	// for parent paths (..) etc since we run in a docker container.
	if filename != "" {
		virtualPath := staticRoot + filename
		_, err := os.Stat(virtualPath)
		if err == nil {
			log.Infof("Serving static virtual file. filename=%s", filename)
			return c.File(virtualPath)
		}
	}

	// Append markdown suffix and handle / - path.
	filename = h.fixFilename(filename)

	// Check if the file is in cache and can be used.
	html, found := h.useCache(log, filename)
	if !found {
		// Try to read file from dropbox storage.
		tmp, stop := h.readFromStorage(c, filename)
		if stop {
			// If we should stop, we always return 404 for security reasons.
			return c.String(http.StatusNotFound, "File not found:"+filename)
		}
		html = tmp
	}

	// Return generated HTML file with correct content type.
	c.Response().Header().Add("Content-Type", "text/html; charset=UTF-8")
	return c.String(http.StatusOK, html)
}

// useCache tries to use the cache entry to serve a precomputed and stored file.
func (h *Handler) useCache(log echo.Logger, filename string) (string, bool) {
	entry, ok := h.Cache.Get(filename)
	if ok {
		log.Infof("Using cache. filename=%s", filename)
		return string(entry), true
	}

	return "", false
}

// readFromStorage reads the given file from dropbox. If there is an error,
// true is returned and an error message is stored in the return value of
// the context, i.e. with c.String(...).
func (h *Handler) readFromStorage(c echo.Context, filename string) (string, bool) {
	log := c.Logger()

	// Read file from dropbox.
	bs, err := h.StorageReader.Read(c.Logger(), filename)
	if err != nil {
		log.Infof("Error reading file: %v for %s", err, filename)
		return "", true
	}

	return h.RenderFile(log, filename, bs)
}

// TODO Codepath is getting a bit obscure, refactor this.
func (h *Handler) RenderFile(log echo.Logger, filename string, data []byte) (string, bool) {
	// Are we allowed to display this file?
	if !isPublic(data) {
		// We use the same error message to prevent
		// guessing non-accessible filenames.
		log.Infof("File not public accessible: %s", filename)
		return "", true
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
		return "", true
	}
	html = strings.ReplaceAll(string(bsTemplate), "${content}", html)
	html = strings.ReplaceAll(html, "${title}", titleLine)

	// Add to cache.
	h.Cache.Add(cache.Entry{
		Name: filename,
		Data: []byte(html),
	})

	return html, false
}

// isPublic checks if a file is allowed to be displayed: Since we are only
// downloading markdown files, we enforce that all files must contain the tag
// `publishTag` to be able to download it.
func isPublic(bs []byte) bool {
	return bytes.Contains(bs, []byte(publishTag))
}

// fixFilename transform the requested filename, i.e. redirects to
// index page or fixes simplified filenames without suffix.
func (h *Handler) fixFilename(filename string) string {
	// If / is requested we redirect to our index page.
	if filename == "" {
		filename = h.RootFilename
	}

	// In our markup, wiki links have no markdown suffix.
	// Append suffix if not yet present.
	if !strings.HasSuffix(filename, ".md") {
		filename = filename + ".md"
	}

	return filename
}
