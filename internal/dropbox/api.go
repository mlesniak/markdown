package dropbox

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"io/ioutil"
	"net/http"
	"time"
)

// apiCallHeader generalizes different api calls to dropbox.
//
// Will later be non-public again after Refactoring.
func (s *Service) apiCallHeader(log echo.Logger, url string, argument interface{}) ([]byte, error) {
	// Create general request.
	client := http.Client{}
	client.Timeout = time.Second * 10
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
	request.Header.Add("Authorization", "Bearer "+s.Token)
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
	client.Timeout = time.Second * 10
	request, err := http.NewRequest("POST", url, reader)
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %s", err)
	}

	// Set token and payload for submitting.
	request.Header.Add("Authorization", "Bearer "+s.Token)
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
