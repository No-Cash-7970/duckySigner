package handlers

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"

	"github.com/algorand/go-algorand-sdk/v2/encoding/msgpack"
	algoTypes "github.com/algorand/go-algorand-sdk/v2/types"
	"github.com/labstack/echo/v4"
	"github.com/wailsapp/wails/v3/pkg/application"

	dc "duckysigner/internal/dapp_connect"
	mw "duckysigner/internal/dapp_connect/middleware"
	"duckysigner/internal/dapp_connect/session"
	"duckysigner/internal/wallet_session"
)

type (
	// TransactionSignPostReq is the request data for `POST /transaction/sign`
	TransactionSignPostReq struct {
		// Unsigned transaction
		Txn string `json:"transaction" validate:"required"`
		// Address of the signer
		Signer string `json:"signer,omitempty"`
	}

	// TransactionSignPostResp is the response data to a
	// `POST /transaction/sign` request
	TransactionSignPostResp struct {
		// Signed transaction
		SignedTxn string `json:"signed_transaction"`
	}

	// TxnSignPromptEvtData is the data passed to the UI when prompting the user to
	// sign the transaction
	TxnSignPromptEvtData struct {
		TxnData TransactionSignPostReq `json:"data"`
	}

	// TxnSignRespEvtData is the data the UI sends when the user responds to the
	// prompt to sign the transaction
	TxnSignRespEvtData struct {
		// If the transaction has been approved
		Approved bool `json:"approved"`
	}
)

// TxnSignPromptEventName is the name for the event for triggering the
// UI to prompt the user to approve the signing a transaction
const TxnSignPromptEventName string = "txn_sign_prompt"

// TxnSignRespEventName is the name for the event that the UI uses to
// forward the user's response to the transaction signing request
const TxnSignRespEventName string = "txn_sign_response"

