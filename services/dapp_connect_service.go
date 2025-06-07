package services

import (
	"context"
	"crypto/ecdh"
	"crypto/rand"
	"encoding/base64"
	"io"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/wailsapp/wails/v3/pkg/application"
)

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
	// connection session key pair. Typically used to set a mock curve when
	// testing.
	// Default: `ecdh.X25519()` from the `crypto/ecdh` package
	ECDHCurve ECDHCurve

	// Current Echo instance used to control the server
	echo *echo.Echo
	// If the server is currently running
	serverRunning bool
}

// ECDH curve interface based on the `ecdh.Curve` interface in the `crypto/ecdh`
// package
type ECDHCurve interface {
	GenerateKey(rand io.Reader) (*ecdh.PrivateKey, error)
	NewPrivateKey(key []byte) (*ecdh.PrivateKey, error)
	NewPublicKey(key []byte) (*ecdh.PublicKey, error)
}

// A set of credentials for authenticating using Hawk
type HawkCredentials struct {
	// Authentication ID
	ID string `json:"id"`
	// Authentication key
	Key string `json:"key"`
	// The hash algorithm used to create the message authentication code (MAC).
	Algorithm string `json:"algorithm"`
}

// An error message
type ApiError struct {
	// Error name
	Name string `json:"name,omitempty"`
	// Error message
	Message string `json:"message,omitempty"`
}

// DApp information
type DAppInfo struct {
	// Name of the application connecting to wallet
	Name string `json:"name"`
	// URL for the app connecting to the wallet
	Url string `json:"url,omitempty"`
	// Description of the app
	Description string `json:"description,omitempty"`
	// Icon for the app connecting to wallet as a Base64 encoded JPEG, PNG or SVG data URI
	Icon string `json:"icon,omitempty"`
	// TODO: Document this and change the name
	DAppID string `json:"dapp_session_pk"`
}

// TODO: Document this
type ConnectionSession struct {
	DAppID           *ecdh.PublicKey
	SessionID        *ecdh.PublicKey
	ServerKey        *ecdh.PrivateKey
	SharedSessionKey []byte
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
	dc.echo.HidePort = dc.HideServerPort
	// Set ECDH curve if it is not set
	if dc.ECDHCurve == nil {
		dc.ECDHCurve = ecdh.X25519()
	}
	dc.setupServerRoutes(dc.echo)

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
				dc.echo.Logger.Error(err)
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
func (dc *DappConnectService) setupServerRoutes(e *echo.Echo) {
	// Set up CORS
	e.Use(middleware.CORS())

	e.GET("/", func(c echo.Context) error {
		dc.WailsApp.EmitEvent("session_init_response", []string{"account 1", "account 2"})
		return c.JSON(http.StatusOK, "OK")
	})

	e.POST("/session/init", func(c echo.Context) error {
		// Read request data
		dappInfo := new(DAppInfo)
		if err := c.Bind(dappInfo); err != nil {
			return c.JSON(http.StatusBadRequest, ApiError{
				Name:    "bad_request",
				Message: err.Error(),
			})
		}

		dc.echo.Logger.Debug("Incoming request: ", dappInfo)

		// Check if Wails app is properly initialized
		if dc.WailsApp == nil {
			return c.JSON(http.StatusInternalServerError, ApiError{
				Name:    "connect_service_improper_init",
				Message: "The dApp connections service was improperly initialized",
			})
		}

		// Contains the user's response to session initialization prompt
		userResp := make(chan []string)

		// Listen for event that contains user's response
		dc.WailsApp.OnEvent("session_init_response", func(e *application.CustomEvent) {
			dc.echo.Logger.Info("Got user response: ", e.Data)
			// NOTE: For some reason, the actual event data is always within a slice
			userResp <- e.Data.([]interface{})[0].([]string)
			close(userResp)
		})
		defer dc.WailsApp.OffEvent("session_init_response")

		// Prompt user to approve of connection after setting up listener for
		// user response event because a user (like an automated user in a unit
		// test) may respond before being prompted, which is not ideal.
		dc.WailsApp.EmitEvent("session_init_prompt", dappInfo)

		select {
		case <-time.After(dc.UserResponseTimeout): // Time ran out
			dc.echo.Logger.Info("Ran out of time waiting for user response")
			return c.JSON(
				http.StatusRequestTimeout,
				ApiError{"session_no_response", "User did not respond"},
			)
		case accounts := <-userResp: // Got user's response
			// If no accounts were approved, that means the user rejected
			if len(accounts) == 0 {
				return c.JSON(
					http.StatusForbidden,
					ApiError{"session_rejected", "Session was rejected"},
				)
			}

			// TODO: Create session key pair
			// sessionId, sessionSk, err := CreateSessionKeyPair(dc.ECDHCurve)
			sessionId, _, err := CreateSessionKeyPair(dc.ECDHCurve)
			if err != nil {
				dc.echo.Logger.Fatal(err)
				return c.JSON(http.StatusInternalServerError, ApiError{
					Name:    "session_create_fail",
					Message: "Failed to create server session",
				})
			}

			// TODO: Store connection session data

			// dappIdBytes, err := base64.StdEncoding.DecodeString(dappInfo.DAppID)
			// if err != nil {
			// 	dc.echo.Logger.Fatal(err)
			// 	return c.JSON(http.StatusBadRequest, ApiError{
			// 		Name:    "invalid_dapp_id_b64",
			// 		Message: "DApp ID is not a valid Base64 string",
			// 	})
			// }

			// dappId, err := curve.NewPublicKey(dappIdBytes)
			// if err != nil {
			// 	dc.echo.Logger.Fatal(err)
			// 	return c.JSON(http.StatusBadRequest, ApiError{
			// 		Name:    "invalid_dapp_id_pk",
			// 		Message: "DApp ID is invalid",
			// 	})
			// }

			// Derive shared key using private key and dApp session key
			// sharedSessionKey, err := sessionSk.ECDH(dappId)
			// if err != nil {
			// 	dc.echo.Logger.Fatal(err)
			// 	return c.JSON(http.StatusBadRequest, ApiError{
			// 		Name:    "invalid_dapp_id",
			// 		Message: "DApp ID is invalid",
			// 	})
			// }

			// connectionSession := ConnectionSession{
			//     DAppID: dappId,
			//     SessionID: sessionId,
			//     ServerKey: sessionSk,
			//     SharedSessionKey: sharedSessionKey,
			// }

			dc.echo.Logger.Info("Session init success")
			return c.JSON(http.StatusOK, HawkCredentials{
				Algorithm: "sha256",
				// TODO: Create token (e.g. JWT) to use as ID (Maybe?)
				ID: dappInfo.DAppID,
				// The dApp will have to derive the real shared key using its private key and this session ID
				Key: base64.StdEncoding.EncodeToString(sessionId.Bytes()),
			})
		}
	})
}

// TODO: document this
func CreateSessionKeyPair(curve ECDHCurve) (id *ecdh.PublicKey, sk *ecdh.PrivateKey, err error) {
	sk, err = curve.GenerateKey(rand.Reader)
	if err != nil {
		return
	}

	id = sk.PublicKey()

	return
}
