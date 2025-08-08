package session

import (
	"crypto/ecdh"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	_ "github.com/marcboeker/go-duckdb"

	dc "duckysigner/internal/dapp_connect"
)

const (
	// DefaultConfigFile is the default file name for the session configuration
	// file
	DefaultConfigFile = "session_config.json" // TODO: Encrypt this file to ensure privacy and prevent tampering
	// DefaultSessionsFile is the default file name for the database file where
	// established sessions are stored
	DefaultSessionsFile = "sessions.parquet"
	// DefaultConfirmsFile is the default file name for the database file where
	// pending confirmations are stored
	DefaultConfirmsFile = "confirms.parquet"
	// DefaultDataDir is the name of the directory where data files, such as the
	// sessions database file and the confirmations database file, are stored
	DefaultDataDir = "./dapp_connect"
	// DefaultSessionLifetime is the default amount of time a session lasts
	// before expiring
	DefaultSessionLifetime = 7 * 24 * time.Hour // 1 week
	// DefaultConfirmLifetime is the default amount of time an outstanding
	// confirmation can last before expiring
	DefaultConfirmLifetime = 10 * time.Minute

	// dataDirPermission is the OS file permissions used for the data directory
	// when it is created
	dataDirPermissions = 0700

	// SessionExistsErrMsg is the error message text for when a session already
	// exists
	SessionExistsErrMsg = "session already exists"
	// NoSessionGivenErrMsg is the error message text for when no session is
	// provided
	NoSessionGivenErrMsg = "no session was given"
	// NoSessionKeyGivenErrMsg is the error message text for when no session key
	// is given
	NoSessionKeyGivenErrMsg = "no session key was given"
	// NoDappIdGivenErrMsg is the error message text for when no dApp ID is
	// given
	NoDappIdGivenErrMsg = "no dApp ID was given"
)

// SessionConfig is used to configure the session manager when creating a new
// session manager
type SessionConfig struct {
	// File name of the database file where the established sessions are stored
	SessionsFile string `json:"sessions_file,omitempty"`
	// File name of the database file where the pending confirmations are stored
	ConfirmsFile string `json:"confirms_file,omitempty"`
	// Name of the directory where the data files (e.g. database files) are
	// stored
	DataDir string `json:"data_dir,omitempty"`
	// Amount of time a session lasts
	SessionLifetimeSecs uint64 `json:"session_lifetime_secs"`
	// Amount of time an outstanding confirmation can last
	ConfirmLifetimeSecs uint64 `json:"confirm_lifetime_secs"`
}

// Manager is the dApp connect session manager
type Manager struct {
	// The ECDH curve to use for generating a keys or processing stored keys
	curve dc.ECDHCurve
	// File name of the database file where the established sessions are stored
	sessionsFile string
	// File name of the database file where the pending confirmations are stored
	confirmsFile string
	// Name of the directory where the data files (e.g. database files) are
	// stored
	dataDir string
	// Amount of time a session lasts
	sessionLifetime time.Duration
	// Amount of time an outstanding confirmation can last
	confirmLifetime time.Duration
}

// sessionsCreateTblSQL is the SQL statement for created the `sessions` database
// table
const sessionsCreateTblSQL = `
CREATE TABLE sessions (
    id VARCHAR PRIMARY KEY,
    key BLOB NOT NULL,
    expiry TIMESTAMP_S NOT NULL,
    est TIMESTAMP_S NOT NULL,
    dapp_id VARCHAR NOT NULL,
    dapp_name VARCHAR,
    dapp_url VARCHAR,
    dapp_desc VARCHAR,
    dapp_icon BLOB
);
`

// addParquetKeySQL is the SQL statement for setting the Parquet file encryption
// key within DuckDB. The file encryption key is used for encrypting and
// decrypting all the Parquet files that are used as the database files.
const addParquetKeySQL = "PRAGMA add_parquet_key('key', '%s');"

// sessionsWriteToDbFileSQL is the SQL statement for directing DuckDB to write
// the temporary sessions table to a new file (or overwrite the file if it
// exists)
const sessionsWriteToDbFileSQL = "COPY sessions TO '%s' (ENCRYPTION_CONFIG {footer_key: 'key'});"

