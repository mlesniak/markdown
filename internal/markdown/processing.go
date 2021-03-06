package markdown

import (
	"fmt"
	"github.com/mlesniak/markdown/internal/utils"
	"regexp"
	"strings"
)

const (
	// Default title if the title can not be extracted from the markdown file.
	defaultTitle = "mlesniak.com"
)

// processRawMarkdown performs various conversion steps which are not supported by
// the markdown processor. In addition, it uses the first line of the file to compute
// a potential title.
func processRawMarkdown(rawMarkdown []byte) string {
	markdown := string(rawMarkdown)
	markdown = convertTags(markdown)
	markdown = convertWikiLinks(markdown)
	markdown = convertImages(markdown)
	return markdown
}

func convertImages(markdown string) string {
	regex := regexp.MustCompile(`!\[\]\((.*?) (.*?)\)`)
	submatches := regex.FindAllStringSubmatch(markdown, -1)
	for _, matches := range submatches {
		text := matches[0]
		image := matches[1]
		width := matches[2]
		html := fmt.Sprintf(`<img src="%s" width="%s"/>`, image, width)
		markdown = strings.ReplaceAll(markdown, text, html)
	}

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
		if !strings.Contains(fileLinkName, ".") {
			fileLinkName = fileLinkName + ".md"
		}
		fileLinkName = strings.ReplaceAll(fileLinkName, " ", "-")
		markdownLink := fmt.Sprintf(`[%s](/%s)`, displayedName, fileLinkName)
		markdown = strings.ReplaceAll(markdown, wikiLink, markdownLink)
	}

	return markdown
}

func convertTags(markdown string) string {
	linkTags := utils.GetTags([]byte(markdown))

	for _, tag := range linkTags {
		link := fmt.Sprintf(`<a href="/tag-%s.md" class="tag">%s </a>`, tag[1:], tag)
		markdown = strings.ReplaceAll(markdown, tag, link)
	}

	return markdown
}
