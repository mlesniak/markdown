package dropbox

import (
	"github.com/labstack/echo/v4"
	"github.com/mlesniak/markdown/internal/utils"
	"sync"
)

// Updater describes a function which is called when a file has been changed
// in the dropbox.
type Updater func(log echo.Logger, filename string, data []byte)

// func (s *Service) PreloadCache(filenames ...string) {
// 	now := time.Now()
//
// 	s.loadCache(filenames)
//
// 	// All files have been loaded once. Scan them to create backlink map and
// 	// cache them again.
// 	c := cache.Get()
// 	for _, filename := range c.List() {
// 		s.Log.Infof("Scanning for backlinks, filename=%s", filename)
// 		bs, _ := c.GetEntry(filename)
// 		links := utils.getLinks(bs)
// 		backlinks.Get().AddTargets(filename, links)
// 	}
//
// 	s.Log.Info("Recaching with computed backlinks")
// 	s.loadCache(filenames)
//
// 	s.Log.Infof("Rebuilding site took %d ms", time.Now().Sub(now).Milliseconds())
// }

func (s *Service) loadCache(filenames []string) {
	visited := make(map[string]struct{})
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
}

func (s *Service) finalizer(visited map[string]struct{}, filename string, wg *sync.WaitGroup) func(data []byte) {
	return func(data []byte) {
		defer wg.Done()

		links := utils.GetLinks(data)
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
