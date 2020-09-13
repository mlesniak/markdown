package handler

import (
	"github.com/labstack/echo/v4"
	"strings"
)

// StorageReader allows to read data from an abstract storage.
type StorageReader interface {
	// Implemented by the dropbox interface.
	Read(log echo.Logger, filename string) ([]byte, error)
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

	return h.RenderFile(log, false, filename, bs)
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
