package mocks

import (
	"crypto/ecdh"
	"encoding/base64"
	"io"
)

// TODO: Document this
type EcdhCurveMock struct {
	GeneratedPrivateKey string
}

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
