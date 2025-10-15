package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"

	mw "duckysigner/internal/dapp_connect/middleware"
	"duckysigner/internal/dapp_connect/session"
	"duckysigner/internal/wallet_session"
)

// RootGet is the route handler for `GET /`
func RootGet(
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
			OptionalAuth: true,
		}
		// Hawk authentication (optional)
		hawkServer, cred, apiErr := mw.HawkAuth(nil, &hawkOpt)
		if apiErr != nil {
			// Set WWW-Authenticate header
			c.Response().Header().Set("WWW-Authenticate", "Hawk")
			// Respond with 401 Unauthorized
			return mw.HawkRespJSON(http.StatusUnauthorized, apiErr, hawkServer, cred, &hawkOpt)
		}

		return mw.HawkRespJSON(http.StatusOK, "OK", hawkServer, cred, &hawkOpt)
	}
}
