package handler

import (
	"github.com/labstack/echo/v4"
	"github.com/mlesniak/markdown/internal/cache"
	"github.com/mlesniak/markdown/internal/dropbox"
	"net/http"
	"os"
	"strings"
)

const (
	// Directory containing static files for website.
	staticRoot = "static/"
)

// ContentHandler is the default handler for all non-static content. It uses the parameter name
// to download the correct markdown file from dropbox, perform various transformations
// and convert it to html.
func ContentHandler(dropbox *dropbox.Service) echo.HandlerFunc {
	return func(c echo.Context) error {
		log := c.Logger()
		filename := c.Param("name")

		// Check if filename exists in static root directory. This is secure without checking
		// for parent paths (..) etc since we run in a docker container.
		ok := serveStaticFile(c, filename)
		if ok {
			return nil
		}

		// Compute suffix.
		parts := strings.Split(filename, ".")
		if len(parts) < 1 {
			log.Warnf("Filename without suffix. filename=%s", filename)
			return c.String(http.StatusNotFound, "File not found:"+filename)
		}
		suffix := parts[1]

		// Load data based on suffix.
		switch suffix {
		case "png":
			bs, inCache := useCache(log, filename)
			if !inCache {
				// Load file once into cache.
				tmp, err := dropbox.Read("media/" + filename)
				if err != nil {
					return c.String(http.StatusNotFound, "File not found:"+filename)
				}
				bs = tmp
				inCache = true
				cache.Get().AddEntry(cache.Entry{
					Name: filename,
					Data: bs,
				})
				return c.Blob(http.StatusOK, "image/png", bs)
			}
		case "md":
			// Markdown files are initially cached.
			bs, inCache := useCache(log, filename)
			if !inCache {
				return c.String(http.StatusNotFound, "File not found:"+filename)
			}
			return c.Blob(http.StatusOK, "text/html; charset=UTF-8", bs)
		default:
			return c.String(http.StatusNotFound, "File not found:"+filename)
		}

		return nil
	}
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
// TODO Remove this?
func useCache(log echo.Logger, filename string) ([]byte, bool) {
	entry, ok := cache.Get().GetEntry(filename)
	if ok {
		log.Infof("Using cache. filename=%s", filename)
		return entry, true
	}

	return []byte{}, false
}
