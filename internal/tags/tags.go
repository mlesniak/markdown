package tags

import "strings"

type tags = map[string]struct{}

type Tags struct {
	tags map[string]tags
}

func New() *Tags {
	return &Tags{
		tags: make(map[string]tags),
	}
}

func (t *Tags) Update(filename string, tags tags) {
	// Ignore adding tags.
	if strings.HasPrefix(filename, "#") {
		return
	}

	t.tags[filename] = tags
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
