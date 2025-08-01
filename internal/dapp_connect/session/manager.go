package session

import (
	"crypto/ecdh"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"os"
	"time"

	// "database/sql"
	_ "github.com/marcboeker/go-duckdb"

	dc "duckysigner/internal/dapp_connect"
)

// TODO: Add documentation
const (
	DefaultConfigFile      = "session_config.json" // TODO: Encrypt this file to ensure privacy and prevent tampering
	DefaultSessionsFile    = "sessions.parquet"
	DefaultConfirmsFile    = "confirms.parquet"
	DefaultDataDir         = "./dapp_connect"
	DefaultSessionLifetime = 7 * 24 * time.Hour // 1 week
	DefaultConfirmLifetime = 10 * time.Minute

	dataDirPermissions = 0700
)

// TODO: Add documentation
type SessionConfig struct {
	SessionsFile        string `json:"sessions_file,omitempty"`
	ConfirmsFile        string `json:"confirms_file,omitempty"`
	DataDir             string `json:"data_dir,omitempty"`
	SessionLifetimeSecs uint64 `json:"session_lifetime_secs"`
	ConfirmLifetimeSecs uint64 `json:"confirm_lifetime_secs"`
}

// Manager is the dApp connect session manager
type Manager struct {
	// TODO: Add documentation
	Curve           dc.ECDHCurve
	SessionsFile    string
	ConfirmsFile    string
	DataDir         string
	SessionLifetime time.Duration
	ConfirmLifetime time.Duration
}

// SQL statement for created the `sessions` database table
const sessionsCreateTblSQL = `
CREATE TABLE sessions (
    id VARCHAR PRIMARY KEY,
    key VARCHAR NOT NULL,
    expiry TIMESTAMP_S NOT NULL,
    est TIMESTAMP_S NOT NULL,
    dapp_id VARCHAR NOT NULL,
    dapp_name VARCHAR,
    dapp_url VARCHAR,
    dapp_desc VARCHAR,
    dapp_icon BLOB
);
`
const addParquetKeySQL = "PRAGMA add_parquet_key('key', '%s');"
const sessionsWriteToDbFileSQL = "COPY sessions TO '%s' (ENCRYPTION_CONFIG {footer_key: 'key'});"
const sessionsSimpleInsertSQL = "INSERT INTO sessions VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)"

// NOTE: Sorting by ID reduces the files size for some reason (Maybe due to compression algorithm?)
const sessionsParquetInsertSQL = `
COPY (
    (FROM read_parquet('%s', encryption_config = {footer_key: 'key'})
    UNION
    FROM sessions)
)
TO '%s' (ENCRYPTION_CONFIG {footer_key: 'key'});
`

// NewManager creates a new session manager using the given configuration for
// creating sessions. The given ECDH curve will be used by this new session
// manager to generate session keys.
func NewManager(curve dc.ECDHCurve, sessionConfig *SessionConfig) *Manager {
	var (
		sessionLife  time.Duration
		confirmLife  time.Duration
		dataDir      string
		sessionsFile string
		confirmsFile string
	)

	// Use default config if no config was given
	if sessionConfig == nil {
		sessionLife = DefaultSessionLifetime
		confirmLife = DefaultConfirmLifetime
		dataDir = DefaultDataDir
		sessionsFile = DefaultSessionsFile
		confirmsFile = DefaultConfirmsFile
	} else {
		sessionLife = time.Duration(sessionConfig.SessionLifetimeSecs) * time.Second
		confirmLife = time.Duration(sessionConfig.ConfirmLifetimeSecs) * time.Second
		dataDir = sessionConfig.DataDir
		sessionsFile = sessionConfig.SessionsFile
		confirmsFile = sessionConfig.ConfirmsFile

		// If no session lifetime was given
		if sessionsFile == "" {
			sessionLife = DefaultSessionLifetime
		}
		// If no confirmation lifetime was given
		if confirmsFile == "" {
			confirmLife = DefaultConfirmLifetime
		}

		// If no file name for the sessions database file was given
		if sessionsFile == "" {
			sessionsFile = DefaultSessionsFile
		}
		// If no file name for the confirmations database file was given
		if confirmsFile == "" {
			confirmsFile = DefaultConfirmsFile
		}
	}

	return &Manager{
		Curve:           curve,
		SessionLifetime: sessionLife,
		ConfirmLifetime: confirmLife,
		DataDir:         dataDir,
		SessionsFile:    sessionsFile,
		ConfirmsFile:    confirmsFile,
	}
}

