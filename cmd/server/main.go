package main

import (
	"encoding/json"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/mlesniak/markdown/pkg/dropbox"
	"github.com/ziflex/lecho/v2"
	"net/http"
	"os"
)

const (
	// Default title if the title can not be extracted from the markdown file.
	defaultTitle = "mlesniak.com"

	// Name of the root file if no filename is specified.
	rootFilename = "202009010520 index"

	// Tag name to define markdown files which are allowed to be published.
	publishTag = "#public"

	// External mountable directory to publish files.
	downloadRoot = "download/"

	// Directory containing static files for website.
	staticRoot = "static/"
)

var dropboxService *dropbox.Service

func init() {
	dropboxToken := os.Getenv("TOKEN")
	if dropboxToken == "" {
		panic("No dropbox token set, aborting.")
	}
	dropboxService = dropbox.New(dropboxToken, "notes/")
}

func main() {
	e := echo.New()

	// Configure logging.
	logger := lecho.New(
		os.Stdout,
		lecho.WithLevel(log.INFO),
		lecho.WithTimestamp(),
		lecho.WithCallerWithSkipFrameCount(3),
	)
	e.Logger = logger

	// Configure middlewares.
	e.Use(middleware.RequestID())
	e.Use(lecho.Middleware(lecho.Config{
		Logger: logger,
	}))

	// Serve static and downloadable files.
	e.Static("/static", staticRoot)
	e.Static("/download", downloadRoot)

	// Serve dynamic files.
	e.GET("/", handle)
	e.GET("/:name", handle)

	// Handle cache invalidation through dropbox webhooks.
	e.GET("/dropbox/webhook", dropboxChallenge)
	e.POST("/dropbox/webhook", dropboxUpdate)

	// Start server.
	e.HideBanner = true
	e.HidePort = true
	e.Logger.Fatal(e.Start(":8080"))
}

func dropboxChallenge(c echo.Context) error {
	challenge := c.Request().FormValue("challenge")
	// Initial dropbox challenge to register webhook.
	header := c.Response().Header()
	header.Add("Content-Type", "text/plain")
	header.Add("X-Content-Type-Options", "nosniff")
	return c.String(http.StatusOK, challenge)
}

var cursor string

type entry struct {
	Tag  string `json:".tag"`
	Name string `json:"name"`
}

// Here is a simple DOS attach possible preventing good cache behaviour? Think about this.
func dropboxUpdate(c echo.Context) error {
	// We do not need to check the body since it's an internal application and
	// you do not need to verify which user account has changed data, since it
	// was mine by definition.
	type entries struct {
		Entries []entry `json:"entries"`
		Cursor  string  `json:"cursor"`
	}

	if cursor == "" {
		go func() {
			// Note that try to populate the cache also with non-public entries,
			// which are then not downloaded, so everything is fine, albeit this
			// is quite a bit slower.
			//
			// TODO If we have no cursor, maybe use search for #public tag instead of listing all files?
			argument := struct {
				Path string `json:"path"`
			}{
				Path: "/notes",
			}
			bs, err := dropboxService.ApiCallBody(c.Logger(), "https://api.dropboxapi.com/2/files/list_folder", argument)
			if err != nil {
				println("Ouch " + err.Error())
				return
			}
			var es entries
			json.Unmarshal(bs, &es)
			performCacheUpdate(c, es.Entries)
			cursor = es.Cursor
		}()
	} else {
		go func() {
			argument := struct {
				Cursor string `json:"cursor"`
			}{
				Cursor: cursor,
			}
			bs, err := dropboxService.ApiCallBody(c.Logger(), "https://api.dropboxapi.com/2/files/list_folder/continue", argument)
			if err != nil {
				println("Ouch " + err.Error())
				return
			}
			var es entries
			json.Unmarshal(bs, &es)
			performCacheUpdate(c, es.Entries)
			cursor = es.Cursor
		}()
	}

	return c.NoContent(http.StatusOK)
}

func performCacheUpdate(c echo.Context, entries []entry) {
	log := c.Logger()

	for _, e := range entries {
		log.Infof("Updating cache entry. filename=%s", e.Name)
		readFromStorage(c, e.Name)
	}
}
