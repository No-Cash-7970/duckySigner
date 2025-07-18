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

		// Validate dApp ID before doing anything else
		dappIdPk, dappIdApiErr, err := ValidateDappID(dappInfo.DappId, ecdhCurve)
		if err != nil {
			echoInstance.Logger.Error(err)
			return c.JSON(http.StatusBadRequest, dappIdApiErr)
		}

		// TODO: Check if session exists with dApp ID
		// - Reject if session with dApp ID exists?
		// - Ask user if they want new session with dApp with ID?

		// Prompt user to approve dApp connect session
		userResp, err := PromptUI(
			dappInfo,
			DCSessionInitUIPromptEventName,
			DCSessionInitUIRespEventName,
			wailsApp,
			echoInstance.Logger,
		)
		// Remove listener for UI response event when the server request ends,
		// which is definitely after the UI response event data is received from
		// the channel
		defer wailsApp.Event.Off(DCSessionInitUIRespEventName)

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

			// It's now safe to start creating a new connect session
			dcSession := DappConnectSession{DappId: dappIdPk}

			// Create session key pair and add it into the connect session
			if err := CreateDCSessionKeyPair(&dcSession, ecdhCurve); err != nil {
				echoInstance.Logger.Error(err)
				return c.JSON(http.StatusInternalServerError, ApiError{
					Name:    "session_create_fail",
					Message: "Failed to create server session",
				})
			}

			// Store connect session data for use in other server requests later on
			if err := StoreDCSessionData(&dcSession, ecdhCurve, echoInstance.Logger); err != nil {
				echoInstance.Logger.Error(err)
				return c.JSON(http.StatusInternalServerError, ApiError{
					Name:    "session_create_fail",
					Message: "Failed to create server session",
				})
			}

			// Create and respond with Hawk credentials
			return c.JSON(http.StatusOK, HawkCredentials{
				Algorithm: "sha256",
				// TODO: Create token (e.g. JWT) to use as ID (Maybe?)
				ID: dappInfo.DappId,
				// The dApp will have to derive the real shared key, but it will
				// need this session ID along with its private dApp key
				Key: base64.StdEncoding.EncodeToString(dcSession.SessionID.Bytes()),
			})
		}
	}
}
