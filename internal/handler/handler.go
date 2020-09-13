package handler

import (
	"github.com/labstack/echo/v4"
	"github.com/mlesniak/markdown/internal/cache"
	"github.com/mlesniak/markdown/internal/tags"
	"net/http"
	"os"
)

const (
	// Directory containing static files for website.
	staticRoot = "static/"
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
	log := c.Logger()
	filename := c.Param("name")

	// Check if filename exists in static root directory. This is secure without checking
	// for parent paths (..) etc since we run in a docker container.
	ok := h.serveStaticFile(c, filename)
	if ok {
		return nil
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

func (h *Handler) HandleTag(c echo.Context) error {
	tag := c.Param("tag")
	tag = "#" + tag
	fileList := h.Tags.List(tag)
	return c.JSON(http.StatusOK, fileList)
}
