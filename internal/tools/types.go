package tools

import (
	"crypto/ecdh"
	"io"
)

// ECDHCurve is a ECDH curve interface based on the `ecdh.Curve` interface in
// the `crypto/ecdh` package
type ECDHCurve interface {
	GenerateKey(rand io.Reader) (*ecdh.PrivateKey, error)
	NewPrivateKey(key []byte) (*ecdh.PrivateKey, error)
	NewPublicKey(key []byte) (*ecdh.PublicKey, error)
}
