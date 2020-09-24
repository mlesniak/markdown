package dropbox

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/labstack/echo/v4"
	"io/ioutil"
	"net/http"
)

type Updater func()

type entry struct {
	Tag  string `json:".tag"`
	Name string `json:"name"`
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
				updater()
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
	mac := hmac.New(sha256.New, []byte(s.AppSecret))
	mac.Write(bs)
	expectedMAC := mac.Sum(nil)

	// Convert actual signature.
	signature := c.Request().Header.Get("X-Dropbox-Signature")
	submittedMAC, err := hex.DecodeString(signature)

	// Compare.
	return hmac.Equal(submittedMAC, expectedMAC)
}
