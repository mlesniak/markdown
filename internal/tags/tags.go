package tags

import (
	"strings"
)

type tags = map[string]struct{}

type Tags struct {
	tags map[string]tags
}

func New() *Tags {
	return &Tags{
		tags: make(map[string]tags),
	}
}

func (t *Tags) Clear() {
	t.tags = make(map[string]tags)
}

func (t *Tags) Update(filename string, tags []string) {
	// Ignore adding tags.
	if strings.HasPrefix(filename, "#") {
		return
	}

	tm := make(map[string]struct{})
	for _, tag := range tags {
		tm[tag] = struct{}{}
	}

	t.tags[filename] = tm
}

func (t *Tags) List(tag string) []string {
	filenames := []string{}

	for filename, tags := range t.tags {
		if _, found := tags[tag]; found {
			filenames = append(filenames, filename)
		}
	}

	return filenames
}
