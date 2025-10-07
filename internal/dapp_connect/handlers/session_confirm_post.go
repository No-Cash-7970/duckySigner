package handlers

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"

	"github.com/hiyosi/hawk"
	"github.com/labstack/echo/v4"
	"github.com/wailsapp/wails/v3/pkg/application"

	dc "duckysigner/internal/dapp_connect"
	"duckysigner/internal/dapp_connect/session"
	"duckysigner/internal/wallet_session"
)

type (
	// SessionConfirmPostReq is the request data for `POST /session/confirm`
	SessionConfirmPostReq struct {
		// Confirmation token
		Token string `json:"token" validate:"required"`
		// DApp data
		DappData dc.DappData `json:"dapp"`
	}

	// SessionConfirmPostResp is the response data to a `POST /session/confirm`
	// request
	SessionConfirmPostResp struct {
		// Session ID
		Id string `json:"id"`
		// Session expiration date-time in Unix Epoch
		Expiration int64 `json:"exp"`
		// The addresses that are allowed to sign things in the session
		Addresses []string `json:"addrs"`
	}

	// UserPromptData is the data passed to the UI when prompting the user to
	// approve the session
	UserPromptData struct {
		DappData dc.DappData `json:"dapp"`
	}

	// UserRespData is the data the UI sends when the user responds to the
	// prompt
	UserRespData struct {
		Code      string   `json:"code"`
		Addresses []string `json:"addrs"`
	}

	// ConfirmCredentialStoreConfig is the configuration for a
	// ConfirmCredentialStore. The primary purpose of this is to contain
	// pointers that can be set within the credential store function.
	ConfirmCredentialStoreConfig struct {
		// The current wallet session
		WalletSession *wallet_session.WalletSession
		// An instance of the session manager
		SessionManager *session.Manager
		// The confirmation extracted from the decrypted confirmation token.
		// Used to avoid needing to decrypt the token twice. This is set by the
		// credential function, so it serves as a hackish way for the credential
		// function to return something other than Hawk credentials.
		ExtractedConfirm *session.Confirmation
		// The ECDH curve used for encryption
		ECDHCurve dc.ECDHCurve
	}

	// ConfirmCredentialStore is a Hawk CredentialStore use for retrieving
	// dApp connect session confirmation credentials
	ConfirmCredentialStore struct {
		confirmToken string
		config       *ConfirmCredentialStoreConfig
	}
)

// UIPromptEventName is the name for the event for triggering the
// UI to prompt the user to approve the dApp connect session confirmation
// request
const SessionConfirmPromptEventName string = "session_confirm_prompt"

// UIRespEventName is the name for the event that the UI uses to
// forward the user's response to the dApp connect session confirmation request
const SessionConfirmRespEventName string = "session_confirm_response"

