package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"clusterviz/api"
	"clusterviz/internal/pkg/configurations"
	"clusterviz/internal/pkg/handler"
)

type API struct {
	conf        *configurations.ServiceConfigurations
	eAccHandler *handler.EndPointHandler
	Router      *gin.Engine
	server      *http.Server

	mutex         sync.Mutex
	isRunning     bool
	name          string
	waitTimeInSec int64
}

func New() (*API, error) {
	conf, err := configurations.GetServiceConfigurations()
	if err != nil {
		return nil, err
	}

	eAccHandler, err := handler.NewEndPointHandler(conf)
	if err != nil {
    	log.Error("Error creating EndPointHandler:", err)
    	return nil,err
	}

	return &API{
		conf:          conf,
		eAccHandler:   eAccHandler,
		mutex:         sync.Mutex{},
		isRunning:     false,
		name:          "ClusterViz Server",
		waitTimeInSec: 10,
	}, nil
}

func (app *API) Start() {
	app.Router = gin.New()

	// Add CORS middleware
	config := cors.DefaultConfig()
	app.Router.Use(cors.New(config))

	api.RegisterHandlersWithOptions(app.Router, app.eAccHandler,
		api.GinServerOptions{BaseURL: "/api/v1", Middlewares: []api.MiddlewareFunc{api.MiddlewareFunc(handler.Authenticator())}})
	log.Infof("Starting %s Server...", app.name)
	app.startGinServer(app.conf)
	log.Infof("%s server started successfully at %s...", app.name, app.conf.Port)

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Infof("Shutting down %s server...", app.name)
	app.StopServer()
}

func (app *API) startGinServer(conf *configurations.ServiceConfigurations) {
	app.server = &http.Server{
		Addr:              fmt.Sprintf(":%s", conf.Port),
		Handler:           app.Router,
		ReadHeaderTimeout: time.Second * time.Duration(conf.HeaderReadTimeout),
	}

	// Initializing the server in a goroutine so that it won't block the graceful shutdown handling below
	go func() {
		app.mutex.Lock()
		app.isRunning = true
		app.mutex.Unlock()

		if err := app.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Errorf("listen: %s\n", err)
			os.Exit(1)
		}

		app.mutex.Lock()
		app.isRunning = false
		app.mutex.Unlock()
	}()
}

func (app *API) StopServer() {
	// The context is used to inform the server it has 5 seconds to finish the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(app.waitTimeInSec)*time.Second)
	defer cancel()

	if err := app.server.Shutdown(ctx); err != nil {
		log.Errorf("Server Shutdown: %v", err)
	}

	log.Infof("Server stopped successfully ...")
}

// IsRunning returns true if the server is Listening, false otherwise.
func (app *API) IsRunning() bool {
	app.mutex.Lock()
	defer app.mutex.Unlock()

	return app.isRunning
}

