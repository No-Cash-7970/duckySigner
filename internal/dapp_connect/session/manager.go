package session

import (
	"crypto/ecdh"
	"crypto/rand"

	dc "duckysigner/internal/dapp_connect"
)

// Manager is the dApp connect session manager
type Manager struct {
	// TODO: Complete this
}

// GenerateSession creates a new session by generating a new session key pair
// for the dApp with the given ID with the given dApp data
func (sm *Manager) GenerateSession(dappId *ecdh.PublicKey, dappData *dc.DappData, curve dc.ECDHCurve) (session *Session, err error) {
	// Generate session key pair
	sessionKey, err := curve.GenerateKey(rand.Reader)
	if err != nil {
		return
	}

	// TODO: Complete this
	session = &Session{
		key: sessionKey,
		// exp: time.Now().Add(time.Duration(service.Config.SessionLifetimeSecs) * time.Second),
		dappId:   dappId,
		dappData: dappData,
	}

	return
}

// GetSession attempt to retrieve the stored session with the given ID
func (sm *Manager) GetSession(sessionId string) (*Session, error) {
	// TODO: Complete this
	return &Session{}, nil
}

// GetAllSessions attempts to retrieve all stored sessions
func (sm *Manager) GetAllSessions() ([]*Session, error) {
	// TODO: Complete this
	return []*Session{}, nil
}

// StoreSession attempts to store the given session
func (sm *Manager) StoreSession(session *Session) error {
	// TODO: Complete this
	return nil
}

// RemoveSession attempts to remove the stored session with the given ID
func (sm *Manager) RemoveSession(sessionId string) error {
	// TODO: Complete this
	return nil
}

// PurgeAllSessions attempts to completely delete all stored sessions. It
// returns the number of sessions that were deleted.
func (sm *Manager) PurgeAllSessions() (int, error) {
	// TODO: Complete this
	return 0, nil
}

// PurgeInvalidSessions attempts to delete all expired or invalid stored
// sessions. It returns the number of sessions that were deleted.
func (sm *Manager) PurgeInvalidSessions() (int, error) {
	// TODO: Complete this
	return 0, nil
}

// StoreConfirmation attempts to store the given confirmation
func (sm *Manager) StoreConfirmation(confirm *Confirmation) error {
	// TODO: Complete this
	return nil
}

// PurgeAllConfirmations attempts to completely delete all confirmations. It
// returns the number of confirmations that were deleted.
func (sm *Manager) PurgeAllConfirmations() (int, error) {
	// TODO: Complete this
	return 0, nil
}
