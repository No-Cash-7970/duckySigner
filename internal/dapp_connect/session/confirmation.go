package session

import (
	"crypto/ecdh"
	"encoding/base64"
	"errors"
	"time"

	"aidanwoods.dev/go-paseto"

	dc "duckysigner/internal/dapp_connect"
)

const (
	// DappIdClaimName is the name for the "claim" that contains the dApp ID
	// within a PASETO for a confirmation token
	DappIdClaimName = "dapp"
	// SessionKeyClaimName is the name for the "claim" that contains the session
	// key within a PASETO for a confirmation token
	SessionKeyClaimName = "skey"
	// ConfirmCodeClaimName is the name for the "claim" that contains the
	// confirmation code within a PASETO for a confirmation token
	ConfirmCodeClaimName = "code"

	// MissingConfirmTokenDappIdErrMsg is the error message for when the dApp ID
	// is missing within the confirmation token
	MissingConfirmTokenDappIdErrMsg = "missing dApp ID in confirmation token"
	// MissingConfirmTokenSessionKeyErrMsg is the error message for when the
	// session key is missing within the confirmation token
	MissingConfirmTokenSessionKeyErrMsg = "missing session key in confirmation token"
	// MissingConfirmTokenConfirmKeyErrMsg is the error message for when the
	// confirmation key is missing within the confirmation token
	MissingConfirmTokenConfirmKeyErrMsg = "missing confirmation key in confirmation token"
	// MissingConfirmTokenCodeErrMsg is the error message for when the
	// confirmation code is missing within the confirmation token
	MissingConfirmTokenCodeErrMsg = "missing confirmation code in confirmation token"
)

// Confirmation contains the data about a session confirmation and the data
// needed to create a confirmation token. The confirmation token is used to
// "confirm" an initialized dApp connect session.
// NOTE: None of the struct fields are exported to make them read-only after the
// struct is created
type Confirmation struct {
	key        *ecdh.PrivateKey
	code       string
	exp        time.Time
	dappId     *ecdh.PublicKey
	sessionKey *ecdh.PrivateKey
	// TODO: issuedAt date?
}

// NewConfirmation creates a new Confirmation with the given dApp ID, session
// secret key, confirmation secret key, confirmation code and confirmation
// expiration date-time
func NewConfirmation(
	dappId *ecdh.PublicKey,
	sessionKey *ecdh.PrivateKey,
	confirmKey *ecdh.PrivateKey,
	confirmCode string,
	confirmExp time.Time,
) *Confirmation {
	return &Confirmation{
		dappId:     dappId,
		sessionKey: sessionKey,
		key:        confirmKey,
		code:       confirmCode,
		exp:        confirmExp,
	}
}

// ID returns the confirmation ID that uniquely identifies an unconfirmed
// initialized session
func (confirm *Confirmation) ID() *ecdh.PublicKey {
	return confirm.key.PublicKey()
}

// Key returns the confirmation key that will be used to create the confirmation
// token
func (confirm *Confirmation) Key() *ecdh.PrivateKey {
	return confirm.key
}

// Code returns the confirmation code, which is used when confirming an
// initialized session
func (confirm *Confirmation) Code() string {
	return confirm.code
}

// DappId returns the ID of the dApp the confirmation is for
func (confirm *Confirmation) DappId() *ecdh.PublicKey {
	return confirm.dappId
}

// SessionKey returns the session key for the session to be confirmed. The
// session key is included within the confirmation token to prevent the token
// from being used to confirm an initialized session more than once.
func (confirm *Confirmation) SessionKey() *ecdh.PrivateKey {
	return confirm.sessionKey
}

// Expiration returns the date-time the confirmation expires and cannot be used
func (confirm *Confirmation) Expiration() time.Time {
	return confirm.exp
}

// SharedKey returns the confirmation *shared* secret key that is derived from
// confirmation secret key and the given dApp ID
func (confirm *Confirmation) SharedKey() ([]byte, error) {
	return confirm.key.ECDH(confirm.dappId)
}

