package backlinks

type filename = map[string]struct{}

type Backlinks struct {
	// Link from filename to list of files which are referencing it.
	links map[string]filename
}

func New() *Backlinks {
	return &Backlinks{
		links: make(map[string]filename),
	}
}

// For debugging
func (t *Backlinks) DebugGet() map[string]filename {
	return t.links
}

func (t *Backlinks) GetLinks(filename string) []string {
	mapFilenames, found := t.links[filename]
	if !found {
		return []string{}
	}

	filenames := []string{}
	for k, _ := range mapFilenames {
		filenames = append(filenames, k)
	}
	return filenames
}

func (t *Backlinks) Clear() {
	t.links = make(map[string]filename)
}

func (t *Backlinks) AddTargets(filename string, targets []string) {
	for _, name := range targets {
		// Race condition?
		if t.links[name] == nil {
			t.links[name] = make(map[string]struct{})
		}
		t.links[name][filename] = struct{}{}
	}
}
