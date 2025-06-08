package mocks

import (
	"crypto/ecdh"
	"encoding/base64"
	"io"
)

// TODO: Rename to X25519CurveMock
// Mock of the `ecdh.Curve` for X25519 that always "generates" the given predetermined
// private key. This allows for tests to be more predictable and consistent.
type EcdhCurveMock struct {
	// Base64 encoded private key that will always be "generated" by the
	// `GenerateKey()` method
	// TODO: Rename to PredeterminedKey
	GeneratedPrivateKey string
}

// GenerateKey will always return the predetermined key in the form of a
// `ecdh.PrivateKey`
func (c *EcdhCurveMock) GenerateKey(r io.Reader) (*ecdh.PrivateKey, error) {
	key, _ := base64.StdEncoding.DecodeString(c.GeneratedPrivateKey)
	return ecdh.X25519().NewPrivateKey(key)
}

func (c *EcdhCurveMock) NewPrivateKey(key []byte) (*ecdh.PrivateKey, error) {
	return ecdh.X25519().NewPrivateKey(key)
}

func (c *EcdhCurveMock) NewPublicKey(key []byte) (*ecdh.PublicKey, error) {
	return ecdh.X25519().NewPublicKey(key)
}
