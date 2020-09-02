package main

import (
	"github.com/labstack/echo/v4"
	"github.com/mlesniak/markdown/pkg/dropbox"
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
	dropboxService = dropbox.New(dropboxToken)
}

func main() {
	e := echo.New()

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
