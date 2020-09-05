package main

import (
	"bytes"
	"github.com/labstack/echo/v4"
	"github.com/russross/blackfriday/v2"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

// Move this to own service including dropbox callback etc.
type cacheEntry struct {
	name      string
	createdAt time.Time
	data      []byte
}

var cache = make(map[string]cacheEntry)

// handle is the default handler for all non-static content. It uses the parameter name
// to download the correct markdown file from dropbox, perform various transformations
// and convert it to html.
func handle(c echo.Context) error {
	log := c.Logger()
	filename := c.Param("name")

	// Check if filename exists in static root directory. This is secure without checking
	// for parent paths (..) etc since we run in a docker container.
	if filename != "" {
		virtualPath := staticRoot + filename
		println(virtualPath)
		_, err := os.Stat(virtualPath)
		if err == nil {
			log.Infof("Serving static virtual file. filename=%s", filename)
			return c.File(virtualPath)
		}
	}

	// Append markdown suffix and handle / - path.
	filename = fixFilename(filename)

	// Check if the file is in cache and can be used.
	stop := useCache(c, filename)
	if stop {
		return nil
	}

	// Try to read file from dropbox storage.
	html, stop := readFromStorage(c, filename)
	if stop {
		// We stop delivery due to some error which was already stored in the
		// echo context using c.String.
		return nil
	}

	// Return generated HTML file with correct content type.
	c.Response().Header().Add("Content-Type", "text/html; charset=UTF-8")
	return c.String(http.StatusOK, html)
}

// useCache tries to use the cache entry to serve a precomputed and stored file.
func useCache(c echo.Context, filename string) bool {
	log := c.Logger()

	entry, ok := cache[filename]
	if ok {
		// Check TTL
		maxTTL, _ := time.ParseDuration("5s")
		validTill := entry.createdAt.Add(maxTTL)
		if validTill.After(time.Now()) {
			// Return generated HTML file with correct content type.
			log.Infof("Using cache, validTill=%v. filename=%s", validTill, filename)
			c.Response().Header().Add("Content-Type", "text/html; charset=UTF-8")
			c.String(http.StatusOK, string(entry.data))
			return true
		} else {
			// Remove from cache.
			log.Infof("Deleting from cache and retrieving again. filename=%s", filename)
			delete(cache, filename)
		}
	}

	return false
}

// readFromStorage reads the given file from dropbox. If there is an error,
// false is returned and an error message is stored in the return value of
// the context, i.e. with c.String(...).
func readFromStorage(c echo.Context, filename string) (string, bool) {
	log := c.Logger()

	// Read file from dropbox.
	bs, err := dropboxService.Read(c.Logger(), filename)
	if err != nil {
		log.Infof("Error reading file: %v for %s", err, filename)
		_ = c.String(http.StatusNotFound, "File not found:"+filename)
		return "", true
	}

	// Are we allowed to display this file?
	if !isPublic(bs) {
		// We use the same error message to prevent
		// guessing non-accessible filenames.
		log.Infof("File not public accessible: %s", filename)
		_ = c.String(http.StatusNotFound, "File not found:"+filename)
		return "", true
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
		log.Info("Template not found")
		_ = c.String(http.StatusInternalServerError, "Template not found. This should never happen.")
		return "", true
	}
	html = strings.ReplaceAll(string(bsTemplate), "${content}", html)
	html = strings.ReplaceAll(html, "${title}", titleLine)

	// Add to cache.
	cache[filename] = cacheEntry{
		name:      filename,
		createdAt: time.Now(),
		data:      []byte(html),
	}

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
