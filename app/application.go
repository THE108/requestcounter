package app

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"

	"github.com/THE108/requestcounter/config"
	"github.com/THE108/requestcounter/utils/closer"
	"github.com/THE108/requestcounter/utils/log"

	"github.com/gorilla/mux"
)

// Application
type Application struct {
	config   *config.Config
	closer   *closer.Closer
	logger   log.ILogger
	router   *mux.Router
	handlers map[string]*HandlerInfo
	models   models
}

// NewApplication creates and initializes new instance of Application
func NewApplication() *Application {
	return &Application{}
}

// Init initiate the application
func (this *Application) init() error {
	var err error
	this.config, err = config.LoadConfigFromFile()
	if err != nil {
		return fmt.Errorf("error parse config file: %s", err.Error())
	}

	this.logger = log.New(os.Stderr, "Application", this.config.LogLevel)
	this.logger.Debug("starting application")

	this.closer = closer.NewCloser()

	if err = this.initModels(); err != nil {
		return err
	}

	this.initRoutes()

	return nil
}

// Run starts the application
func (this *Application) Run() {
	if err := this.init(); err != nil {
		this.logger.Error("error init app:", err.Error())
		return
	}

	address := net.JoinHostPort(this.config.Host, strconv.Itoa(this.config.Port))
	listener, err := net.Listen("tcp", address)
	if err != nil {
		this.logger.Error("error listen:", err.Error())
		return
	}

	this.closer.AddCloser(listener)

	done := make(chan struct{}, 1)
	go this.serve(listener, done)

	this.logger.Info("start waiting for signals")
	sig, err := this.closer.Run()
	this.logger.Info("got signal ", sig)
	this.logger.ErrorIfNotNil("error on closing", err)

	<-done
}

func (this *Application) serve(listener net.Listener, done chan<- struct{}) {
	server := &http.Server{
		Handler: this.router,
	}

	this.logger.ErrorIfNotNil("error serve (could be caused by interrapting application with ^C) ",
		server.Serve(listener))

	done <- struct{}{}
}
