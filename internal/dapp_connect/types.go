package dappconnect

import (
	"crypto/ecdh"
	"io"
)

// ECDH curve interface based on the `ecdh.Curve` interface in the `crypto/ecdh`
// package
type ECDHCurve interface {
	GenerateKey(rand io.Reader) (*ecdh.PrivateKey, error)
	NewPrivateKey(key []byte) (*ecdh.PrivateKey, error)
	NewPublicKey(key []byte) (*ecdh.PublicKey, error)
}

// A set of credentials for authenticating using Hawk
type HawkCredentials struct {
	// Authentication ID
	ID string `json:"id"`
	// Authentication key
	Key string `json:"key"`
	// The hash algorithm used to create the message authentication code (MAC).
	Algorithm string `json:"algorithm"`
}

// An error message
type ApiError struct {
	// Error name
	Name string `json:"name,omitempty"`
	// Error message
	Message string `json:"message,omitempty"`
}

// DApp information
type DAppInfo struct {
	// Name of the application connecting to wallet
	Name string `json:"name"`
	// URL for the app connecting to the wallet
	Url string `json:"url,omitempty"`
	// Description of the app
	Description string `json:"description,omitempty"`
	// Icon for the app connecting to wallet as a Base64 encoded JPEG, PNG or SVG data URI
	Icon string `json:"icon,omitempty"`
	// TODO: Document this and change the name
	DAppID string `json:"dapp_session_pk"`
}

// TODO: Document this
type WalletConnectionSession struct {
	DAppID              *ecdh.PublicKey
	SessionID           *ecdh.PublicKey
	ServerKey           *ecdh.PrivateKey
	WalletConnectionKey []byte
}
