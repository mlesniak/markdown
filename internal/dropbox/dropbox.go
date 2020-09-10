package dropbox

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// Service contains the necessary data to access a dropbox.
type Service struct {
	appSecret     string
	token         string
	rootDirectory string
	// Since we have only one account, the cursor is part of the service.
	cursor string
}

type entry struct {
	Tag  string `json:".tag"`
	Name string `json:"name"`
}

type Updater func(log echo.Logger, filename string, data []byte)

// New returns a new dropbox service.
//
// The token is either generated by the normal OAuth2 workflow from
// dropbox or a token manually generated using the app console for your
// specific application.
//
// The rootDirectory is the root for all accessed files.
func New(appSecret, token string, rootDirectory string) *Service {
	if !strings.HasSuffix(rootDirectory, "/") {
		rootDirectory = rootDirectory + "/"
	}

	return &Service{
		appSecret:     appSecret,
		token:         token,
		rootDirectory: rootDirectory,
	}
}

// Read downloads the requested file from dropbox.
//
// I'm still not happy that the echo logger interface is polluting our
// dropbox service instead of a more general (log) or custom (zerolog)
// interface. Of course, I could write a wrapper back from lecho to
// zerolog, but this is a lot of work for this small program, hence 🤷‍.
// Although I miss zerlog's context, e.g. for filenames.
func (s *Service) Read(log echo.Logger, filename string) ([]byte, error) {
	start := time.Now()

	argument := struct {
		Path string `json:"path"`
	}{
		Path: "/" + s.rootDirectory + filename,
	}
	bs, err := s.ApiCallHeader(log, "https://content.dropboxapi.com/2/files/download", argument)
	if err != nil {
		return nil, err
	}

	log.Infof("Read file from dropbox. filename=%s, duration=%v", filename, time.Since(start).Milliseconds())
	return bs, err
}

// ApiCallHeader generalizes different api calls to dropbox.
//
// Will later be non-public again after Refactoring.
func (s *Service) ApiCallHeader(log echo.Logger, url string, argument interface{}) ([]byte, error) {
	// Create general request.
	client := http.Client{}
	request, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %s", err)
	}

	// Create payload.
	rawJson, err := json.Marshal(argument)
	if err != nil {
		return nil, fmt.Errorf("unable to create payload: %s", err)
	}

	// Set token and payload for submitting.
	request.Header.Add("Authorization", "Bearer "+s.token)
	request.Header.Add("Dropbox-API-Arg", string(rawJson))

	// Execute request.
	log.Infof("Performing dropbox API call to %s with payload=%v", url, argument)
	resp, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("unable to perform request: %s", err)
	}
	defer resp.Body.Close()

	// Read response.
	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read data from response: %s", err)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("non 200 response from dropbox: `%s`", string(bs))
	}

	// Return data and log elapsed time.
	return bs, err
}

// Consistency is not dropbox's strength, although I understand the idea behind this :-/
func (s *Service) apiCall(log echo.Logger, url string, argument interface{}) ([]byte, error) {
	// Create payload.
	rawJson, err := json.Marshal(argument)
	if err != nil {
		return nil, fmt.Errorf("unable to create payload: %s", err)
	}
	reader := bytes.NewReader(rawJson)

	// Create general request.
	client := http.Client{}
	request, err := http.NewRequest("POST", url, reader)
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %s", err)
	}

	// Set token and payload for submitting.
	request.Header.Add("Authorization", "Bearer "+s.token)
	request.Header.Set("Content-Type", "application/json")

	// Execute request.
	log.Infof("Performing dropbox API call to %s with payload=%v", url, string(rawJson))
	resp, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("unable to perform request: %s", err)
	}
	defer resp.Body.Close()

	// Read response.
	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read data from response: %s", err)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("non 200 response from dropbox: `%s`", string(bs))
	}

	// Return data and log elapsed time.
	return bs, err
}

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
