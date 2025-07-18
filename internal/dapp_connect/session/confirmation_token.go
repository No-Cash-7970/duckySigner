package session

import (
	"crypto/ecdh"
	"time"
)

// ConfirmationToken contains the data needed to create a confirmation token.
// The confirmation token is used to "confirm" an initialized dApp connect
// session.
// NOTE: None of the struct fields are exported to make them read-only after the
// struct is created
type ConfirmationToken struct {
	dappId      *ecdh.PublicKey
	sessionKey  *ecdh.PrivateKey
	confirmKey  *ecdh.PrivateKey
	confirmCode string
	confirmExp  time.Time
}

// NewConfirmationToken creates a new ConfirmationToken with the given dApp ID,
// session secret key, confirmation secret key, confirmation code and
// confirmation expiration date-time
func NewConfirmationToken(
	dappId *ecdh.PublicKey,
	sessionKey *ecdh.PrivateKey,
	confirmKey *ecdh.PrivateKey,
	confirmCode string,
	confirmExp time.Time,
) *ConfirmationToken {
	return &ConfirmationToken{
		dappId:      dappId,
		sessionKey:  sessionKey,
		confirmKey:  confirmKey,
		confirmCode: confirmCode,
		confirmExp:  confirmExp,
	}
}

// DappId returns the ID of the dApp the confirmation is for
func (token *ConfirmationToken) DappId() *ecdh.PublicKey {
	return token.dappId
}

// SessionKey returns the session key for the session to be confirmed. The
// session key is included within the confirmation token to prevent the token
// from being used to confirm an initialized session more than once.
func (token *ConfirmationToken) SessionKey() *ecdh.PrivateKey {
	return token.sessionKey
}

// ConfirmationId returns the confirmation ID that uniquely identifies an
// unconfirmed initialized session.
func (token *ConfirmationToken) ConfirmationId() *ecdh.PublicKey {
	return token.confirmKey.PublicKey()
}

// ConfirmationKey returns the confirmation key that will be used to create the
// confirmation token. It is not included within confirmation token.
func (token *ConfirmationToken) ConfirmationKey() *ecdh.PrivateKey {
	return token.confirmKey
}

// Code returns the confirmation code, which is used when confirming an
// initialized session.
func (token *ConfirmationToken) Code() string {
	return token.confirmCode
}

// Expiration returns the expiration date-time of when the initialized
// session can be confirmed
func (token *ConfirmationToken) Expiration() time.Time {
	return token.confirmExp
}

// ToTokenString uses the confirmation token data to create an encrypted
// confirmation token string. This string is given to the dApp for it to confirm
// its initialized session.
func (token *ConfirmationToken) ToTokenString() (string, error) {
	// TODO: Complete this
	return "", nil
}

// DecryptConfirmationToken attempts to decrypt the given confirmation token
// using the given confirmation key and extract the contents of the token into
// a ConfirmationToken
func DecryptConfirmationToken(confirmToken string, confirmKey *ecdh.PrivateKey) (*ConfirmationToken, error) {
	// TODO: Complete this
	return nil, nil
}
