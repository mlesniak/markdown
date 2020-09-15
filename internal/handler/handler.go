package handler

import (
	"bytes"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/mlesniak/markdown/internal/cache"
	"github.com/mlesniak/markdown/internal/markdown"
	"github.com/mlesniak/markdown/internal/tags"
	"net/http"
	"os"
	"strings"
)

const (
	// Directory containing static files for website.
	staticRoot = "static/"

	// Tag name to define markdown files which are allowed to be published.
	publishTag = "#public"
)

type Handler struct {
	RootFilename  string
	StorageReader StorageReader
	Cache         *cache.Cache
	Tags          *tags.Tags
}

// Handle is the default handler for all non-static content. It uses the parameter name
// to download the correct markdown file from dropbox, perform various transformations
// and convert it to html.
func (h *Handler) Handle(c echo.Context) error {
	// ******************************************************************************
	// TODO Is this complex logic actually necessary? We will always use cache due to
	//      path traversal from root? Or I simply know the filename.
	//		Won't need security, then, anyway.
	// ******************************************************************************

	log := c.Logger()
	filename := c.Param("name")

	// Check if filename exists in static root directory. This is secure without checking
	// for parent paths (..) etc since we run in a docker container.
	ok := h.serveStaticFile(c, filename)
	if ok {
		return nil
	}

	// We only serve HTML files in dynamic content.
	c.Response().Header().Add("Content-Type", "text/html; charset=UTF-8")

	// Append markdown suffix and handle / - path.
	filename = h.fixFilename(filename)

	// Check if the file is in cache and can be used.
	html, inCache := h.useCache(log, filename)
	if inCache {
		return c.String(http.StatusOK, html)
	}

	// Try to read file from dropbox storage.
	bs, stop := h.readFromStorage(c, filename)
	if stop {
		// If we should stop, we always return 404 for security reasons.
		return c.String(http.StatusNotFound, "File not found:"+filename)
	}

	// Are we allowed to display this file?
	if !isPublic(bs) {
		// We use the same error message to prevent guessing non-accessible filenames.
		log.Infof("File not public accessible: %s", filename)
		return c.String(http.StatusNotFound, "File not found:"+filename)
	}

	// TODO Work consistently on []byte instead of switching ...
	tagList := markdown.GetTags(bs)
	h.Tags.Update(filename, tagList)

	// Render file and process markdown.
	html, err := markdown.ToHTML(log, filename, bs)
	if err != nil {
		// If we should stop, we always return 404 for security reasons.
		return c.String(http.StatusNotFound, "File not found:"+filename)
	}

	// Add to cache.
	h.Cache.Add(cache.Entry{
		Name: filename,
		Data: []byte(html),
	})

	return c.String(http.StatusOK, html)
}

// serveStaticFile is a special handler to service static files in the root directory
// which are actually stored in the static folder.
func (h *Handler) serveStaticFile(c echo.Context, filename string) bool {
	log := c.Logger()

	if filename != "" {
		virtualPath := staticRoot + filename
		_, err := os.Stat(virtualPath)
		if err == nil {
			log.Infof("Serving static virtual file. filename=%s", filename)
			c.File(virtualPath)
			return true
		}
	}
	return false
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

// TODO Move to tags package
//
// TODO Internal flag == true is bad design.
// TODO Is this another special case for read from storage?
func (h *Handler) HandleTag(c echo.Context) error {
	tag := c.Param("tag")
	tag = "#" + tag
	fileList := h.Tags.List(tag)

	tags := strings.Builder{}
	for _, file := range fileList {
		tags.WriteString("- [[")
		tags.WriteString(file)
		tags.WriteString("]]\n")
	}

	// Create dynamic markdown.
	md := []byte(fmt.Sprintf("# %s\n\n%s", tag, tags.String()))

	html, _ := markdown.ToHTML(c.Logger(), tag, md)
	// html := markdown
	c.Response().Header().Add("Content-Type", "text/html; charset=UTF-8")
	return c.String(http.StatusOK, html)
}

// isPublic checks if a file is allowed to be displayed: Since we are only
// downloading markdown files, we enforce that all files must contain the tag
// `publishTag` to be able to download it.
func isPublic(bs []byte) bool {
	return bytes.Contains(bs, []byte(publishTag))
}
