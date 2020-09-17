package main

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/mlesniak/markdown/internal/cache"
	"github.com/mlesniak/markdown/internal/dropbox"
	"github.com/mlesniak/markdown/internal/handler"
	"github.com/mlesniak/markdown/internal/markdown"
	"github.com/mlesniak/markdown/internal/tags"
	"github.com/ziflex/lecho/v2"
	"os"
)

const (
	// External mountable directory to publish files.
	downloadRoot = "download/"

	// Directory containing static files for website.
	staticRoot = "static/"

	rootFilename = "202009010520 index"
)

func main() {
	e := echo.New()

	// Configure logging.
	logger := lecho.New(
		os.Stdout,
		lecho.WithLevel(log.INFO),
		lecho.WithTimestamp(),
		lecho.WithCallerWithSkipFrameCount(3),
		lecho.WithField("commit", handler.BuildInformation()),
	)
	e.Logger = logger

	// Initialize services.
	dropboxService := initDropboxStorage()
	tagsService := tags.New()
	handlerService := handler.Handler{
		RootFilename:  rootFilename,
		StorageReader: dropboxService,
		Cache:         cache.New(),
		Tags:          tagsService,
	}

	// Preload files.
	go dropboxService.PreloadCache(e.Logger, func(log echo.Logger, filename string, data []byte) {
		// TODO This is partially the same functionality as in handler.Handler(), combine it.
		html, _ := markdown.ToHTML(log, filename, data)

		tagList := markdown.GetTags(data)
		handlerService.Tags.Update(filename, tagList)

		log.Infof("Inital cache population for filename=%s", filename)
		handlerService.Cache.Add(cache.Entry{
			Name: filename,
			Data: []byte(html),
		})
	})

	// Configure middlewares.
	e.Use(handler.BuildVersionHeader())
	e.Use(middleware.RequestID())
	e.Use(lecho.Middleware(lecho.Config{
		Logger: logger,
	}))

	// Serve static and downloadable files.
	e.Static("/static", staticRoot)
	e.Static("/download", downloadRoot)

	// Serve dynamic files.
	e.GET("/", handlerService.Handle)
	e.GET("/tag/:tag", tagsService.HandleTag)
	e.GET("/:name", handlerService.Handle)

	// Handle cache invalidation through dropbox webhooks.
	e.GET("/dropbox/webhook", dropboxService.HandleChallenge)
	e.POST("/dropbox/webhook", dropboxService.WebhookHandler(func(log echo.Logger, filename string, data []byte) {
		html, _ := markdown.ToHTML(log, filename, data)
		log.Infof("Update cache after webhook for filename=%s", filename)
		handlerService.Cache.Add(cache.Entry{
			Name: filename,
			Data: []byte(html),
		})
	}))

	// Start server.
	e.HideBanner = true
	e.HidePort = true
	e.Logger.Fatal(e.Start(":8080"))
}

// initDropboxStorage initializes the dropbox service by defining the
// dropbox developer token.
func initDropboxStorage() *dropbox.Service {
	dropboxToken := os.Getenv("TOKEN")
	if dropboxToken == "" {
		panic("No dropbox token set, aborting.")
	}
	dropboxAppSecret := os.Getenv("SECRET")
	if dropboxAppSecret == "" {
		panic("No dropbox app secret set, aborting.")
	}

	// TODO configurable list
	preloads := []string{
		"202009010520 index",
		"202009010533 about",
	}

	// TODO Struct instead of parameter list.
	return dropbox.New(dropboxAppSecret, dropboxToken, "notes/", preloads)
}