// GenerateSession creates a new session by generating a new session key pair
// for the dApp with the given ID with the given dApp data
func (sm *Manager) GenerateSession(dappId *ecdh.PublicKey, dappData *dc.DappData) (session *Session, err error) {
	// Generate session key pair
	sessionKey, err := sm.Curve.GenerateKey(rand.Reader)
	if err != nil {
		return
	}

	exp := time.Now().Add(sm.SessionLifetime)
	session = &Session{
		key:      sessionKey,
		exp:      exp,
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

// StoreSession attempts to store the given session using the given file
// encryption key to access the sessions database file
func (sm *Manager) StoreSession(session *Session, fileEncKey []byte) (err error) {
	// Ensure data directory exists
	err = os.Mkdir(sm.DataDir, dataDirPermissions)
	if err != nil && !os.IsExist(err) {
		return
	}

	sessionsFilePath := sm.DataDir + "/" + sm.SessionsFile

	// Check if session file exists
	_, err = os.Stat(sessionsFilePath)
	if err != nil && !os.IsNotExist(err) {
		// Could not access file some reason other than that it does not exist
		// (e.g. permissions, drive failure)
		return
	}

	sessionsFileExists := !os.IsNotExist(err)

	// Open DuckDB in in-memory mode
	db, err := sql.Open("duckdb", "")
	if err != nil {
		return
	}
	defer db.Close()

	// Add key for decrypting and encrypting file
	// NOTE: This PRAGMA statement does not work as a prepared statement, but
	// there is no risk of SQL injection
	_, err = db.Exec(fmt.Sprintf(addParquetKeySQL, base64.StdEncoding.EncodeToString(fileEncKey)))
	if err != nil {
		return
	}

	// Convert some of the session data for storage
	b64Encoder := base64.StdEncoding
	sessionKey := session.Key()
	sessionIdB64 := b64Encoder.EncodeToString(sessionKey.PublicKey().Bytes())
	sessionKeyB64 := b64Encoder.EncodeToString(sessionKey.Bytes())
	dappIdB64 := b64Encoder.EncodeToString(session.DappId().Bytes())
	// Ensure dApp data is not nil
	var dappData *dc.DappData
	if session.dappData == nil {
		dappData = &dc.DappData{}
	} else {
		dappData = session.dappData
	}

	// Create new table
	_, err = db.Exec(sessionsCreateTblSQL)
	if err != nil {
		return
	}

	dappIconB64, decodeErr := b64Encoder.DecodeString(dappData.Icon)
	if decodeErr != nil {
		return decodeErr
	}

	// Insert session
	_, err = db.Exec(sessionsSimpleInsertSQL,
		sessionIdB64,
		sessionKeyB64,
		// DuckDB uses ISO8601 format for timestamps
		// Source: <https://duckdb.org/docs/stable/sql/data_types/timestamp>
		session.Expiration().UTC().Format(time.DateTime),
		session.EstablishedAt().UTC().Format(time.DateTime),
		dappIdB64,
		dappData.Name,
		dappData.URL,
		dappData.Description,
		dappIconB64,
	)
	if err != nil {
		return
	}

	// Write to file
	if sessionsFileExists {
		// TODO: Check if session is already stored

		// Writing a new session into an existing Parquet file is a little more
		// complex than writing a session to a new Parquet file
		_, err = db.Exec(fmt.Sprintf(sessionsParquetInsertSQL, sessionsFilePath, sessionsFilePath))
		if err != nil {
			return
		}
	} else {
		_, err = db.Exec(fmt.Sprintf(sessionsWriteToDbFileSQL, sessionsFilePath))
		if err != nil {
			return
		}
	}

	return
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