// sessionsSimpleInsertSQL is the SQL statement for inserting a session into the
// temporary sessions table
const sessionsSimpleInsertSQL = "INSERT INTO sessions VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)"

// sessionsAddToDbFileSQL is the SQL statement for adding sessions from the
// temporary sessions table to the sessions database file.
// NOTE: Sorting by ID tends to reduce the file size for some reason (Maybe due
// to compression algorithm?)
const sessionsAddToDbFileSQL = `
COPY (
    (FROM read_parquet('%s', encryption_config = {footer_key: 'key'})
    UNION
    FROM sessions)
)
TO '%s' (ENCRYPTION_CONFIG {footer_key: 'key'});
`

// findSessionByIdSQL is the SQL statement for finding a session within the
// sessions database file by ID
const findSessionByIdSQL = `
FROM read_parquet('%s', encryption_config = {footer_key: 'key'})
WHERE id = ?
`

// allSessionsSQL is the SQL statement for getting all stored sessions
const allSessionsSQL = "FROM read_parquet('%s', encryption_config = {footer_key: 'key'})"

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

		// If no data directory is given, interpret it as wanting the directory
		// to be the current directory
		if dataDir == "" {
			dataDir = "."
		}

		// If no session lifetime was given
		if sessionLife == time.Duration(0) {
			sessionLife = DefaultSessionLifetime
		}
		// If no confirmation lifetime was given
		if confirmLife == time.Duration(0) {
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
		curve:           curve,
		sessionLifetime: sessionLife,
		confirmLifetime: confirmLife,
		// URL encode directory and files names to prevent SQL injection. Leave
		// "/" unescaped in data directory name.
		dataDir:      strings.Join(strings.Split(url.QueryEscape(dataDir), "%2F"), "/"),
		sessionsFile: url.QueryEscape(sessionsFile),
		confirmsFile: url.QueryEscape(confirmsFile),
	}
}

// SessionsFile returns the file name of the database file for storing
// established sessions
func (sm *Manager) SessionsFile() string {
	return sm.sessionsFile
}

// ConfirmationsFile returns the file name of the database file for storing
// pending confirmations
func (sm *Manager) ConfirmationsFile() string {
	return sm.confirmsFile
}

// DataDir returns the data directory used for storing the database files
func (sm *Manager) DataDir() string {
	return sm.dataDir
}

// SessionLifetime returns the amount of time a session lasts
func (sm *Manager) SessionLifetime() time.Duration {
	return sm.sessionLifetime
}

// ConfirmLifetime returns the amount of time an outstanding confirmation can
// last
func (sm *Manager) ConfirmLifetime() time.Duration {
	return sm.confirmLifetime
}

// GenerateSession creates a new session by generating a new session key pair
// for the dApp with the given ID with the given dApp data
func (sm *Manager) GenerateSession(dappId *ecdh.PublicKey, dappData *dc.DappData) (session *Session, err error) {
	// Generate session key pair
	sessionKey, err := sm.curve.GenerateKey(rand.Reader)
	if err != nil {
		return
	}

	exp := time.Now().Add(sm.SessionLifetime())
	session = &Session{
		key:      sessionKey,
		exp:      exp,
		dappId:   dappId,
		dappData: dappData,
	}

	return
}

// GetSession attempt to retrieve the stored session with the given ID (in
// base64) using the given file encryption key to decrypt the sessions database
// file. Returns nil without an error if no session with the given ID is found.
func (sm *Manager) GetSession(sessionId string, fileEncKey []byte) (*Session, error) {
	db, err := sm.OpenSessionsDb(fileEncKey)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	sessionsFilePath, err := sm.getSessionsFilePath()
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Run SQL query for finding session by ID
	var (
		retrievedSessionId       string
		retrievedSessionKeyBytes []byte
		retrievedExp             time.Time
		retrievedEst             time.Time
		retrievedDappIdB64       string
		retrievedDappName        string
		retrievedDappURL         string
		retrievedDappDesc        string
		retrievedDappIcon        []byte
	)
	sessionRow := db.QueryRow(
		fmt.Sprintf(findSessionByIdSQL, sessionsFilePath),
		sessionId,
	)
	err = sessionRow.Scan(
		&retrievedSessionId,
		&retrievedSessionKeyBytes,
		&retrievedExp,
		&retrievedEst,
		&retrievedDappIdB64,
		&retrievedDappName,
		&retrievedDappURL,
		&retrievedDappDesc,
		&retrievedDappIcon,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		// An unexpected error occurred
		return nil, err
	}

	return sm.rowToSession(
		retrievedSessionKeyBytes,
		retrievedExp,
		retrievedEst,
		retrievedDappIdB64,
		retrievedDappName,
		retrievedDappURL,
		retrievedDappDesc,
		retrievedDappIcon,
	)
}

