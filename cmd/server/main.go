package main

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/mlesniak/markdown/internal/dropbox"
	"github.com/mlesniak/markdown/internal/handler"
	"github.com/mlesniak/markdown/internal/tags"
	"github.com/rs/zerolog"
	"github.com/ziflex/lecho/v2"
	"os"
)

func main() {
	const rootFilename = "202009010520 index.md"

	log := initializeLogger()
	dropboxService := initializeDropbox(log)

	e := initializeEcho(log)
	e.Static("/static", "static/")
	e.Static("/download", "download/")

	e.GET("/", func(c echo.Context) error {
		c.SetParamNames("name")
		c.SetParamValues(rootFilename)
		return handler.ContentHandler(c)
	})
	e.GET("/:name", handler.ContentHandler)
	e.GET("/tag/:tag", tags.Get().HandleTag)

	e.GET("/dropbox/webhook", dropboxService.HandleChallenge)
	e.POST("/dropbox/webhook", dropboxService.WebhookHandler(func(log echo.Logger, filename string, data []byte) {
		dropboxService.PreloadCache(rootFilename, "202009010533 About me.md")
	}))

	dropboxService.StartCacheQueue()
	dropboxService.PreloadCache(rootFilename, "202009010533 About me.md")

	e.Logger.Info("Starting to listen for requests")
	e.Logger.Fatal(e.Start(":8080"))
}

func initializeEcho(log *lecho.Logger) *echo.Echo {
	e := echo.New()
	e.Use(handler.BuildVersionHeader())
	e.Use(middleware.RequestID())
	e.Use(lecho.Middleware(lecho.Config{
		Logger: log,
	}))
	e.HideBanner = true
	e.HidePort = true
	e.Logger = log
	return e
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

func initializeDropbox(log echo.Logger) *dropbox.Service {
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
