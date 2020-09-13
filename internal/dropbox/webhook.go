package dropbox

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/labstack/echo/v4"
	"io/ioutil"
	"net/http"
	"regexp"
)

// Updater describes a function which is called when a file has been changed
// in the dropbox.
type Updater func(log echo.Logger, filename string, data []byte)

// HandleChallenge returns the dropbox challenge which is used to check
// the webhook dropbox api.
func (s *Service) HandleChallenge(c echo.Context) error {
	challenge := c.Request().FormValue("challenge")
	// Initial dropbox challenge to register webhook.
	header := c.Response().Header()
	header.Add("Content-Type", "text/plain")
	header.Add("X-Content-Type-Options", "nosniff")
	return c.String(http.StatusOK, challenge)
}

// Here is a simple DOS attach possible preventing good cache behaviour? Think about this.
func (s *Service) WebhookHandler(updater Updater) echo.HandlerFunc {
	return func(c echo.Context) error {
		log := c.Logger()

		// Check signature for file updates to prevent DoS attacks.
		if !s.validSignature(c) {
			return c.String(http.StatusBadRequest, "Error with HMAC signature")
		}

		// We do not need to check the body since it's an internal application and
		// you do not need to verify which user account has changed data, since it
		// was mine by definition.
		type entries struct {
			Entries []entry `json:"entries"`
			Cursor  string  `json:"cursor"`
		}

		if s.cursor == "" {
			go func() {
				argument := struct {
					Path string `json:"path"`
				}{
					Path: "/notes",
				}
				bs, err := s.apiCall(c.Logger(), "https://api.dropboxapi.com/2/files/list_folder", argument)
				if err != nil {
					log.Infof("Error in initial dropbox call: %s", err.Error())
					return
				}
				var es entries
				json.Unmarshal(bs, &es)
				s.cursor = es.Cursor
				// Pre-load cache.
				// s.PreloadCache(log)
			}()
		} else {
			go func() {
				argument := struct {
					Cursor string `json:"cursor"`
				}{
					Cursor: s.cursor,
				}
				bs, err := s.apiCall(c.Logger(), "https://api.dropboxapi.com/2/files/list_folder/continue", argument)
				if err != nil {
					log.Infof("Error in continuous dropbox call for listing: %s", err.Error())
					return
				}
				var es entries
				json.Unmarshal(bs, &es)
				s.performCacheUpdate(log, es.Entries, updater)
				s.cursor = es.Cursor
			}()
		}

		return c.NoContent(http.StatusOK)
	}
}

func (s *Service) validSignature(c echo.Context) bool {
	log := c.Logger()

	// In out case we do not need the body elsewhere and can read it fully.
	body := c.Request().Body
	defer body.Close()
	bs, err := ioutil.ReadAll(body)
	if err != nil {
		log.Infof("Error while checking HMAC signature: %s", err.Error())
		return false
	}

	// Compute expected signature.
	mac := hmac.New(sha256.New, []byte(s.appSecret))
	mac.Write(bs)
	expectedMAC := mac.Sum(nil)

	// Convert actual signature.
	signature := c.Request().Header.Get("X-Dropbox-Signature")
	submittedMAC, err := hex.DecodeString(signature)

	// Compare.
	return hmac.Equal(submittedMAC, expectedMAC)
}

func (s *Service) performCacheUpdate(log echo.Logger, entries []entry, updater Updater) {
	for _, e := range entries {
		log.Infof("Updating cache entry. filename=%s", e.Name)
		bs, _ := s.Read(log, e.Name)
		updater(log, e.Name, bs)
	}
}

func (s *Service) PreloadCache(log echo.Logger) {
	visibleFiles := []string{}
	queue := make([]string, len(s.preloadRoot))
	copy(queue, s.preloadRoot)
	visited := make(map[string]struct{})

	for len(queue) > 0 {
		filename := queue[0] + ".md"
		queue = queue[1:]
		if _, found := visited[filename]; found {
			continue
		}
		println("Viewing " + filename)
		visibleFiles = append(visibleFiles, filename)

		// Read file.
		bs, err := s.Read(log, filename)
		if err != nil {
			log.Warnf("Error reading root file. filename=%s, error=%s", filename, err.Error())
			return
		}

		// Parse new filenames.
		markdown := string(bs)
		regex := regexp.MustCompile(`\[\[(.*?)\]\]`)
		submatches := regex.FindAllStringSubmatch(markdown, -1)
		for _, matches := range submatches {
			queue = append(queue, matches[1])
		}
	}

	for _, v := range visibleFiles {
		println(v)
	}
}