// SessionInitPost is the route handler for `POST /session/init`
func SessionConfirmPost(
	echoInstance *echo.Echo,
	wailsApp *application.App,
	walletSession *wallet_session.WalletSession,
	sessionManager *session.Manager,
	ecdhCurve dc.ECDHCurve,
) echo.HandlerFunc {
	return func(c echo.Context) error {
		if walletSession == nil {
			return c.JSON(http.StatusInternalServerError, dc.ApiError{
				Name:    "no_wallet_session",
				Message: "There is currently no valid wallet session. Log in to a wallet and try again.",
			})
		}

		// Read request data
		reqData := new(SessionConfirmPostReq)
		if err := c.Bind(reqData); err != nil {
			return c.JSON(http.StatusBadRequest, dc.ApiError{
				Name:    "bad_request",
				Message: err.Error(),
			})
		}

		// Validate request data
		if err := c.Validate(reqData); err != nil {
			return c.JSON(http.StatusBadRequest, dc.ApiError{
				Name:    "validation_error",
				Message: err.Error(),
			})
		}

		// Set up Hawk server
		rawReqData := c.Request()
		credStoreConfig := ConfirmCredentialStoreConfig{
			WalletSession:  walletSession,
			SessionManager: sessionManager,
			ECDHCurve:      ecdhCurve,
		}
		credStore := ConfirmCredentialStore{
			confirmToken: reqData.Token,
			config:       &credStoreConfig,
		}
		hawkServer := hawk.NewServer(credStore)

		// Authenticate the Hawk request header
		cred, err := hawkServer.Authenticate(rawReqData)
		if err != nil {
			// Set WWW-Authenticate header
			c.Response().Header().Set("WWW-Authenticate", "Hawk")
			// Respond with 401 Unauthorized
			return c.JSON(http.StatusUnauthorized, dc.ApiError{
				Name:    "confirm_auth_request_failed",
				Message: err.Error(),
			})
		}

		// Check if confirmation has expired
		if credStoreConfig.ExtractedConfirm.Expiration().Before(time.Now()) {
			return c.JSON(http.StatusUnauthorized, dc.ApiError{
				Name:    "confirm_expired",
				Message: "The confirmation has expired. Initialize the session again.",
			})
		}

		dappId := credStoreConfig.ExtractedConfirm.DappId()

		// Prompt user to approve dApp connect session
		promptDataJSON, err := json.Marshal(UserPromptData{DappData: reqData.DappData})
		if err != nil {
			echoInstance.Logger.Error(err)
			return c.JSON(http.StatusInternalServerError, dc.ApiError{
				Name:    "prompt_user_fail",
				Message: "Failed to prompt the user",
			})
		}
		userResp, err := dc.PromptUIOnce(
			string(promptDataJSON),
			SessionConfirmPromptEventName,
			SessionConfirmRespEventName,
			wailsApp,
			echoInstance.Logger,
		)
		// Remove listener for UI response event when the server request ends,
		// which is definitely after the UI response event data is received from
		// the channel
		defer wailsApp.Event.Off(SessionConfirmRespEventName)
		if err != nil {
			echoInstance.Logger.Error(err)
			return c.JSON(http.StatusBadRequest, err)
		}

		// Wait for user response...
		select {
		case <-time.After(sessionManager.ApprovalTimeout()): // Time ran out
			echoInstance.Logger.Info("Ran out of time waiting for user response")
			return c.JSON(
				http.StatusRequestTimeout,
				dc.ApiError{Name: "confirm_timeout", Message: "User did not respond"},
			)
		case dataJSON := <-userResp: // Got user's response
			var userRespData []UserRespData

			err := json.Unmarshal([]byte(dataJSON), &userRespData)
			if err != nil {
				echoInstance.Logger.Error(err)
				return c.JSON(http.StatusInternalServerError, dc.ApiError{
					Name:    "user_response_fail",
					Message: "Failed to process user response",
				})
			}

			// If no confirmation code is given, that means the user rejected
			if userRespData[0].Code == "" {
				return c.JSON(
					http.StatusForbidden,
					dc.ApiError{Name: "session_rejected", Message: "Session was rejected"},
				)
			}

			// Check if the code is correct
			if userRespData[0].Code != credStoreConfig.ExtractedConfirm.Code() {
				return c.JSON(
					http.StatusForbidden,
					dc.ApiError{
						Name:    "wrong_confirm_code",
						Message: "The user did not enter the correct confirmation code",
					},
				)
			}

			// Generate session
			session, err := sessionManager.GenerateSession(dappId, &reqData.DappData, nil)
			if err != nil {
				echoInstance.Logger.Error(err)
				return c.JSON(http.StatusInternalServerError, dc.ApiError{
					Name:    "session_create_fail",
					Message: "Failed to create dApp connect session",
				})
			}

			// Get encryption key needed to modify the session database
			mek, err := walletSession.GetMasterKey()
			if err != nil {
				echoInstance.Logger.Error(err)
				return c.JSON(http.StatusInternalServerError, dc.ApiError{
					Name:    "session_create_fail",
					Message: "Failed to create dApp connect session",
				})
			}

			// Store generated session
			err = sessionManager.StoreSession(session, mek)
			if err != nil {
				echoInstance.Logger.Error(err)
				return c.JSON(http.StatusInternalServerError, dc.ApiError{
					Name:    "session_create_fail",
					Message: "Failed to create dApp connect session",
				})
			}

			// Prepare response
			resp := SessionConfirmPostResp{
				Id:         base64.StdEncoding.EncodeToString(session.ID().Bytes()),
				Expiration: session.Expiration().Unix(),
				Addresses:  userRespData[0].Addresses,
			}
			respJSON, err := json.Marshal(resp)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, dc.ApiError{
					Name:    "confirm_auth_response_failed",
					Message: err.Error(),
				})
			}

			// Add Hawk header to response
			hawkRespHeader, err := hawkServer.Header(rawReqData, cred, &hawk.Option{
				TimeStamp:   time.Now().Unix(),
				Payload:     string(respJSON),
				ContentType: "application/json",
			})
			if err != nil {
				return c.JSON(http.StatusInternalServerError, dc.ApiError{
					Name:    "confirm_auth_response_failed",
					Message: err.Error(),
				})
			}
			c.Response().Header().Set("Server-Authorization", hawkRespHeader)

			return c.JSON(http.StatusOK, resp)
		}
	}
}

func (store ConfirmCredentialStore) GetCredential(id string) (*hawk.Credential, error) {
	mek, err := store.config.WalletSession.GetMasterKey()
	if err != nil {
		return nil, err
	}

	key, err := store.config.SessionManager.GetConfirmKey(id, mek)
	if err != nil {
		return nil, err
	}

	confirm, err := session.DecryptToken(store.confirmToken, key, store.config.ECDHCurve)
	if err != nil {
		return nil, err
	}

	// Set the extracted confirmation to avoid needing to decrypt the token
	// twice
	store.config.ExtractedConfirm = confirm

	sharedKey, err := confirm.SharedKey()
	if err != nil {
		return nil, err
	}

	return &hawk.Credential{
		ID:  id,
		Key: base64.StdEncoding.EncodeToString(sharedKey),
		Alg: hawk.SHA256,
	}, nil
}
