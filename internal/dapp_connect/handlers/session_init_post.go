package handlers

import (
	"encoding/base64"
	"net/http"

	"github.com/labstack/echo/v4"

	dc "duckysigner/internal/dapp_connect"
	"duckysigner/internal/dapp_connect/session"
	"duckysigner/internal/wallet_session"
)

type (
	// SessionInitPostReq is the request data for `POST /session/init`
	SessionInitPostReq struct {
		DappId string `json:"dapp_id" validate:"required,base64"`
	}

	// SessionInitPostResp is the response data to a `POST /session/init`
	// request
	SessionInitPostResp struct {
		// Confirmation ID
		Id string `json:"id"`
		// Confirmation code
		Code string `json:"code"`
		// Confirmation token
		Token string `json:"token"`
		// Confirmation expiration date-time in Unix Epoch
		Expiration int64 `json:"exp"`
	}
)

const (
	confirmCreateFailName = "session_confirm_create_fail"
	confirmCreateFailMsg  = "Failed to create a new session confirmation"
)

// SessionInitPost is the route handler for `POST /session/init`
func SessionInitPost(
	echoInstance *echo.Echo,
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
		reqData := new(SessionInitPostReq)
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

		// Ensure dApp ID is valid and convert it into an ECDH public key
		dappId, dappIdApiErr, err := dc.ValidateDappID(reqData.DappId, ecdhCurve)
		if err != nil {
			echoInstance.Logger.Error(err)
			return c.JSON(http.StatusBadRequest, dappIdApiErr)
		}

		// Create confirmation
		confirm, err := sessionManager.GenerateConfirmation(dappId)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, dc.ApiError{
				Name:    confirmCreateFailName,
				Message: confirmCreateFailMsg,
			})
		}

		// Retrieve the master encryption key of the currently opened wallet
		mek, err := walletSession.GetMasterKey()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, dc.ApiError{
				Name:    confirmCreateFailName,
				Message: confirmCreateFailMsg,
			})
		}

		// Store confirmation using master encryption key
		if err := sessionManager.StoreConfirmKey(confirm.Key(), mek); err != nil {
			return c.JSON(http.StatusInternalServerError, dc.ApiError{
				Name:    confirmCreateFailName,
				Message: confirmCreateFailMsg,
			})
		}

		// Generate confirmation token
		token, err := confirm.GenerateTokenString()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, dc.ApiError{
				Name:    confirmCreateFailName,
				Message: confirmCreateFailMsg,
			})
		}

		// Create and respond with session confirmation data
		return c.JSON(http.StatusOK, SessionInitPostResp{
			Id:         base64.StdEncoding.EncodeToString(confirm.ID().Bytes()),
			Code:       confirm.Code(),
			Token:      token,
			Expiration: confirm.Expiration().Unix(),
		})
	}
}
