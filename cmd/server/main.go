package main

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/mlesniak/markdown/internal/backlinks"
	"github.com/mlesniak/markdown/internal/cache"
	"github.com/mlesniak/markdown/internal/dropbox"
	"github.com/mlesniak/markdown/internal/handler"
	"github.com/mlesniak/markdown/internal/tags"
	"github.com/rs/zerolog"
	"github.com/ziflex/lecho/v2"
	"os"
)

const (
	// External mountable directory to publish files.
	downloadRoot = "download/"

	// Directory containing static files for website.
	staticRoot = "static/"

	rootFilename = "202009010520 index.md"
)

func main() {
	log := initializeLogger()

	// TODO use proper singletons here instead of services?
	tagsService := tags.Get()
	cacheService := cache.Get()
	dropboxService := initializeDropboxStorage(log)
	handlerService := handler.Handler{
		Cache: cacheService,
	}

	e := echo.New()
	e.Use(handler.BuildVersionHeader())
	e.Use(middleware.RequestID())
	e.Use(lecho.Middleware(lecho.Config{
		Logger: log,
	}))
	e.Logger = log
	e.Logger.SetHeader("")
	e.Logger.Info("")

	// Serve static and downloadable files.
	e.Static("/static", staticRoot)
	e.Static("/download", downloadRoot)

	// Serve normal markdown files.
	e.GET("/", func(c echo.Context) error {
		c.SetParamNames("name")
		c.SetParamValues(rootFilename)
		return handlerService.Handle(c)
	})
	e.GET("/:name", handlerService.Handle)

	e.GET("/tag/:tag", tagsService.HandleTag)

	// Handle cache invalidation through dropbox webhooks.
	e.GET("/dropbox/webhook", dropboxService.HandleChallenge)
	e.POST("/dropbox/webhook", dropboxService.WebhookHandler(func(log echo.Logger, filename string, data []byte) {
		// updateFile(log, filename, data, tagsService, cacheService)
		// TODO Remove pre-loading data, will be done in the queue.
		// dropboxService.UpdateFile(filename)
	}))

	dropboxService.Start()
	dropboxService.PreloadCache(rootFilename, "202009010533 About me.md")

	// Start server.
	e.HideBanner = true
	e.HidePort = true
	e.Logger.Info("Starting to listen for requests")
	e.Logger.Fatal(e.Start(":8080"))
}

func initializeLogger() *lecho.Logger {
	if os.Getenv("LOCAL") != "" {
		return lecho.New(
			zerolog.ConsoleWriter{
				Out: os.Stderr,
				PartsOrder: []string{
					zerolog.LevelFieldName,
					zerolog.TimestampFieldName,
					zerolog.MessageFieldName,
				},
			},
			lecho.WithLevel(log.INFO),
			lecho.WithTimestamp(),
			lecho.WithCallerWithSkipFrameCount(3),
		)
	}

	return lecho.New(
		os.Stdout,
		lecho.WithLevel(log.INFO),
		lecho.WithTimestamp(),
		lecho.WithCallerWithSkipFrameCount(3),
		lecho.WithField("commit", handler.BuildInformation()),
	)
}

func initializeDropboxStorage(log echo.Logger) *dropbox.Service {
	dropboxToken := os.Getenv("TOKEN")
	if dropboxToken == "" {
		panic("No dropbox token set, aborting.")
	}
	dropboxAppSecret := os.Getenv("SECRET")
	if dropboxAppSecret == "" {
		panic("No dropbox app secret set, aborting.")
	}

	return dropbox.New(dropbox.Service{
		AppSecret:     dropboxAppSecret,
		Token:         dropboxToken,
		RootDirectory: "notes/",
		Log:           log,
	})
}

func initializeCache(e *echo.Echo, dropboxService *dropbox.Service, tagsService *tags.Tags, cacheService *cache.Cache, backlinkService *backlinks.Backlinks) {
	// dropboxService.PreloadCache(e.Logger,
	// 	func(log echo.Logger, filename string, data []byte) {
	// 		updateFile(log, filename, data, tagsService, cacheService)
	// 	},
	// 	func() {
	// 		// Update backlinks.
	// 		for _, name := range cacheService.List() {
	// 			html, _ := cacheService.GetEntry(name)
	// 			links := utils.GetLinks(html)
	// 			backlinkService.AddTargets(name, links)
	// 		}
	// 	})
}

// TODO Move this to a more appropriate place.
// updateFile receives a markdown file, renders its HTML, updates the tag list
// and updates the corresponding cache entry.
// func updateFile(log echo.Logger, filename string, data []byte, tagsService *tags.Tags, cacheService *cache.Cache) {
// 	// Just to be sure we do not accidentally serve a non-public, but linked file.
// 	if !isPublic(data) {
// 		log.Warnf("Preventing caching of non-public filename=%s", filename)
// 		return
// 	}
//
// 	// Update tag list.
// 	tagList := utils.GetTags(data)
// 	tagsService.Update(filename, tagList)
//
// 	// Render file.
// 	html, _ := markdown.ToHTML(log, data)
//
// 	// Populate cache
// 	log.Infof("Update cache for filename=%s", filename)
// 	cacheService.AddEntry(cache.Entry{
// 		Name: filename,
// 		Data: []byte(html),
// 	})
// }
