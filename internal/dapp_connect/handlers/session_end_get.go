package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"

	dc "duckysigner/internal/dapp_connect"
	mw "duckysigner/internal/dapp_connect/middleware"
	"duckysigner/internal/dapp_connect/session"
	"duckysigner/internal/wallet_session"
)

// SessionEndGet is the route handler for `GET /session/end`
func SessionEndGet(
	echoInstance *echo.Echo,
	walletSession *wallet_session.WalletSession,
	sessionManager *session.Manager,
) echo.HandlerFunc {
	return func(c echo.Context) error {
		hawkOpt := mw.HawkOptions{
			EchoContext:  c,
			EchoInstance: echoInstance,
			CredentialStore: mw.SessionCredentialStore{
				WalletSession:  walletSession,
				SessionManager: sessionManager,
			},
		}
		// Hawk authentication
		hawkServer, cred, apiErr := mw.HawkAuth(nil, &hawkOpt)
		if apiErr != nil {
			// Set WWW-Authenticate header
			c.Response().Header().Set("WWW-Authenticate", "Hawk")
			// Respond with 401 Unauthorized
			return mw.HawkRespJSON(http.StatusUnauthorized, apiErr, hawkServer, cred, &hawkOpt)
		}

		// Retrieve the master encryption key of the currently opened wallet
		mek, err := walletSession.GetMasterKey()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, dc.ApiError{
				Name:    confirmCreateFailName,
				Message: confirmCreateFailMsg,
			})
		}

		// Remove the session
		err = sessionManager.RemoveSession(cred.ID, mek)
		if err != nil {
			echoInstance.Logger.Error(err.Error())
		}

		return c.JSON(http.StatusOK, "OK")
	}
}
