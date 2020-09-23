package dropbox

import (
	"strings"
)

// Service contains the necessary data to access a dropbox.
type Service struct {
	AppSecret     string
	Token         string
	RootDirectory string
	InitialRoots  []string
	// Since we have only one account, the cursor is part of the service.
	cursor string
}

type entry struct {
	Tag  string `json:".tag"`
	Name string `json:"name"`
}

// New returns a new dropbox service.
//
// The token is either generated by the normal OAuth2 workflow from
// dropbox or a token manually generated using the app console for your
// specific application.
//
// The rootDirectory is the root for all accessed files.
func New(s Service) *Service {
	if !strings.HasSuffix(s.RootDirectory, "/") {
		s.RootDirectory = s.RootDirectory + "/"
	}

	return &s
}
