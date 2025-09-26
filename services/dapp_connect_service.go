package services

import (
	"context"
	"crypto/ecdh"
	"encoding/base64"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/awnumar/memguard"
	"github.com/hiyosi/hawk"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/wailsapp/wails/v3/pkg/application"

	dc "duckysigner/internal/dapp_connect"
	"duckysigner/internal/dapp_connect/handlers"
	"duckysigner/internal/dapp_connect/session"
	"duckysigner/internal/wallet_session"
)

// DappConnectService is a Wails binding allows for a Wails frontend to interact
// with and manage the dApp connect server
type DappConnectService struct {
	// The address the server should serve at. When only the port is given
	// (e.g. ":1323"), the server will be served at localhost.
	// Default: ":1323"
	ServerAddr string
	// The level of log messages to include in the log, which is output to the
	// console in dev mode.
	// Default: 2 (INFO)
	LogLevel log.Lvl
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
	ECDHCurve dc.ECDHCurve
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
		dcs.echo.Logger.Warn("Attempted to start a server is already running")
		return dcs.serverRunning
	}

	// Set server log level to default if none was specified
	if dcs.LogLevel == 0 {
		dcs.LogLevel = log.INFO
	}

	// Setup
	dcs.echo = echo.New()
	dcs.echo.Logger.SetLevel(dcs.LogLevel)
	dcs.echo.HideBanner = dcs.HideServerBanner
	dcs.echo.HidePort = dcs.HideServerPort
	// Set ECDH curve if it is not set
	if dcs.ECDHCurve == nil {
		dcs.ECDHCurve = ecdh.X25519()
	}
	// Set up other stuff
	dc.SetupCustomValidator(dcs.echo)
	dcs.setupServerRoutes(dcs.echo)

	// Allow for the server to be gracefully stop if there was an interrupt
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
				dcs.echo.Logger.Error(err)
				dcs.echo.Logger.Fatal("Unexpected error occurred in starting the server")
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
		dcs.echo.Logger.Warn("Attempted to shut down a server that is not running")
		return dcs.serverRunning
	}

	dcs.echo.Logger.Info("Shutting down server...")
	//gracefully shutdown the server with a timeout of 10 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := dcs.echo.Shutdown(ctx); err != nil {
		dcs.echo.Logger.Fatal(err)
		return dcs.serverRunning
	}

	dcs.echo.Logger.Info("Server has been shut down")
	dcs.serverRunning = false

	return dcs.serverRunning
}

// IsOn gives whether or not the server is on and running
func (dcs *DappConnectService) IsOn() bool {
	return dcs.serverRunning
}

// CleanUp end memory enclave sessions and release resources used by this dApp
// connection service
//
// FOR THE BACKEND ONLY
func (dcs *DappConnectService) CleanUp() {
	// Purge the MemGuard session
	defer memguard.Purge()
	// XXX: May do other stuff (e.g. end db session properly) to clean up in the future
}

// setupServerRoutes declares the server routes
func (dcs *DappConnectService) setupServerRoutes(e *echo.Echo) {
	// Set up CORS
	e.Use(middleware.CORS())

	walletSession := dcs.KMDService.Session()

	if walletSession == nil {
		e.GET("/", handlers.RootGet(dcs.WailsApp))
		// Don't bother setting up other routes
		return
	}

	sessionManager := session.NewManager(dcs.ECDHCurve, &session.SessionConfig{
		DataDir:             walletSession.FilePath,
		ApprovalTimeoutSecs: dcs.ApprovalTimeout,
	})

	sessionCredStore := SessionCredentialStore{
		WalletSession:  walletSession,
		SessionManager: sessionManager,
	}

	e.GET("/", handlers.RootGet(dcs.WailsApp),
		HawkMiddleware(
			dcs.echo,
			&HawkMiddlewareOptions{
				Required:        false,
				CredentialStore: sessionCredStore,
			},
		),
	)

	e.POST("/session/init", handlers.SessionInitPost(
		dcs.echo, dcs.KMDService.Session(), sessionManager, dcs.ECDHCurve,
	))
}

// HawkMiddlewareOptions is the set of options for the Hawk middleware
type HawkMiddlewareOptions struct {
	// If authentication is required
	Required bool
	// The store that contains a function for retrieving credentials.
	CredentialStore hawk.CredentialStore
}

// HawkMiddleware is a middleware function that enables Hawk authentication on a
// route
func HawkMiddleware(echoInstance *echo.Echo, opt *HawkMiddlewareOptions) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			var req = c.Request()

			if req.Header.Get("Authorization") == "" && !opt.Required {
				echoInstance.Logger.Debug(
					"No Authorization header. Skipping Hawk authentication because it is not required.",
				)
				next(c)
				return nil
			}

			echoInstance.Logger.Debug("Authenticating Hawk request header")

			hawkServer := hawk.NewServer(opt.CredentialStore)

			// Authenticate the Hawk request header
			cred, err := hawkServer.Authenticate(req)
			if err != nil {
				// Set WWW-Authenticate header
				c.Response().Header().Set("WWW-Authenticate", "Hawk")
				// Respond with 401 Unauthorized
				return c.JSON(http.StatusUnauthorized, dc.ApiError{
					Name:    "auth_request_failed",
					Message: err.Error(),
				})
			}

			// Continue to do other request stuff
			next(c)

			echoInstance.Logger.Debug("Creating Hawk response header")

			// Add Hawk response header after the request is finished
			hawkRespHeader, err := hawkServer.Header(req, cred, &hawk.Option{})
			if err != nil {
				return c.JSON(http.StatusInternalServerError, dc.ApiError{
					Name:    "auth_response_failed",
					Message: err.Error(),
				})
			}
			c.Response().Header().Set("Server-Authorization", hawkRespHeader)

			return nil
		}
	}
}

// SessionCredentialStore is a Hawk CredentialStore use for retrieving dApp
// connect session credentials
type SessionCredentialStore struct {
	WalletSession  *wallet_session.WalletSession
	SessionManager *session.Manager
}

// GetCredential returns the a set of Hawk credentials by retrieving stored
// session data
func (store SessionCredentialStore) GetCredential(id string) (*hawk.Credential, error) {
	// The wallet session's master key is needed to read the dApp connect
	// session database file
	mek, err := store.WalletSession.GetMasterKey()
	if err != nil {
		return nil, err
	}

	session, err := store.SessionManager.GetSession(id, mek)
	if err != nil {
		return nil, err
	}

	// Derive shared key
	sharedKey, err := session.SharedKey()
	if err != nil {
		return nil, err
	}

	return &hawk.Credential{
		ID:  id,
		Key: base64.StdEncoding.EncodeToString(sharedKey),
		Alg: hawk.SHA256,
	}, nil
}
