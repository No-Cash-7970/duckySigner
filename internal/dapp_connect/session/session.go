package session

import (
	"crypto/ecdh"
	"time"

	dc "duckysigner/internal/dapp_connect"
)

// Session contains data about an established dApp connect session
// NOTE: None of the struct fields are exported to make them read-only after the
// struct is created
type Session struct {
	// Session secret key
	key *ecdh.PrivateKey
	// Session expiration date-time
	exp time.Time
	// DApp ID for the dApp the session is for
	dappId *ecdh.PublicKey
	// Data about the dApp the session is for
	dappData *dc.DappData
}

// New creates a new Session using the given session secret key, session
// expiration date-time, dApp ID and dApp data
func New(
	key *ecdh.PrivateKey,
	exp time.Time,
	dappId *ecdh.PublicKey,
	dappData *dc.DappData,
) Session {
	return Session{
		key:      key,
		exp:      exp,
		dappId:   dappId,
		dappData: dappData,
	}
}

// ID returns the session ID
func (session *Session) ID() *ecdh.PublicKey {
	return session.key.PublicKey()
}

// Key returns the session secret key
func (session *Session) Key() *ecdh.PrivateKey {
	return session.key
}

// Expiration returns the date-time when the session expires
func (session *Session) Expiration() time.Time {
	return session.exp
}

// DappId returns the ID of the dApp the session is for
func (session *Session) DappId() *ecdh.PublicKey {
	return session.dappId
}

// DappData returns the data of the dApp the session is for
func (session *Session) DappData() *dc.DappData {
	return session.dappData
}

// SharedKey returns the session shared secret key that is derived from session
// secret key and the dApp ID
func (session *Session) SharedKey() ([]byte, error) {
	return session.key.ECDH(session.dappId)
}
