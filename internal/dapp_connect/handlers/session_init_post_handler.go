package handlers

import (
	"encoding/base64"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/wailsapp/wails/v3/pkg/application"

	. "duckysigner/internal/dapp_connect"
)

// SessionInitPostHandler is the route handler for `POST /session/init`
func SessionInitPostHandler(
	echoInstance *echo.Echo,
	wailsApp *application.App,
	userResponseTimeout time.Duration,
	ecdhCurve ECDHCurve,
) func(echo.Context) error {
	return func(c echo.Context) error {
		// Read request data
		dappInfo := new(DAppInfo)
		if err := c.Bind(dappInfo); err != nil {
			return c.JSON(http.StatusBadRequest, ApiError{
				Name:    "bad_request",
				Message: err.Error(),
			})
		}

		echoInstance.Logger.Debug("Incoming request:", dappInfo)

		dAppIdPk, dappIdApiErr, err := ValidateDAppID(dappInfo.DAppID, ecdhCurve)
		if err != nil {
			echoInstance.Logger.Fatal(err)
			return c.JSON(http.StatusBadRequest, dappIdApiErr)
		}

		// TODO: Check if session exists with dApp ID
		// - Reject if session with dApp ID exists?
		// - Ask user if they want new session with dApp with ID?

		// Prompt user to approve wallet connection session
		userResp, err := PromptUI(
			dappInfo,
			WCSessionInitUIPromptEventName,
			WCSessionInitUIRespEventName,
			wailsApp,
			echoInstance.Logger,
		)
		// Remove listener for UI response event when the server request ends,
		// which is definitely after the UI response event data is received from
		// the channel
		defer wailsApp.OffEvent(WCSessionInitUIRespEventName)

		// TODO: Handle error from prompting UI *after* setting up the removal of the UI response event listener

		// Wait for user response...
		select {
		case <-time.After(userResponseTimeout): // Time ran out
			echoInstance.Logger.Info("Ran out of time waiting for user response")
			return c.JSON(
				http.StatusRequestTimeout,
				ApiError{Name: "session_no_response", Message: "User did not respond"},
			)
		case accounts := <-userResp: // Got user's response
			// If no accounts were approved, that means the user rejected
			if len(accounts) == 0 {
				return c.JSON(
					http.StatusForbidden,
					ApiError{Name: "session_rejected", Message: "Session was rejected"},
				)
			}

			// TODO: Wrap session key in Memguard enclave.

			// Create session key pair
			// TODO: Rename sessionSk -> sessionKey (or sessionKeySk?)
			sessionId, sessionSk, err := CreateWCSessionKeyPair(ecdhCurve)
			if err != nil {
				echoInstance.Logger.Fatal(err)
				return c.JSON(http.StatusInternalServerError, ApiError{
					Name:    "session_create_fail",
					Message: "Failed to create server session",
				})
			}

			// Store connection session data
			wcStoreErr := StoreWCSessionData(sessionId, sessionSk, dAppIdPk, echoInstance.Logger)
			if wcStoreErr != nil {
				echoInstance.Logger.Fatal(wcStoreErr)
				return c.JSON(http.StatusInternalServerError, ApiError{
					Name:    "session_create_fail",
					Message: "Failed to create server session",
				})
			}

			// Create and respond with Hawk credentials
			return c.JSON(http.StatusOK, HawkCredentials{
				Algorithm: "sha256",
				// TODO: Create token (e.g. JWT) to use as ID (Maybe?)
				ID: dappInfo.DAppID,
				// The dApp will have to derive the real shared key using its private key and this session ID
				Key: base64.StdEncoding.EncodeToString(sessionId.Bytes()),
			})
		}
	}
}
