package services

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

type DappConnectService struct {
	// The address the server should serve at. When only the port is given
	// (e.g. ":1323"), the server will be served at localhost. Default: ":1323"
	ServerAddr string
	// The level of log messages to include in the log, which is output to the
	// console in dev mode. Default: 2 (INFO)
	LogLevel log.Lvl
	// Hide the banner that is printed in the console when the server starts.
	// Default: no (false)
	HideServerBanner bool

	// Current Echo instance used to control the server
	echo *echo.Echo
	// If the server is currently running
	serverRunning bool
}

const DefaultServerAddr string = ":1323"

// Start sets up the server and starts it with the given address it if it has
// not been started. It does nothing if the server is running. Returns whether
// the server is currently running.
func (dc *DappConnectService) Start() bool {
	/*
	 * Resources about using Context to create a graceful server start/stop functionality
	 * - https://medium.com/@jamal.kaksouri/the-complete-guide-to-context-in-golang-efficient-concurrency-management-43d722f6eaea
	 * - https://echo.labstack.com/docs/cookbook/graceful-shutdown
	 * - https://thegodev.com/graceful-shutdown/
	 */

	// Do nothing if the server is already running
	if dc.serverRunning {
		dc.echo.Logger.Warn("Attempted to start a server is already running")
		return dc.serverRunning
	}

	// Set server log level to default if none was specified
	if dc.LogLevel == 0 {
		dc.LogLevel = log.INFO
	}

	// Setup
	dc.echo = echo.New()
	dc.echo.Logger.SetLevel(dc.LogLevel)
	dc.echo.HideBanner = dc.HideServerBanner
	setupServerRoutes(dc.echo)

	// Allow for the server to be gracefully stop if there was an interrupt
	// (e.g. Ctrl+C)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	dc.serverRunning = true

	// Start server
	go func(ctx context.Context) {
		select {
		case <-ctx.Done():
			dc.Stop()
		default:
			// Set server address to default if none was specified
			if dc.ServerAddr == "" {
				dc.ServerAddr = DefaultServerAddr
			}

			// NOTE: echo.Start() function does not end until the process running it is killed
			if err := dc.echo.Start(dc.ServerAddr); err != nil && err != http.ErrServerClosed {
				dc.serverRunning = false
				dc.echo.Logger.Fatal("Unexpected error occurred in starting the server")
			}

			stop()
		}
	}(ctx)

	return true
}

// Stop gracefully stops the server if it is running. It does nothing if the
// server is not running. Returns whether the server is currently running.
func (dc *DappConnectService) Stop() bool {
	if !dc.serverRunning {
		dc.echo.Logger.Warn("Attempted to shut down a server that is not running")
		return dc.serverRunning
	}

	dc.echo.Logger.Info("Shutting down server...")
	//gracefully shutdown the server with a timeout of 10 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := dc.echo.Shutdown(ctx); err != nil {
		dc.echo.Logger.Fatal(err)
		return dc.serverRunning
	}

	dc.echo.Logger.Info("Server has been shut down")
	dc.serverRunning = false

	return dc.serverRunning
}

// IsOn gives whether or not the server is on and running
func (dc *DappConnectService) IsOn() bool {
	return dc.serverRunning
}

// setupServerRoutes declares the server routes
func setupServerRoutes(e *echo.Echo) {
	// Set up CORS
	e.Use(middleware.CORS())

	e.GET("/", func(c echo.Context) error {
		return c.JSON(http.StatusOK, "OK")
	})

	e.POST("/wallet/init", func(c echo.Context) error {
		return c.JSON(http.StatusOK, "Init wallet")
	})
}
