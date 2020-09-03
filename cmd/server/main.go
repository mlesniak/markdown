package main

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/mlesniak/markdown/pkg/dropbox"
	"github.com/ziflex/lecho/v2"
	"os"
)

const (
	// Default title if the title can not be extracted from the markdown file.
	defaultTitle = "mlesniak.com"

	// Name of the root file if no filename is specified.
	rootFilename = "202009010520 index"

	// Tag name to define markdown files which are allowed to be published.
	publishTag = "#public"
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

	// Serve static files.
	e.Static("/static", "static")

	// Serve dynamic files.
	e.GET("/", handle)
	e.GET("/:name", handle)

	// Start server.
	e.HideBanner = true
	e.HidePort = true
	e.Logger.Fatal(e.Start(":8080"))
}
