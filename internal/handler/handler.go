package handler

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/mlesniak/markdown/internal/backlinks"
	"github.com/mlesniak/markdown/internal/cache"
	"net/http"
	"os"
	"strings"
)

const (
	// Directory containing static files for website.
	staticRoot = "static/"
)

type Handler struct {
	Cache     *cache.Cache
	Backlinks *backlinks.Backlinks
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

	// We only serve HTML files in dynamic content.
	c.Response().Header().Add("Content-Type", "text/html; charset=UTF-8")

	// Check if the file is in cache and can be used.
	filename = h.fixFilename(filename)
	html, inCache := h.useCache(log, filename)
	if inCache {
		// Add backlinks. TODO always empty list
		links := h.Backlinks.GetLinks(filename)
		// .md extensions is here but not in backlinks structure since not in links
		// design decision: must be present, fix in GetLinks
		slinks := fmt.Sprintf("%v", links)
		html = strings.ReplaceAll(html, "{{backlinks}}", slinks)
		return c.String(http.StatusOK, html)
	}

	// This can only happen if we are starting, since otherwise the cache is filled.
	log.Warn("File not in cache: %s", filename)
	return c.String(http.StatusNotFound, "File not found:"+filename)
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
