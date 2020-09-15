// TODO Split functions into multiple files.
package markdown

import (
	"fmt"
	"regexp"
	"sort"
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
		markdownLink := fmt.Sprintf(`[%s](/%s)`, displayedName, fileLinkName)
		markdown = strings.ReplaceAll(markdown, wikiLink, markdownLink)
	}

	return markdown
}

func convertTags(markdown string) string {
	tags := GetTags([]byte(markdown))

	for _, tag := range tags {
		// link := fmt.Sprintf("[%s](/tag/%s)", tag, tag[1:])
		link := fmt.Sprintf(`<a href="/tag/%s" class="tag">%s </a>`, tag[1:], tag)
		markdown = strings.ReplaceAll(markdown, tag, link)
	}

	return markdown
}

func GetTags(data []byte) []string {
	markdown := string(data)
	regex := regexp.MustCompile(` *(#\w+)`)
	matches := regex.FindAllString(markdown, -1)

	tags := []string{}

	for _, tag := range matches {
		if tag != "" {
			// Triming is easier than using the matcher's group.
			tag = strings.Trim(tag, " \n\r\t")
			tags = append(tags, tag)
		}
	}

	sort.Strings(tags)
	return tags
}
