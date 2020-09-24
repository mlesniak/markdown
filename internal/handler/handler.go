package handler

import (
	"github.com/labstack/echo/v4"
	"github.com/mlesniak/markdown/internal/cache"
	"net/http"
	"os"
)

const (
	// Directory containing static files for website.
	staticRoot = "static/"
)

// ContentHandler is the default handler for all non-static content. It uses the parameter name
// to download the correct markdown file from dropbox, perform various transformations
// and convert it to html.
func ContentHandler(c echo.Context) error {
	log := c.Logger()
	filename := c.Param("name")

	// Check if filename exists in static root directory. This is secure without checking
	// for parent paths (..) etc since we run in a docker container.
	ok := serveStaticFile(c, filename)
	if ok {
		return nil
	}

	// We only serve HTML files in dynamic content.
	c.Response().Header().Add("Content-Type", "text/html; charset=UTF-8")

	// Check if the file is in cache and can be used.
	html, inCache := useCache(log, filename)
	if inCache {
		return c.String(http.StatusOK, html)
	}

	// This can only happen if we are starting, since otherwise the cache is filled.
	log.Warn("File not yet in cache: %s", filename)
	return c.String(http.StatusNotFound, "File not found:"+filename)
}

// serveStaticFile is a special handler to service static files in the root directory
// which are actually stored in the static folder.
func serveStaticFile(c echo.Context, filename string) bool {
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
func useCache(log echo.Logger, filename string) (string, bool) {
	entry, ok := cache.Get().GetEntry(filename)
	if ok {
		log.Infof("Using cache. filename=%s", filename)
		return string(entry), true
	}

	return "", false
}
