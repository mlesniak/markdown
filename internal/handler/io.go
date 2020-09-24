package handler

import (
	"strings"
)

// fixFilename transform the requested filename, i.e. redirects to
// index page or fixes simplified filenames without suffix.
func fixFilename(filename string) string {
	// In our markup, wiki links have no markdown suffix.
	// Append suffix if not yet present.
	if !strings.HasSuffix(filename, ".md") {
		filename = filename + ".md"
	}

	return filename
}
