package tags

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/mlesniak/markdown/internal/markdown"
	"github.com/mlesniak/markdown/internal/utils"
	"sort"
	"strings"
)

func GenerateTagPage(log echo.Logger, tag string, filenames []string) []byte {
	titlesFilenames := make(map[string]string)
	for _, filename := range filenames {
		parts := strings.SplitN(filename, " ", 2)
		var titleName string
		if len(parts) < 2 {
			titleName = parts[0]
		} else {
			titleName = parts[1]
		}
		// Remove .md suffix
		titleName = titleName[:len(titleName)-3]
		titlesFilenames[titleName] = filename
	}

	// Get list and sort.
	titles := []string{}
	for k, _ := range titlesFilenames {
		titles = append(titles, k)
	}
	sort.Slice(titles, func(i, j int) bool {
		return strings.ToLower(titles[i]) < strings.ToLower(titles[j])
	})

	tags := strings.Builder{}
	for _, title := range titles {
		displayTitle := utils.AutoCaptialize(title)

		name := titlesFilenames[title]
		link := fmt.Sprintf(`- <a href="/%s">%s</a>`, name, displayTitle)
		tags.WriteString("\n")
		tags.WriteString(link)
	}
	content := tags.String()

	// Create dynamic markdown.
	md := []byte(fmt.Sprintf("# Articles tagged %s\n\n%s", tag[1:], content))

	html, _ := markdown.ToHTML(log, "", md)
	html = strings.ReplaceAll(html, "{{title}}", tag)
	html = strings.ReplaceAll(html, "{{backlinks}}", "")

	return []byte(html)
}
