package markdown

import (
	"fmt"
	"github.com/mlesniak/markdown/internal/backlinks"
	"github.com/mlesniak/markdown/internal/utils"
	"regexp"
	"sort"
	"strings"
)

func generateBacklinkHTML(filename string) string {
	buf := strings.Builder{}
	links := backlinks.Get().GetLinks(filename)
	if len(links) > 0 {
		// Sort links by timestamp (for now).
		sort.Strings(links)

		// Generate HTML.
		buf.WriteString(`<hr/>This page is referenced by<ul>`)
		for _, name := range links {
			displayName := visibleLink(name)
			link := fmt.Sprintf(`<li><a href="/%s">%s</a></li>`, name, displayName)
			buf.WriteString("\n")
			buf.WriteString(link)
		}
		buf.WriteString(`</ul>`)
	}
	backLinkHTML := buf.String()
	return backLinkHTML
}

// visibleLink converts a filename to a displayable variant, i.e. for
// the name 202009010520 Index foo bar.md it returns `Index Foo Bar`.
func visibleLink(filename string) string {
	rx := regexp.MustCompile(`\d* ?(.*?)\.md`)
	matches := rx.FindStringSubmatch(filename)
	if len(matches) < 1 {
		return utils.AutoCaptialize(filename)
	}

	return utils.AutoCaptialize(matches[1])
}