// GetAllSessions attempts to retrieve all stored sessions using the given file
// encryption key to decrypt the sessions database file.
func (sm *Manager) GetAllSessions(fileEncKey []byte) ([]*Session, error) {
	db, err := sm.OpenSessionsDb(fileEncKey)
	if err != nil {
		return []*Session{}, err
	}
	defer db.Close()

	sessionsFilePath, err := sm.getSessionsFilePath()
	if os.IsNotExist(err) {
		return []*Session{}, nil
	}
	if err != nil {
		return []*Session{}, err
	}

	sessionsRows, err := db.Query(fmt.Sprintf(allSessionsSQL, sessionsFilePath))
	if err != nil {
		return []*Session{}, err
	}

	var retrievedSessions []*Session

	// Convert each row into a Session
	for sessionsRows.Next() {
		var (
			retrievedSessionId       string
			retrievedSessionKeyBytes []byte
			retrievedExp             time.Time
			retrievedEst             time.Time
			retrievedDappIdB64       string
			retrievedDappName        string
			retrievedDappURL         string
			retrievedDappDesc        string
			retrievedDappIcon        []byte
		)

		err := sessionsRows.Scan(
			&retrievedSessionId,
			&retrievedSessionKeyBytes,
			&retrievedExp,
			&retrievedEst,
			&retrievedDappIdB64,
			&retrievedDappName,
			&retrievedDappURL,
			&retrievedDappDesc,
			&retrievedDappIcon,
		)
		if err != nil {
			// An unexpected error occurred
			// Return the incomplete set along with the error
			return retrievedSessions, err
		}

		session, err := sm.rowToSession(
			retrievedSessionKeyBytes,
			retrievedExp,
			retrievedEst,
			retrievedDappIdB64,
			retrievedDappName,
			retrievedDappURL,
			retrievedDappDesc,
			retrievedDappIcon,
		)
		if err != nil {
			// Return the incomplete set along with the error
			return retrievedSessions, err
		}

		retrievedSessions = append(retrievedSessions, session)
	}

	return retrievedSessions, nil
}

