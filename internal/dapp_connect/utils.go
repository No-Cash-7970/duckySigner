package dapp_connect

import (
	"crypto/ecdh"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// DefaultServerAddr is the default address for the dApp connect server
// Default: localhost:1323
const DefaultServerAddr string = ":1323"

// CustomValidator is a custom validator to use with Echo to validate request
// data
type CustomValidator struct {
	validator *validator.Validate
}

// Validate validates the given struct
// Modified from: <https://echo.labstack.com/docs/request#validate-data>
func (cv *CustomValidator) Validate(i any) error {
	if err := cv.validator.Struct(i); err != nil {
		return err
	}

	return nil
}

// SetupCustomValidator sets up the custom validator for the given Echo instance
func SetupCustomValidator(e *echo.Echo) {
	e.Validator = &CustomValidator{
		validator: validator.New(validator.WithRequiredStructEnabled()),
	}
}

// ValidateDappID validates the given Base64-encoded dApp ID according to the
// given curve (that is to be used for ECDH (Elliptic-curve Diffie–Hellman)).
// Returns the dAppId as a ECDH public key if successful. Returns an error and
// an API error message if unsuccessful.
func ValidateDappID(dAppID string, curve ECDHCurve) (dappIdPk *ecdh.PublicKey, apiErr ApiError, err error) {
	// Check if given dApp ID is a valid Base64 string by attempting to decode it
	dappIdBytes, err := base64.StdEncoding.DecodeString(dAppID)
	if err != nil {
		apiErr = ApiError{
			Name:    "bad_dapp_id",
			Message: "DApp ID is not a valid Base64 string",
		}
		return
	}

	// Check if given dApp ID is a valid ECDH public key by attempting to
	// convert it from a byte slice to a PublicKey
	dappIdPk, err = curve.NewPublicKey(dappIdBytes)
	// dappId, err := curve.NewPublicKey(dappIdBytes)
	if err != nil {
		apiErr = ApiError{
			Name:    "bad_dapp_id",
			Message: "DApp ID is not a valid public key",
		}
		return
	}

	return
}

// PromptUIOnce sends the given data to the UI by emitting an event with given
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
//	    uiRespCh, err := PromptUIOnce(data, "prompt_ui_evt", "ui_resp_evt", app, logger)
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
func PromptUIOnce(
	promptData string,
	promptEvent string,
	respEvent string,
	wailsApp *application.App,
	logger echo.Logger,
) (uiRespCh chan string, err error) {
	// Check if Wails app is properly initialized
	if wailsApp == nil {
		err = errors.New("missing Wails app instance, dApp connect service not properly initialized")
		return
	}

	// Contains the UI's response to prompt
	uiRespCh = make(chan string)

	logger.Debug("Removing any existing event listeners for: ", respEvent)
	wailsApp.Event.Off(respEvent)

	// Set up listener for event that will contain UI's response
	logger.Debug("Listening for ", respEvent, " event from UI")
	wailsApp.Event.On(respEvent, func(e *application.CustomEvent) {
		logger.Debug("Event from UI: ", respEvent, "\nEvent data:", e.Data)
		// NOTE: For some reason, the actual event data is always within an array
		uiRespCh <- fmt.Sprint(e.Data)
		// Only need to know about the first instance of this response, so
		// immediately close the channel once we get a response. The UI should
		// only be able to respond to emitted prompt event once. Closing this
		// channel does not remove the event listener. That will need to be done
		// somewhere else after the channel value is received and read.
		close(uiRespCh)
	})

	// Prompt the UI *after* setting up the listener for the UI response event
	// because the UI could quickly respond to the emitted prompt event before
	// the server has the chance to set up the listener to receive the UI's
	// response. The server misses the UI's only response, which can be a
	// problem when unit testing.
	logger.Debug("Emitted ", promptEvent, " event to UI")
	wailsApp.Event.Emit(promptEvent, promptData)

	return
}
