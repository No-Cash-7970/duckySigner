package session

import (
	"crypto/ecdh"
	"time"

	dc "duckysigner/internal/dapp_connect"
)

// Session contains data about a dApp connect session
// NOTE: None of the struct fields are exported to make them read-only after the
// struct is created
type Session struct {
	// Session secret key
	key *ecdh.PrivateKey
	// DApp ID for the dApp the session is for
	dappId *ecdh.PublicKey
	// Data about the dApp the session is for
	dappData *dc.DappData
	// Session expiration date-time, if the session has been established
	exp time.Time
	// The date-time the session was established, if it has been established
	establishedAt time.Time
	// List of addresses the dApp is allowed to connect to. If empty or nil, the
	// dApp is allowed to connect to all addresses in the wallet.
	addrs []string
}

// New creates a new Session using the given session data
func New(
	key *ecdh.PrivateKey,
	dappId *ecdh.PublicKey,
	exp time.Time,
	establishedAt time.Time,
	dappData *dc.DappData,
	addrs []string,
) Session {
	return Session{
		key:           key,
		dappId:        dappId,
		exp:           exp,
		establishedAt: establishedAt,
		dappData:      dappData,
		addrs:         addrs,
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

// DappId returns the ID of the dApp the session is for
func (session *Session) DappId() *ecdh.PublicKey {
	return session.dappId
}

// Expiration returns the date-time when the session expires. Returns a 0
// date-time value if the expiration was not set.
func (session *Session) Expiration() time.Time {
	return session.exp
}

// EstablishedAt returns the date-time when the session was established, if it
// has been established. Returns a 0 date-time value if the establishment
// date-time was not set because the session has not been established yet. A
// session is established when the dApp has gone through both the session
// initialization and confirmation processes. If the session has not been
// established, a 0 date-time value is returned.
func (session *Session) EstablishedAt() time.Time {
	return session.establishedAt
}

// DappData returns the data of the dApp the session is for
func (session *Session) DappData() *dc.DappData {
	return session.dappData
}

// Addresses returns the list of addresses the dApp is allowed to connected to
func (session *Session) Addresses() []string {
	return session.addrs
}

// SharedKey returns the session shared secret key that is derived from session
// secret key and the dApp ID
func (session *Session) SharedKey() ([]byte, error) {
	return session.key.ECDH(session.dappId)
}