// StoreSession attempts to store the given session using the given file
// encryption key to access the sessions database file
func (sm *Manager) StoreSession(session *Session, fileEncKey []byte) (err error) {
	if session == nil {
		return errors.New(NoSessionGivenErrMsg)
	}

	if session.key == nil {
		return errors.New(NoSessionKeyGivenErrMsg)
	}

	if session.dappId == nil {
		return errors.New(NoDappIdGivenErrMsg)
	}

	db, err := sm.OpenSessionsDb(fileEncKey)
	if err != nil {
		return err
	}
	defer db.Close()

	// Convert some of the session data for storage
	b64Encoder := base64.StdEncoding
	sessionKey := session.Key()
	sessionIdB64 := b64Encoder.EncodeToString(sessionKey.PublicKey().Bytes())
	dappIdB64 := b64Encoder.EncodeToString(session.DappId().Bytes())
	// Ensure dApp data is not nil
	var dappData *dc.DappData
	if session.dappData == nil {
		dappData = &dc.DappData{}
	} else {
		dappData = session.dappData
	}
	// The dApp icon needs to be stored as bytes, so decode the base64
	dappIconB64, decodeErr := b64Encoder.DecodeString(dappData.Icon)
	if decodeErr != nil {
		return decodeErr
	}

	// Create temporary table in memory
	_, err = db.Exec(sessionsCreateTblSQL)
	if err != nil {
		return
	}

	// Insert session into temporary table
	_, err = db.Exec(sessionsSimpleInsertSQL,
		sessionIdB64,
		sessionKey.Bytes(),
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

	sessionsFilePath, err := sm.getSessionsFilePath()
	if err != nil && !os.IsNotExist(err) {
		return
	}

	// Write the session data in the temporary table into the file
	if !os.IsNotExist(err) { // If session file exists
		// Check if session is already stored in the file
		sessionRow := db.QueryRow(
			fmt.Sprintf(findSessionByIdSQL, sessionsFilePath),
			sessionIdB64,
		)
		if sessionRow.Scan() != sql.ErrNoRows {
			return errors.New(SessionExistsErrMsg)
		}

		// Writing a new session into an existing Parquet file is a little more
		// complex than writing a session to a new Parquet file
		_, err = db.Exec(fmt.Sprintf(sessionsAddToDbFileSQL, sessionsFilePath, sessionsFilePath))
		if err != nil {
			return
		}
	} else { // Session file does not exist
		// Create a new file and write the session data into it
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

// OpenSessionsDb is a helper function that opens the database connection for
// the sessions database and sets the file encryption key that will be used to
// encrypt and decrypt database file(s). This function usually does not need to
// be called directly outside of testing. The file encryption key is only set
// within the database connection. The key is NOT checked if it is correct.
// Returns a database handle if there are no errors.
func (sm *Manager) OpenSessionsDb(fileEncKey []byte) (db *sql.DB, err error) {
	// Open DuckDB in in-memory mode
	db, err = sql.Open("duckdb", "")
	if err != nil {
		return
	}

	// Add key for decrypting and encrypting file
	// NOTE: This PRAGMA statement does not work as a prepared statement, but
	// there is no risk of SQL injection in this case
	_, err = db.Exec(fmt.Sprintf(addParquetKeySQL, base64.StdEncoding.EncodeToString(fileEncKey)))
	if err != nil {
		db.Close() // Fatal error, so close the database connection
		return nil, err
	}

	return
}

// getSessionsFilePath ALWAYS returns the file path of the sessions database
// file. If the data directory does not exist, it tries to created it. It also
// checks if the file exists. If the file does not exist, an os.ErrNotExist is
// returned as the error.
func (sm *Manager) getSessionsFilePath() (filePath string, err error) {
	filePath = sm.DataDir() + "/" + sm.SessionsFile()

	// Ensure data directory exists
	err = os.Mkdir(sm.DataDir(), dataDirPermissions)
	if err != nil && !os.IsExist(err) {
		return
	}

	// Check if session file exists
	_, err = os.Stat(filePath)
	if err != nil {
		// Could not access file some reason other than that it does not exist
		// (e.g. permissions, drive failure)
		return
	}

	return
}

// rowToSession converts the given data of a sessions database row into a
// Session
func (sm *Manager) rowToSession(
	sessionKeyBytes []byte,
	exp time.Time,
	est time.Time,
	dappIdB64 string,
	dappName string,
	dappURL string,
	dappDesc string,
	dappIcon []byte,
) (*Session, error) {
	// Convert session key bytes to an ECDH private key
	retrievedSessionKey, err := sm.curve.NewPrivateKey(sessionKeyBytes)
	if err != nil {
		return nil, err
	}

	// Convert dApp ID base64-encoded bytes to an ECDH public key
	retrievedDappIdBytes, err := base64.StdEncoding.DecodeString(dappIdB64)
	if err != nil {
		return nil, err
	}
	retrievedDappId, err := sm.curve.NewPublicKey(retrievedDappIdBytes)
	if err != nil {
		return nil, err
	}

	return &Session{
		key:           retrievedSessionKey,
		exp:           exp,
		establishedAt: est,
		dappId:        retrievedDappId,
		dappData: &dc.DappData{
			Name:        dappName,
			URL:         dappURL,
			Description: dappDesc,
			Icon:        base64.StdEncoding.EncodeToString(dappIcon),
		},
	}, nil
}
