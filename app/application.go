package app

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"

	"github.com/THE108/requestcounter/closer"
	"github.com/THE108/requestcounter/config"
	"github.com/THE108/requestcounter/log"

	"github.com/gorilla/mux"
)

// Application
type Application struct {
	config   *config.Config
	closer   *closer.Closer
	router   *mux.Router
	handlers map[string]*HandlerInfo
	models   models
}

// NewApplication creates and initializes new instance of Application
func NewApplication() *Application {
	return &Application{}
}

// Init initiate the application
func (this *Application) Init() error {
	var err error
	this.config, err = config.LoadConfigFromFile()
	if err != nil {
		return fmt.Errorf("error parse config file: %s", err.Error())
	}

	this.closer = closer.NewCloser()

	if err = this.initModels(); err != nil {
		return err
	}

	this.initRoutes()

	return nil
}

// Run starts the application
func (this *Application) Run() {
	logger := log.New(os.Stderr, "main", log.ERROR)

	address := net.JoinHostPort(this.config.Host, strconv.Itoa(this.config.Port))
	listener, err := net.Listen("tcp", address)
	logger.ErrorIfNotNil("error listen", err)

	this.closer.AddCloser(listener)

	done := make(chan struct{}, 1)
	go this.serve(listener, logger, done)

	logger.Error("before closer run")

	this.closer.Run()

	<-done
}

func (this *Application) serve(listener net.Listener, logger log.ILogger, done chan<- struct{}) {
	server := &http.Server{
		Handler: this.router,
	}
	logger.ErrorIfNotNil("error serve (could be caused by interrapting application with ^C) ", server.Serve(listener))
	done <- struct{}{}
}
