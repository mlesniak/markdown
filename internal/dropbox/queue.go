package dropbox

import (
	"github.com/labstack/echo/v4"
)

type Queue struct {
	log   echo.Logger
	queue chan string
}

func (q *Queue) Add(filename string) {
	q.queue <- filename
}

func (q *Queue) Start() {
	go func() {
		for {
			filename := <-q.queue
			q.log.Infof("Updating file %s", filename)
		}
	}()
}
