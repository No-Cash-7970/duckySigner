package session

import (
	"crypto/ecdh"
)

// Confirmation contains data about a dApp connect session "confirmation". A
// "confirmation" is created when a dApp initializes a session and is no longer
// needed when the session is confirmed.
// NOTE: None of the struct fields are exported to make them read-only after the
// struct is created
type Confirmation struct {
	// Confirmation secret key
	key *ecdh.PrivateKey
}

// New creates a new Confirmation using the given confirmation secret key
func NewConfirmation(key *ecdh.PrivateKey) Confirmation {
	return Confirmation{
		key: key,
	}
}

// ID returns the confirmation ID
func (confirm *Confirmation) ID() *ecdh.PublicKey {
	return confirm.key.PublicKey()
}

// Key returns the confirmation secret key
func (confirm *Confirmation) Key() *ecdh.PrivateKey {
	return confirm.key
}

// SharedKey returns the confirmation shared secret key that is derived from
// confirmation secret key and the given dApp ID
func (confirm *Confirmation) SharedKey(dappId *ecdh.PublicKey) ([]byte, error) {
	return confirm.key.ECDH(dappId)
}