// GenerateTokenString uses the confirmation token data to create an encrypted
// confirmation token string. This string is given to the dApp for it to confirm
// its initialized session.
func (confirm *Confirmation) GenerateTokenString() (string, error) {
	if confirm.key == nil {
		return "", errors.New(MissingConfirmTokenConfirmKeyErrMsg)
	}

	if confirm.dappId == nil {
		return "", errors.New(MissingConfirmTokenDappIdErrMsg)
	}

	if confirm.code == "" {
		return "", errors.New(MissingConfirmTokenCodeErrMsg)
	}

	if confirm.sessionKey == nil {
		return "", errors.New(MissingConfirmTokenSessionKeyErrMsg)
	}

	// NOTE: Refer to the following links for explanations of "claims", a term
	// that PASETO borrowed from JWT.
	// - <https://github.com/paseto-standard/paseto-spec/blob/master/docs/02-Implementation-Guide/04-Claims.md>
	// - <https://medium.com/@dmosyan/json-web-token-claims-explained-e78a708ec43c>

	// Add "claims" to be included and encrypted in the PASETO
	confirmPaseto := paseto.NewToken()
	confirmPaseto.SetExpiration(confirm.exp)
	confirmPaseto.SetString(DappIdClaimName, base64.StdEncoding.EncodeToString(confirm.dappId.Bytes()))
	confirmPaseto.SetString(ConfirmCodeClaimName, confirm.code)
	confirmPaseto.Set(SessionKeyClaimName, base64.StdEncoding.EncodeToString(confirm.sessionKey.Bytes()))
	// Use confirmation key to encrypt the PASETO
	pasetoKey, err := paseto.V4SymmetricKeyFromBytes(confirm.key.Bytes())
	if err != nil {
		return "", err
	}

	return confirmPaseto.V4Encrypt(pasetoKey, nil), nil
}

// DecryptToken attempts to decrypt the given confirmation token using the given
// confirmation key and extracts the contents of the token into a Confirmation.
// The given curve is used to parse ECDH public and private key strings within
// the token. The same curve used to generate the ECDH keys must be used here.
func DecryptToken(confirmToken string, confirmKey *ecdh.PrivateKey, curve dc.ECDHCurve) (*Confirmation, error) {
	pasetoKey, err := paseto.V4SymmetricKeyFromBytes(confirmKey.Bytes())
	if err != nil {
		return nil, err
	}

	// Decrypt PASETO
	parser := paseto.NewParserWithoutExpiryCheck() // Expiry check is expected to be done elsewhere
	parsedToken, err := parser.ParseV4Local(pasetoKey, confirmToken, nil)
	if err != nil {
		return nil, err
	}

	// Extract dApp ID
	dappIdB64, err := parsedToken.GetString(DappIdClaimName)
	if err != nil {
		return nil, err
	}
	dappIdBytes, err := base64.StdEncoding.DecodeString(dappIdB64)
	if err != nil {
		return nil, err
	}
	dappId, err := curve.NewPublicKey(dappIdBytes)
	if err != nil {
		return nil, err
	}

	// Extract session key
	sessionKeyB64, err := parsedToken.GetString(SessionKeyClaimName)
	if err != nil {
		return nil, err
	}
	sessionKeyBytes, err := base64.StdEncoding.DecodeString(sessionKeyB64)
	if err != nil {
		return nil, err
	}
	sessionKey, err := curve.NewPrivateKey(sessionKeyBytes)
	if err != nil {
		return nil, err
	}

	// Extract confirmation code
	code, err := parsedToken.GetString(ConfirmCodeClaimName)
	if err != nil {
		return nil, err
	}

	// Extract expiration
	exp, err := parsedToken.GetExpiration()
	if err != nil {
		return nil, err
	}

	return NewConfirmation(dappId, sessionKey, confirmKey, code, exp), nil
}
