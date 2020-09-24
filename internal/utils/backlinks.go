// utils prevents cyclic imports between markdown and tags packages.
package utils

import (
	"regexp"
	"strings"
)

func GetLinks(data []byte) []string {
	markdown := string(data)
	regex := regexp.MustCompile(`<a href="\/([0-9].+?)"`)

	links := []string{}

	submatches := regex.FindAllStringSubmatch(markdown, -1)
	for _, matches := range submatches {
		link := matches[1]
		if !strings.HasSuffix(link, ".md") {
			link = link + ".md"
		}
		links = append(links, link)
	}

	return links
}