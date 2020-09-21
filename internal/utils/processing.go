// utils prevents cyclic imports between markdown and tags packages.
package utils

import (
	"regexp"
	"strings"
)

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

	return tags
}

func GetLinks(data []byte) []string {
	markdown := string(data)
	regex := regexp.MustCompile(`\[\[(.*?)\]\]`)

	links := []string{}

	submatches := regex.FindAllStringSubmatch(markdown, -1)
	for _, matches := range submatches {
		link := matches[1]
		links = append(links, link)
	}

	return links
}
