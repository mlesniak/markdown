package backlinks

type parent = map[string]struct{}

type Backlinks struct {
	// Link from parent to list of files which are referencing it.
	links map[string]parent
}

var singleton *Backlinks

func init() {
	singleton = &Backlinks{
		links: make(map[string]parent),
	}
}

func Get() *Backlinks {
	return singleton
}

func (t *Backlinks) GetParents(filename string) []string {
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

func (t *Backlinks) AddChildren(filename string, targets []string) {
	for _, name := range targets {
		// Race condition?
		if t.links[name] == nil {
			t.links[name] = make(map[string]struct{})
		}
		t.links[name][filename] = struct{}{}
	}
}
