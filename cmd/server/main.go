package main

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/mlesniak/markdown/pkg/dropbox"
	"github.com/ziflex/lecho/v2"
	"io/ioutil"
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

	// Normal hook. Simply print it.
	bs, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		panic(err)
	}
	defer c.Request().Body.Close()
	println(bs)

	return c.NoContent(http.StatusOK)
}

func dropboxUpdate(c echo.Context) error {
	// Normal hook. Simply print it.
	bs, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		panic(err)
	}
	defer c.Request().Body.Close()
	println(bs)

	return c.NoContent(http.StatusOK)
}
