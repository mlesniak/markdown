package handler

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	// Default title if the title can not be extracted from the markdown file.
	defaultTitle = "mlesniak.com"

	// Tag name to define markdown files which are allowed to be published.
	publishTag = "#public"
)

// processRawMarkdown performs various conversion steps which are not supported by
// the markdown processor. In addition, it uses the first line of the file to compute
// a potential title.
func processRawMarkdown(rawMarkdown []byte) string {
	markdown := string(rawMarkdown)
	markdown = removeTags(markdown)
	markdown = convertWikiLinks(markdown)
	return markdown
}

// convertWikiLinks converts wikiLinks to normal markdown links.
func convertWikiLinks(markdown string) string {
	regex := regexp.MustCompile(`\[\[(.*?)\]\]`)
	submatches := regex.FindAllStringSubmatch(markdown, -1)
	for _, matches := range submatches {
		if len(matches) < 2 {
			continue
		}
		fileLinkName := matches[1]
		wikiLink := matches[0]
		// Handle case in which a wikiLink links to a file without a timestamp.
		filenameParts := strings.SplitN(fileLinkName, " ", 2)
		var displayedName string
		if len(filenameParts) < 2 {
			displayedName = filenameParts[0]
		} else {
			displayedName = filenameParts[1]
		}
		markdownLink := fmt.Sprintf(`[%s](%s)`, displayedName, fileLinkName)
		markdown = strings.ReplaceAll(markdown, wikiLink, markdownLink)
	}

	return markdown
}

// removeTags removes all tags from the file.
func removeTags(markdown string) string {
	regex := regexp.MustCompile(`[\s]?#\w+`)
	matches := regex.FindAllString(markdown, -1)
	for _, match := range matches {
		if match != "" {
			markdown = strings.ReplaceAll(markdown, match, "")
		}
	}
	return markdown
}

// computeTitle uses the first line in markdown as title if available and feasible.
// Otherwise, default title is used.
func computeTitle(markdown string) string {
	titleLine := defaultTitle
	lines := strings.SplitN(markdown, "\n", 2)
	if len(lines) > 0 {
		titleLine = lines[0]
		// titleLine = strings.ReplaceAll(titleLine, "#", "")
		titleLine = strings.Trim(titleLine, " #")
	}
	return titleLine
}
