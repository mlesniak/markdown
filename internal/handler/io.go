package handler

import (
	"strings"
)

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
