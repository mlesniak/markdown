package tags

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/mlesniak/markdown/internal/markdown"
	"net/http"
	"sort"
	"strings"
)

func (t *Tags) HandleTag(c echo.Context) error {
	tag := c.Param("tag")
	tag = "#" + tag
	filenameList := t.List(tag)

	// Take first h1 of file from cache?
	// Sort by this then?
	titlesFilenames := make(map[string]string)
	for _, filename := range filenameList {
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
		name := titlesFilenames[title]
		link := fmt.Sprintf(`- <a href="/%s">%s</a>`, name, title)
		tags.WriteString("\n")
		tags.WriteString(link)
	}

	// Create dynamic markdown.
	md := []byte(fmt.Sprintf("# Articles tagged %s\n\n%s", tag, tags.String()))

	html, _ := markdown.ToHTML(c.Logger(), tag, md)
	// html := markdown
	c.Response().Header().Add("Content-Type", "text/html; charset=UTF-8")
	return c.String(http.StatusOK, html)
}
