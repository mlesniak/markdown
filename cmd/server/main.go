package main

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/mlesniak/markdown/internal/cache"
	"github.com/mlesniak/markdown/internal/dropbox"
	"github.com/mlesniak/markdown/internal/handler"
	"github.com/ziflex/lecho/v2"
	"os"
)

const (
	// External mountable directory to publish files.
	downloadRoot = "download/"

	// Directory containing static files for website.
	staticRoot = "static/"
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
	cacheService := cache.New()

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
	h := handler.Handler{
		RootFilename:  "202009010520 index",
		StorageReader: dropboxService,
		Cache:         cacheService,
	}
	e.GET("/", h.Handle)
	e.GET("/:name", h.Handle)

	// Handle cache invalidation through dropbox webhooks.
	e.GET("/dropbox/webhook", dropboxService.HandleChallenge)
	e.POST("/dropbox/webhook", dropboxService.WebhookHandler(func(log echo.Logger, filename string, data []byte) {
		h.RenderFile(log, filename, data)
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
	return dropbox.New(dropboxAppSecret, dropboxToken, "notes/")
}
