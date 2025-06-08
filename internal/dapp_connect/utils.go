package dappconnect

import (
	"crypto/ecdh"
	"crypto/rand"
	"encoding/base64"
	"errors"

	"github.com/labstack/echo/v4"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// The default address for the dApp connection server
// Default: localhost:1323
const DefaultServerAddr string = ":1323"

// The name for the event for triggering the UI to prompt the user to approve
// the dApp connection session initialization request
// TODO: Rename to DCSessionInitUIPromptEventName
const WCSessionInitUIPromptEventName string = "session_init_prompt"

// The name for the event that the UI uses to forward the user's response to the
// dApp connection session initialization request
// TODO: Rename to DCSessionInitUIRespEventName
const WCSessionInitUIRespEventName string = "session_init_response"

// CreateWCSessionKeyPair generates an Elliptic-curve Diffie–Hellman (ECDH) key
// pair that is to be used for a connection session with a dApp. Returns the
// generated key pair if successful.
// TODO: Rename to GenerateDCSessionKeyPair
func CreateWCSessionKeyPair(curve ECDHCurve) (id *ecdh.PublicKey, sk *ecdh.PrivateKey, err error) {
	sk, err = curve.GenerateKey(rand.Reader)
	if err != nil {
		return
	}

	id = sk.PublicKey()
	return
}

// ValidateDAppID validates the given Base64-encoded dApp ID according to the
// given curve (that is to be used for ECDH (Elliptic-curve Diffie–Hellman)).
// Returns the dAppId as a ECDH public key if successful. Returns an error and
// an API error message if unsuccessful.
// TODO: Rename
func ValidateDAppID(dAppID string, curve ECDHCurve) (dAppIdPk *ecdh.PublicKey, apiErr ApiError, err error) {
	// Check if given dApp ID is a valid Base64 string by attempting to decode it
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
	// TODO: Rename
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

// PromptUI sends the given data to the UI by emitting an event with given
// "prompt event" name to the UI once. It listens and waits only for the first
// response event from the UI with the given "response event" name. Subsequent
// responses are ignored. Returns a "UI response" channel that will contain data
// sent by the UI when the UI responds.
//
// NOTE: The event listener for the "response event" should be closed after
// reading the returned UI response channel or timing out waiting for data to
// come through the channel.
//
// Example:
//
//	func f() (string, error) {
//	    uiRespCh, err := PromptUI(data, "prompt_ui_evt", "ui_resp_evt", app, logger)
//	    defer wailsApp.OffEvent("ui_resp_evt")
//	    // Check and possibly return error AFTER setting up to remove the
//	    // listener when the function terminates
//	    if err != nil {
//	        return "", err
//	    }
//	    // Wait for UI response...
//		select {
//		case <-time.After(5 * time.Minute)
//	        return "Time ran out waiting for UI to respond", nil
//	    case uiResp := <-uiRespCh: // Got response from UI
//	        return uiResp, nil
//	    }
//	}
//
// TODO: Rename to PromptUIOnce
func PromptUI(
	dappInfo *DAppInfo, // TODO: Change to `data any`
	promptEvent string,
	respEvent string,
	wailsApp *application.App,
	logger echo.Logger,
	// TODO: rename to uiRespCh
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

	// Prompt the UI *after* setting up the listener for the UI response event
	// because the UI could quickly respond to the emitted prompt event before
	// the server has the chance to set up the listener to receive the UI's
	// response. The server misses the UI's only response, which can be a
	// problem when unit testing.
	logger.Debug("Emitted", promptEvent, "event to UI")
	wailsApp.EmitEvent(promptEvent, dappInfo)

	return
}

// StoreWCSessionData stores the dApp connection data to a database file
// TODO: Rename to StoreDCSessionData
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

	// TODO: Store connection session data into an encrypted (or password protected?) db file
	// TODO: Also store DApp info into the db file

	return nil
}
