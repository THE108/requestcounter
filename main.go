package main

import (
	"os"

	"github.com/THE108/requestcounter/app"
	"github.com/THE108/requestcounter/log"
)

func main() {
	logger := log.New(os.Stderr, "main", log.DEBUG)
	logger.Debug("starting application")

	application := app.NewApplication()

	if err := application.Init(); err != nil {
		logger.Error("error init app:", err.Error())
		return
	}

	application.Run()
}
