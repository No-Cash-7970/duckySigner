package services

import (
	"context"
	"crypto/ecdh"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/awnumar/memguard"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/wailsapp/wails/v3/pkg/application"

	dc "duckysigner/internal/dapp_connect"
	"duckysigner/internal/dapp_connect/handlers"
	"duckysigner/internal/dapp_connect/session"
	"duckysigner/internal/tools"
)

// DappConnectService is a Wails binding allows for a Wails frontend to interact
// with and manage the dApp connect server
type DappConnectService struct {
	// The address the server should serve at. When only the port is given
	// (e.g. ":1323"), the server will be served at localhost.
	// Default: ":1323"
	ServerAddr string
	// The level of log messages to include in the server log, which is output
	// to the console in dev mode.
	// Default: 2 (INFO)
	ServerLogLevel log.Lvl
	// Hide the banner that is printed in the console when the server starts.
	// Default: no (false)
	HideServerBanner bool
	// Hide the listener port message that is printed in the console when the
	// server starts.
	// Default: no (false)
	HideServerPort bool
	// The current running instance of a the Wails app. Used to trigger and
	// listen to events from the UI.
	WailsApp *application.App
	// Length of time to wait for a user response for certain actions (e.g.
	// approving a session)
	UserResponseTimeout time.Duration
	// Instance of the an ECDH curve to be used for generating the wallet
	// session key pair. Typically used to set a mock curve when
	// testing.
	// Default: `ecdh.X25519()` from the `crypto/ecdh` package
	ECDHCurve tools.ECDHCurve
	// Instance of the KMD service that is being used to access the wallets
	KMDService *KMDService
	// Amount of time (in seconds) to wait for user approval of a session
	ApprovalTimeout uint64

	// Current Echo instance used to control the server
	echo *echo.Echo
	// If the server is currently running
	serverRunning bool
}

// Start sets up the server and starts it with the given address it if it has
// not been started. It does nothing if the server is running. Returns whether
// the server is currently running.
func (dcs *DappConnectService) Start() bool {
	/*
	 * Resources about using Context to create a graceful server start/stop functionality
	 * - https://medium.com/@jamal.kaksouri/the-complete-guide-to-context-in-golang-efficient-concurrency-management-43d722f6eaea
	 * - https://echo.labstack.com/docs/cookbook/graceful-shutdown
	 * - https://thegodev.com/graceful-shutdown/
	 */

	// Do nothing if the server is already running
	if dcs.serverRunning {
		log.Warn("Attempted to start a server is already running")
		return dcs.serverRunning
	}

	// Set server log level to default if none was specified
	if dcs.ServerLogLevel == 0 {
		dcs.ServerLogLevel = log.INFO
	}

	// Setup
	dcs.echo = echo.New()
	dcs.echo.Logger.SetLevel(dcs.ServerLogLevel)
	dcs.echo.HideBanner = dcs.HideServerBanner
	dcs.echo.HidePort = dcs.HideServerPort
	// Set ECDH curve if it is not set
	if dcs.ECDHCurve == nil {
		dcs.ECDHCurve = ecdh.X25519()
	}
	// Set up other stuff
	dc.SetupCustomValidator(dcs.echo)
	dcs.setupServerRoutes(dcs.echo)

	// Allow for the server to gracefully stop if there was an interrupt
	// (e.g. Ctrl+C)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	dcs.serverRunning = true

	// Start server
	go func(ctx context.Context) {
		select {
		case <-ctx.Done():
			dcs.Stop()
		default:
			// Set server address to default if none was specified
			if dcs.ServerAddr == "" {
				dcs.ServerAddr = dc.DefaultServerAddr
			}

			// NOTE: echo.Start() function does not end until the process running it is killed
			if err := dcs.echo.Start(dcs.ServerAddr); err != nil && err != http.ErrServerClosed {
				dcs.serverRunning = false
				log.Error(err)
				log.Fatal("Unexpected error occurred in starting the server")
			}

			stop()
		}
	}(ctx)

	return true
}

// Stop gracefully stops the server if it is running. It does nothing if the
// server is not running. Returns whether the server is currently running.
func (dcs *DappConnectService) Stop() bool {
	if !dcs.serverRunning {
		log.Warn("Attempted to shut down a server that is not running")
		return dcs.serverRunning
	}

	log.Info("Shutting down server...")
	//gracefully shutdown the server with a timeout of 10 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := dcs.echo.Shutdown(ctx); err != nil {
		dcs.echo.Logger.Fatal(err)
		return dcs.serverRunning
	}

	log.Info("Server has been shut down")
	dcs.serverRunning = false

	return dcs.serverRunning
}

// IsOn gives whether or not the server is on and running
func (dcs *DappConnectService) IsOn() bool {
	return dcs.serverRunning
}

// CleanUp performs various operations to clean up and release resources used by
// the dApp connection service.
// FOR THE BACKEND ONLY
func (dcs *DappConnectService) CleanUp() error {
	/* Purge the MemGuard session */
	memguard.Purge()

	// XXX: May do other stuff (e.g. end db session properly) to clean up in the future

	return nil
}

// setupServerRoutes declares the server routes
func (dcs *DappConnectService) setupServerRoutes(e *echo.Echo) {
	// Set up CORS
	e.Use(middleware.CORS())

	walletSession := dcs.KMDService.Session()

	if walletSession == nil {
		e.GET("/", handlers.RootGet(dcs.echo, nil, nil))
		// Don't bother setting up other routes
		return
	}

	sessionManager := session.NewManager(dcs.ECDHCurve, &session.SessionConfig{
		DataDir:             walletSession.FilePath,
		ApprovalTimeoutSecs: dcs.ApprovalTimeout,
	})

	e.GET("/", handlers.RootGet(
		dcs.echo, walletSession, sessionManager,
	))
	e.POST("/session/init", handlers.SessionInitPost(
		dcs.echo, walletSession, sessionManager, dcs.ECDHCurve,
	))
	e.POST("/session/confirm", handlers.SessionConfirmPost(
		dcs.echo, dcs.WailsApp, walletSession, sessionManager, dcs.ECDHCurve,
	))
	e.GET("/session/end", handlers.SessionEndGet(
		dcs.echo, walletSession, sessionManager,
	))
	e.POST("/transaction/sign", handlers.TransactionSignPost(
		dcs.echo, dcs.WailsApp, walletSession, sessionManager, dcs.ECDHCurve,
	))
}