func TransactionSignPost(
	echoInstance *echo.Echo,
	wailsApp *application.App,
	walletSession *wallet_session.WalletSession,
	sessionManager *session.Manager,
	ecdhCurve dc.ECDHCurve,
) echo.HandlerFunc {
	return func(c echo.Context) error {
		if walletSession == nil {
			apiErr := dc.ApiError{
				Name:    "no_wallet_session",
				Message: "There is currently no valid wallet session. Log in to a wallet and try again.",
			}
			return c.JSON(http.StatusInternalServerError, apiErr)
		}

		// Read request data
		rawReqBody, err := dc.GetRawRequestBody(c.Request())
		if err != nil {
			apiErr := dc.ApiError{Name: "bad_request", Message: err.Error()}
			return c.JSON(http.StatusBadRequest, apiErr)
		}
		// Parse request data
		reqData := new(TransactionSignPostReq)
		if err := json.Unmarshal(rawReqBody, &reqData); err != nil {
			apiErr := dc.ApiError{Name: "bad_request", Message: err.Error()}
			return c.JSON(http.StatusBadRequest, apiErr)
		}

		// Hawk authentication
		hawkOpt := mw.HawkOptions{
			EchoContext:  c,
			EchoInstance: echoInstance,
			CredentialStore: mw.SessionCredentialStore{
				WalletSession:  walletSession,
				SessionManager: sessionManager,
			},
		}
		hawkServer, cred, apiErr := mw.HawkAuth(rawReqBody, &hawkOpt)
		if apiErr != nil {
			// Set WWW-Authenticate header
			c.Response().Header().Set("WWW-Authenticate", "Hawk")
			// Respond with 401 Unauthorized
			return c.JSON(http.StatusUnauthorized, apiErr)
		}

		// Validate request data
		if err := c.Validate(reqData); err != nil {
			apiErr := dc.ApiError{Name: "validation_error", Message: err.Error()}
			return mw.HawkRespJSON(http.StatusBadRequest, apiErr, hawkServer, cred, &hawkOpt)
		}

		// Decode and check transaction data
		var unsignedTxn algoTypes.Transaction
		unsignedTxnBytes, err := base64.StdEncoding.DecodeString(reqData.Txn)
		if err != nil {
			apiErr := dc.ApiError{Name: "invalid_txn", Message: err.Error()}
			return mw.HawkRespJSON(http.StatusBadRequest, apiErr, hawkServer, cred, &hawkOpt)
		}
		err = msgpack.Decode(unsignedTxnBytes, &unsignedTxn)
		if err != nil {
			apiErr := dc.ApiError{Name: "invalid_txn", Message: err.Error()}
			return mw.HawkRespJSON(http.StatusBadRequest, apiErr, hawkServer, cred, &hawkOpt)
		}

		// Check if sender exists in wallet
		senderInWallet, err := walletSession.CheckAddrInWallet(unsignedTxn.Sender.String())
		if err != nil {
			echoInstance.Logger.Error(err)
			apiErr := dc.ApiError{Name: "invalid_sender", Message: "Failed to parse sender address"}
			return mw.HawkRespJSON(http.StatusBadRequest, apiErr, hawkServer, cred, &hawkOpt)
		}
		if !senderInWallet {
			apiErr := dc.ApiError{Name: "invalid_sender", Message: "Sender account not in wallet"}
			return mw.HawkRespJSON(http.StatusBadRequest, apiErr, hawkServer, cred, &hawkOpt)
		}

		// Check if signer exists in wallet
		if reqData.Signer != "" {
			signerInWallet, err := walletSession.CheckAddrInWallet(reqData.Signer)
			if err != nil {
				echoInstance.Logger.Error(err)
				apiErr := dc.ApiError{
					Name:    "invalid_signer",
					Message: "Failed to parse signer address",
				}
				return mw.HawkRespJSON(http.StatusBadRequest, apiErr, hawkServer, cred, &hawkOpt)
			}
			if !signerInWallet {
				apiErr := dc.ApiError{Name: "invalid_signer", Message: "Signer account not in wallet"}
				return mw.HawkRespJSON(http.StatusBadRequest, apiErr, hawkServer, cred, &hawkOpt)
			}
		}

		// Prompt user to approve transaction
		promptDataJSON, err := json.Marshal(TxnSignPromptEvtData{TxnData: *reqData})
		if err != nil {
			echoInstance.Logger.Error(err)
			apiErr := dc.ApiError{
				Name:    "prompt_user_fail",
				Message: "Failed to prompt the user",
			}
			return mw.HawkRespJSON(http.StatusInternalServerError, apiErr, hawkServer, cred, &hawkOpt)
		}
		userResp, err := dc.PromptUIOnce(
			string(promptDataJSON),
			TxnSignPromptEventName,
			TxnSignRespEventName,
			wailsApp,
			echoInstance.Logger,
		)
		// Remove listener for UI response event when the server request ends,
		// which is definitely after the UI response event data is received from
		// the channel
		defer wailsApp.Event.Off(SessionConfirmRespEventName)
		if err != nil {
			echoInstance.Logger.Error(err)
			apiErr := dc.ApiError{Name: "unexpected_fail", Message: err.Error()}
			return mw.HawkRespJSON(http.StatusInternalServerError, apiErr, hawkServer, cred, &hawkOpt)
		}

		// Wait for user response...
		select {
		case <-time.After(sessionManager.ApprovalTimeout()): // Time ran out
			echoInstance.Logger.Info("Ran out of time waiting for user response")
			apiErr := dc.ApiError{Name: "txn_sign_timeout", Message: "User did not respond"}
			return mw.HawkRespJSON(http.StatusRequestTimeout, apiErr, hawkServer, cred, &hawkOpt)
		case dataJSON := <-userResp: // Got user's response
			echoInstance.Logger.Debug("Received transaction approval user response:", dataJSON)

			var userRespData []TxnSignRespEvtData

			err := json.Unmarshal([]byte(dataJSON), &userRespData)
			if err != nil {
				echoInstance.Logger.Error(err)
				apiErr := dc.ApiError{
					Name:    "user_response_fail",
					Message: "Failed to process user response",
				}
				return mw.HawkRespJSON(http.StatusInternalServerError, apiErr, hawkServer, cred, &hawkOpt)
			}

			// Respond with error if user rejects
			if !userRespData[0].Approved {
				apiErr := dc.ApiError{
					Name:    "txn_sign_rejected",
					Message: "User rejected the transaction",
				}
				return mw.HawkRespJSON(http.StatusForbidden, apiErr, hawkServer, cred, &hawkOpt)
			}

			stxn, err := walletSession.SignTransaction(reqData.Txn, reqData.Signer)
			if err != nil {
				echoInstance.Logger.Error(err)
				apiErr := dc.ApiError{
					Name:    "txn_sign_fail",
					Message: "Failed to sign transaction",
				}
				return mw.HawkRespJSON(http.StatusInternalServerError, apiErr, hawkServer, cred, &hawkOpt)
			}

			resp := TransactionSignPostResp{SignedTxn: stxn}

			return mw.HawkRespJSON(http.StatusOK, resp, hawkServer, cred, &hawkOpt)
		}
	}
}
