package dropbox

import (
	"github.com/labstack/echo/v4"
	"github.com/mlesniak/markdown/internal/utils"
	"sync"
)

// Updater describes a function which is called when a file has been changed
// in the dropbox.
type Updater func(log echo.Logger, filename string, data []byte)

func (s *Service) PreloadCache(filenames ...string) {
	visited := make(map[string]struct{})

	s.Log.SetLevel(1)
	wg := sync.WaitGroup{}

	for _, filename := range filenames {
		visited[filename] = struct{}{}
		wg.Add(1)
		s.queue <- queueEntry{
			filename:  filename,
			finalizer: s.finalizer(visited, filename, &wg),
		}
	}

	wg.Wait()
	// TODO Update backlinks
}

var cnt int64

func (s *Service) finalizer(visited map[string]struct{}, filename string, wg *sync.WaitGroup) func(data []byte) {
	return func(data []byte) {
		defer wg.Done()

		links := utils.GetLinks(data)
		s.Log.Infof("For %s, found links=%v", filename, links)
		for _, link := range links {
			if _, found := visited[link]; found {
				continue
			}

			visited[link] = struct{}{}
			wg.Add(1)
			s.queue <- queueEntry{
				filename:  link,
				finalizer: s.finalizer(visited, link, wg),
			}
		}
	}
}

// func (s *Service) XPreloadCache(log echo.Logger, updater Updater, finalizer func()) {
// 	// Tree-search starting at the root file.
// 	// queue := make([]string, len(s.InitialRoots))
// 	// copy(queue, s.InitialRoots)
// 	queue := []string{}
// 	visited := make(map[string]struct{})
//
// 	wg := sync.WaitGroup{}
// 	wg.Add(len(queue))
//
// 	for len(queue) > 0 {
// 		filename := queue[0] + ".md"
// 		queue = queue[1:]
// 		if _, found := visited[filename]; found {
// 			continue
// 		}
//
// 		// Read file.
// 		bs, err := s.Read(log, filename)
// 		if err != nil {
// 			log.Warnf("Error reading file (continuing). filename=%s, error=%s", filename, err.Error())
// 			continue
// 		}
//
// 		// Update cache entry for this file asynchronously.
// 		go func(filenmae string, bs []byte) {
// 			updater(log, filename, bs)
// 			wg.Done()
// 		}(filename, bs)
//
// 		// Parse new filenames by searching for wikilinks.
// 		markdown := string(bs)
// 		regex := regexp.MustCompile(`\[\[(.*?)\]\]`)
// 		submatches := regex.FindAllStringSubmatch(markdown, -1)
// 		for _, matches := range submatches {
// 			wg.Add(1)
// 			queue = append(queue, matches[1])
// 		}
// 	}
//
// 	// We iterated through all files.
// 	wg.Wait()
// 	finalizer()
// }
