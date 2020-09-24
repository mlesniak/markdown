package dropbox

import (
	"time"
)

// Read downloads the requested file from dropbox.
//
// I'm still not happy that the echo logger interface is polluting our
// dropbox service instead of a more general (log) or custom (zerolog)
// interface. Of course, I could write a wrapper back from lecho to
// zerolog, but this is a lot of work for this small program, hence 🤷‍.
// Although I miss zerlog's context, e.g. for filenames.
func (s *Service) Read(filename string) ([]byte, error) {
	start := time.Now()

	argument := struct {
		Path string `json:"path"`
	}{
		Path: "/" + s.RootDirectory + filename,
	}
	bs, err := s.apiCallHeader("https://content.dropboxapi.com/2/files/download", argument)
	if err != nil {
		return nil, err
	}

	s.Log.Infof("Read file from dropbox. filename=%s, duration=%v", filename, time.Since(start).Milliseconds())
	return bs, err
}
