package dappconnect

import (
	"crypto/ecdh"
	"crypto/rand"
	"encoding/base64"
	"errors"

	"github.com/labstack/echo/v4"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// TODO: Document this
const DefaultServerAddr string = ":1323"

// TODO: Document this
const WCSessionInitUIPromptEventName string = "session_init_prompt"
const WCSessionInitUIRespEventName string = "session_init_response"

// TODO: document this
func CreateWCSessionKeyPair(curve ECDHCurve) (id *ecdh.PublicKey, sk *ecdh.PrivateKey, err error) {
	sk, err = curve.GenerateKey(rand.Reader)
	if err != nil {
		return
	}

	id = sk.PublicKey()
	return
}

// TODO: Document this
func ValidateDAppID(dAppID string, curve ECDHCurve) (dAppIdPk *ecdh.PublicKey, apiErr ApiError, err error) {
	// Check if given DApp ID is valid Base64 by attempting to decode it
	dappIdBytes, err := base64.StdEncoding.DecodeString(dAppID)
	if err != nil {
		apiErr = ApiError{
			Name:    "invalid_dapp_id_b64",
			Message: "DApp ID is not a valid Base64 string",
		}
		return
	}

	// Check if given dApp ID is a valid ECDH public key by attempting to
	// convert it from a byte slice to a PublicKey
	dAppIdPk, err = curve.NewPublicKey(dappIdBytes)
	// dappId, err := curve.NewPublicKey(dappIdBytes)
	if err != nil {
		apiErr = ApiError{
			Name:    "invalid_dapp_id_pk",
			Message: "DApp ID is invalid",
		}
		return
	}

	return
}

// TODO: Document this
// TODO: Note that event listener should be removed after data in channel is received and read
func PromptUI(
	dappInfo *DAppInfo,
	promptEvent string,
	respEvent string,
	wailsApp *application.App,
	logger echo.Logger,
) (uiResp chan []string, err error) {
	// Check if Wails app is properly initialized
	if wailsApp == nil {
		err = errors.New("Missing Wails app instance. The dApp connection service was improperly initialized.")
		return
	}

	// Contains the UI's response to prompt
	uiResp = make(chan []string)

	// Set up listener for event that will contain UI's response
	logger.Debug("Listening for", respEvent, "event from UI")
	wailsApp.OnEvent(respEvent, func(e *application.CustomEvent) {
		logger.Debug("Event from UI:", respEvent, "\nEvent data:", e.Data)
		// NOTE: For some reason, the actual event data is always within a slice
		uiResp <- e.Data.([]interface{})[0].([]string)
		// Only need to know about the first instance of this response, so
		// immediately close the channel once we get a response. The UI should
		// only be able to respond to emitted prompt event once. Closing this
		// channel does not remove the event listener. That will need to be done
		// somewhere else after the channel value is received and read.
		close(uiResp)
	})

	// Prompt the UI after setting up listener for the expected UI response
	// event because the UI may respond before being prompted, which is not
	// ideal when it happens in a unit test.
	logger.Debug("Emitted", promptEvent, "event to UI")
	wailsApp.EmitEvent(promptEvent, dappInfo)

	return
}

// TODO: Document this
func StoreWCSessionData(
	sessionId *ecdh.PublicKey,
	sessionKey *ecdh.PrivateKey,
	dappId *ecdh.PublicKey,
	logger echo.Logger,
) error {
	// Derive wallet connection shared key using session key and dApp ID
	wcKey, err := sessionKey.ECDH(dappId)
	if err != nil {
		return err
	}

	wcSession := WalletConnectionSession{
		DAppID:              dappId,
		SessionID:           sessionId,
		ServerKey:           sessionKey,
		WalletConnectionKey: wcKey,
	}

	// TODO: Remove the logs below. It is only here to avoid Go's unused variable error
	logger.Debug("Created wallet connection session for dApp with ID:", wcSession.DAppID)

	// TODO: Store connection session data into an encrypted db file

	return nil
}
