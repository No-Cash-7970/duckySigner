package middleware

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"

	"github.com/hiyosi/hawk"
	"github.com/labstack/echo/v4"

	dc "duckysigner/internal/dapp_connect"
	"duckysigner/internal/dapp_connect/session"
	"duckysigner/internal/wallet_session"
)

// NOTE:
// This "middleware" for Hawk authentication is not actually middleware. The
// parts of this "middleware" have to be placed in the route handlers because of
// the need to access the request and response bodies for calculating the
// payload hash. There are 2 parts to this "middleware": (1) the Hawk
// authentication check that should occur in the beginning of a route handler
// and (2) the creation of the Hawk response header that should occur towards
// the end of a route handler.

// HawkOptions is the set of options for the Hawk authentication "middleware"
type HawkOptions struct {
	// The Echo context used for request
	EchoContext echo.Context
	// Instance of Echo
	EchoInstance *echo.Echo
	// If authentication is optional
	OptionalAuth bool
	// The store that contains a function for retrieving credentials.
	CredentialStore hawk.CredentialStore
}

// HawkAuth is the part of the Hawk "middleware" that handles the request. It
// processes the request's `Authentication` header, if present, using the given
// request body contents and the given Hawk options. It returns a Hawk server
// instance and the retrieved credentials if there is no error. The Hawk server
// instance and retrieved credentials are typically used for creating a Hawk
// header for server responses.
//
// The given request body is used to calculate the payload hash. If nil, the
// payload hash will not be calculated. If it is an empty slice, an error will
// be returned.
func HawkAuth(reqBody []byte, opt *HawkOptions) (*hawk.Server, *hawk.Credential, *dc.ApiError) {
	var req = opt.EchoContext.Request()

	// Maybe require authentication
	if req.Header.Get("Authorization") == "" && opt.OptionalAuth {
		opt.EchoInstance.Logger.Debug(
			"No Authorization header. Skipping Hawk authentication because it is not required.",
		)
		return nil, nil, nil
	}

	opt.EchoInstance.Logger.Debug("Authenticating Hawk request header")

	hawkServer := hawk.NewServer(opt.CredentialStore)

	// Require payload hash only if payload is not nil
	if reqBody != nil {
		// Setting "Payload" makes the payload hash required
		hawkServer.Payload = string(reqBody)
	}

	// Authenticate the Hawk request header
	cred, err := hawkServer.Authenticate(req)
	if err != nil {
		return nil, nil, &dc.ApiError{Name: "auth_request_failed", Message: err.Error()}
	}

	return hawkServer, cred, nil
}

// HawkRespJSON is the part of the Hawk "middleware" that handles the response.
// It is a wrapper to the echo.context.JSON() function that sets the Hawk
// response header using the given Hawk server and Hawk options if the given
// Hawk credentials is not nil. The Hawk credential is usually nil when Hawk
// authentication is optional and the request was not authenticated.
func HawkRespJSON(
	respStatusCode int,
	respData any,
	hawkServer *hawk.Server,
	hawkCred *hawk.Credential,
	hawkOpt *HawkOptions,
) error {
	var c = hawkOpt.EchoContext
	var req = c.Request()

	// Build Hawk response header only when the request was successfully
	// authenticated (when credentials are not nil)
	if hawkCred != nil {
		hawkOpt.EchoInstance.Logger.Debug("Creating Hawk response header")

		// Convert response body data into JSON to be used for calculating
		// payload hash
		respBytes, err := json.Marshal(respData)
		if err != nil {
			apiErr := dc.ApiError{Name: "auth_response_failed", Message: err.Error()}
			return c.JSON(http.StatusInternalServerError, apiErr)
		}

		// Add Hawk header to response
		hawkRespHeader, err := hawkServer.Header(req, hawkCred, &hawk.Option{
			TimeStamp:   time.Now().Unix(),
			Payload:     string(respBytes),
			ContentType: "application/json",
		})
		if err != nil {
			apiErr := dc.ApiError{Name: "auth_response_failed", Message: err.Error()}
			return c.JSON(http.StatusInternalServerError, apiErr)
		}

		c.Response().Header().Set("Server-Authorization", hawkRespHeader)
	}

	return c.JSON(respStatusCode, respData)
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
