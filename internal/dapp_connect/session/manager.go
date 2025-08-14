package session

import (
	"crypto/ecdh"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"math/big"
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
	// the keys for pending confirmations are stored
	DefaultConfirmsFile = "confirms.parquet"
	// DefaultDataDir is the name of the directory where data files, such as the
	// sessions database and the confirmation keystore, are stored
	DefaultDataDir = "./dapp_connect"
	// DefaultSessionLifetime is the default amount of time a session lasts
	// before expiring
	DefaultSessionLifetime = 7 * 24 * time.Hour // 1 week
	// DefaultConfirmLifetime is the default amount of time an outstanding
	// confirmation can last before expiring
	DefaultConfirmLifetime = 10 * time.Minute
	// DefaultConfirmCodeCharset is a the set of characters used for generating
	// a confirmation code
	DefaultConfirmCodeCharset = "0123456789"
	// DefaultConfirmCodeLen is the length of a confirmation code
	DefaultConfirmCodeLen = 5

	// dataDirPermission is the OS file permissions used for the data directory
	// when it is created
	dataDirPermissions = 0700
	// tempFileSuffix is the suffix used to create a temporary file. The
	// temporary file name is the original file name + this suffix.
	tempFileSuffix = ".new"

	// sessionsTblName is the name of the in-memory table used to temporarily
	// store new sessions
	sessionsTblName = "sessions"
	// confirmsTblName is the name of the in-memory table used to temporarily
	// store new confirmation keys
	confirmsTblName = "confirms"

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
	// RemoveSessionNotExistErrMsg is the error message text for when there is
	// an attempt to remove a session that is not stored
	RemoveSessionNotStoredErrMsg = "cannot remove session that is not stored"
	// WrongConfirmCodeErrMsg is the error message text for when the given
	// confirmation code does not match the code that is with in the given
	// confirmation token
	WrongConfirmCodeErrMsg = "wrong confirmation code"
	// NoConfirmTokenGivenErrMsg is the error message text for when the given
	// confirmation token is empty (i.e. no token was given)
	NoConfirmTokenGivenErrMsg = "no confirmation token given"
	// ConfirmKeyExistsErrMsg is the error message text for when a confirmation
	// key already exists
	ConfirmKeyExistsErrMsg = "confirmation key already exists"
	// NoConfirmKeyGivenErrMsg is the error message text for when no
	// confirmation key is provided
	NoConfirmKeyGivenErrMsg = "no confirmation key was given"
	// RemoveConfirmKeyNotExistErrMsg is the error message text for when there is
	// an attempt to remove a confirmation key that is not stored
	RemoveConfirmKeyNotStoredErrMsg = "cannot remove confirmation key that is not stored"
)

// SessionConfig is used to configure the session manager when creating a new
// session manager
type SessionConfig struct {
	// File name of the database file where the established sessions are stored
	SessionsFile string `json:"sessions_file,omitempty"`
	// File name of the database file where the keys for pending confirmations
	// are stored
	ConfirmsFile string `json:"confirms_file,omitempty"`
	// Name of the directory where the data files (e.g. database files) are
	// stored
	DataDir string `json:"data_dir,omitempty"`
	// Amount of time a session lasts
	SessionLifetimeSecs uint64 `json:"session_lifetime_secs,omitempty"`
	// Amount of time an outstanding confirmation can last
	ConfirmLifetimeSecs uint64 `json:"confirm_lifetime_secs,omitempty"`
	// The character set used to generate a confirmation code
	ConfirmCodeCharset string `json:"confirm_code_charset,omitempty"`
	// The length of a confirmation code
	ConfirmCodeLen uint `json:"confirm_code_len,omitempty"`
}

// Manager is the dApp connect session manager
type Manager struct {
	// The ECDH curve to use for generating a keys or processing stored keys
	curve dc.ECDHCurve
	// File name of the database file where the established sessions are stored
	sessionsFile string
	// File name of the database file where the keys for pending confirmations
	// are stored
	confirmsFile string
	// Name of the directory where the data files (e.g. database files) are
	// stored
	dataDir string
	// Amount of time a session lasts
	sessionLifetime time.Duration
	// Amount of time an outstanding confirmation can last
	confirmLifetime time.Duration
	// The character set used to generate a confirmation code
	confirmCodeCharset string
	// The length of a confirmation code
	confirmCodeLen uint
}

// sessionsCreateTblSQL is the SQL statement for created the `sessions`
// in-memory database table
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
// Requires the file encryption key in Base64.
const addParquetKeySQL = "PRAGMA add_parquet_key('key', '%s');"

// itemsWriteToDbFileSQL is the SQL statement for directing DuckDB to write
// a in-memory table to a new file (or overwrite the file if it exists).
// Requires the in-memory table name and the name of the new file.
const itemsWriteToDbFileSQL = `
COPY %s TO '%s' (FORMAT parquet, ENCRYPTION_CONFIG {footer_key: 'key'});
`

// sessionSimpleInsertSQL is the SQL statement for inserting a session into a
// in-memory table.
const sessionSimpleInsertSQL = "INSERT INTO sessions VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)"

// itemsAddToDbFileSQL is the SQL statement for adding items (e.g. sessions,
// confirmation keys) from a in-memory table to a database file. Adding an item
// to a database file is done by combining the in-memory table and the items
// already stored in the file, and then writing that combination to a new file
// (or overwriting the existing file).
// NOTE: Sorting by ID tends to reduce the file size for some reason (Maybe due
// to compression algorithm?). Requires the database file name, the name of the
// in-memory table and the new database (if overwriting the database file, use
// the name of the old file).
const itemsAddToDbFileSQL = `
COPY (
    (FROM read_parquet('%s', encryption_config = {footer_key: 'key'})
    UNION
    FROM %s)
)
TO '%s' (FORMAT parquet, ENCRYPTION_CONFIG {footer_key: 'key'});
`

// findItemByIdSQL is the SQL statement for finding an item (e.g. session)
// within a database file by ID. Requires the database file name.
const findItemByIdSQL = `
FROM read_parquet('%s', encryption_config = {footer_key: 'key'})
WHERE id = ?
`

// allItemsSQL is the SQL statement for getting all stored items (eg. sessions,
// confirmation keys). Requires the database file name.
const allItemsSQL = "FROM read_parquet('%s', encryption_config = {footer_key: 'key'})"

// removeItemSQL is the SQL statement for removing an item (e.g. session,
// confirmation key) from a database file. Removing an item is done by copying
// all items except for the item to be removed to a new file (or overwriting the
// existing file). Requires the database file name, the ID of the item to
// remove, and the file name of the new database (if overwriting the database
// file, use the name of the old file).
const removeItemSQL = `
COPY (
    FROM read_parquet('%s', encryption_config = {footer_key: 'key'})
    WHERE id != ?
) TO '%s' (FORMAT parquet, ENCRYPTION_CONFIG {footer_key: 'key'});
`

// countItemsSQL is the SQL statement for counting the number of items stored in
// a database file.
const countItemsSQL = `
SELECT count(id) FROM read_parquet('%s', encryption_config = {footer_key: 'key'})
`

// removeExpiredItemsSQL is the SQL statement for removing all expired items
// (e.g. sessions, confirmation keys) from a database file. Removing all expired
// items is done by copying all items except the invalid items to a new file (or
// overwriting the existing file). Requires the database file name, the current
// date-time, and the file name of the new database (if overwriting the database
// file, use the name of the old file).
const removeExpiredItemsSQL = `
COPY (
    FROM read_parquet('%s', encryption_config = {footer_key: 'key'})
    WHERE expiry > ?
) TO '%s' (FORMAT parquet, ENCRYPTION_CONFIG {footer_key: 'key'});
`

// confirmsCreateTblSQL is the SQL statement for created the `confirms`
// in-memory database table
const confirmsCreateTblSQL = `
CREATE TABLE confirms (
    id VARCHAR PRIMARY KEY,
    key BLOB NOT NULL
);
`

// confirmSimpleInsertSQL is the SQL statement for inserting a confirmation key
// pair into a in-memory table.
const confirmSimpleInsertSQL = "INSERT INTO confirms VALUES (?, ?)"

/*******************************************************************************
 * Manager
 ******************************************************************************/

// NewManager creates a new session manager using the given configuration for
// creating sessions. The given ECDH curve will be used by this new session
// manager to generate session keys.
func NewManager(curve dc.ECDHCurve, sessionConfig *SessionConfig) *Manager {
	var (
		sessionLife        time.Duration
		confirmLife        time.Duration
		dataDir            string
		sessionsFile       string
		confirmsFile       string
		confirmCodeCharset string
		confirmCodeLen     uint
	)

	// Use default config if no config was given
	if sessionConfig == nil {
		sessionLife = DefaultSessionLifetime
		confirmLife = DefaultConfirmLifetime
		dataDir = DefaultDataDir
		sessionsFile = DefaultSessionsFile
		confirmsFile = DefaultConfirmsFile
		confirmCodeCharset = DefaultConfirmCodeCharset
		confirmCodeLen = DefaultConfirmCodeLen
	} else {
		sessionLife = time.Duration(sessionConfig.SessionLifetimeSecs) * time.Second
		confirmLife = time.Duration(sessionConfig.ConfirmLifetimeSecs) * time.Second
		dataDir = sessionConfig.DataDir
		sessionsFile = sessionConfig.SessionsFile
		confirmsFile = sessionConfig.ConfirmsFile
		confirmCodeCharset = sessionConfig.ConfirmCodeCharset
		confirmCodeLen = sessionConfig.ConfirmCodeLen

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

		// If no confirmation code character set was given
		if confirmCodeCharset == "" {
			confirmCodeCharset = DefaultConfirmCodeCharset
		}

		// If no confirmation code length was given
		if confirmCodeLen == 0 {
			confirmCodeLen = DefaultConfirmCodeLen
		}
	}

	return &Manager{
		curve:           curve,
		sessionLifetime: sessionLife,
		confirmLifetime: confirmLife,
		// URL encode directory and files names to prevent SQL injection. Leave
		// "/" unescaped in data directory name.
		dataDir:            strings.Join(strings.Split(url.QueryEscape(dataDir), "%2F"), "/"),
		sessionsFile:       url.QueryEscape(sessionsFile),
		confirmsFile:       url.QueryEscape(confirmsFile),
		confirmCodeCharset: confirmCodeCharset,
		confirmCodeLen:     confirmCodeLen,
	}
}

// SessionsFile returns the file name of the database file for storing
// established sessions
func (sm *Manager) SessionsFile() string {
	return sm.sessionsFile
}

// ConfirmationsFile returns the file name of the database file for storing the
// keys for pending confirmations
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

// ConfirmCodeCharset returns the charset used to generate a confirmation code
func (sm *Manager) ConfirmCodeCharset() string {
	return sm.confirmCodeCharset
}

// ConfirmCodeLen returns the charset used to generate a confirmation code
func (sm *Manager) ConfirmCodeLen() uint {
	return sm.confirmCodeLen
}

/*******************************************************************************
 * Managing sessions
 ******************************************************************************/

// GenerateSession creates a new unestablished session (with no established-at
// date-time) by generating a new session key pair for the dApp with the given
// ID with the given dApp data
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

// GetSession attempts to retrieve the stored session with the given ID (in
// base64) using the given file encryption key to decrypt the sessions database
// file. Returns nil without an error if no session with the given ID is found.
func (sm *Manager) GetSession(sessionId string, fileEncKey []byte) (*Session, error) {
	sessionsFilePath, err := sm.getSessionsFilePath()
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	db, err := sm.OpenDb(fileEncKey)
	if err != nil {
		return nil, err
	}
	defer db.Close()

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
	sessionRow := db.QueryRow(fmt.Sprintf(findItemByIdSQL, sessionsFilePath), sessionId)
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
	sessionsFilePath, err := sm.getSessionsFilePath()
	if os.IsNotExist(err) {
		return []*Session{}, nil
	}
	if err != nil {
		return []*Session{}, err
	}

	db, err := sm.OpenDb(fileEncKey)
	if err != nil {
		return []*Session{}, err
	}
	defer db.Close()

	sessionsRows, err := db.Query(fmt.Sprintf(allItemsSQL, sessionsFilePath))
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

	db, err := sm.OpenDb(fileEncKey)
	if err != nil {
		return err
	}
	defer db.Close()

	// Convert some of the session data for storage
	b64Encoder := base64.StdEncoding
	sessionIdB64 := b64Encoder.EncodeToString(session.key.PublicKey().Bytes())
	dappIdB64 := b64Encoder.EncodeToString(session.dappId.Bytes())
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
	_, err = db.Exec(sessionSimpleInsertSQL,
		sessionIdB64,
		session.key.Bytes(),
		// DuckDB uses ISO8601 format for timestamps
		// Source: <https://duckdb.org/docs/stable/sql/data_types/timestamp>
		session.exp.UTC().Format(time.DateTime),
		session.establishedAt.UTC().Format(time.DateTime),
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
			fmt.Sprintf(findItemByIdSQL, sessionsFilePath),
			sessionIdB64,
		)
		if sessionRow.Scan() != sql.ErrNoRows {
			return errors.New(SessionExistsErrMsg)
		}

		// Combine the sessions within the file with the new session
		_, err = db.Exec(fmt.Sprintf(
			itemsAddToDbFileSQL, sessionsFilePath, sessionsTblName, sessionsFilePath+tempFileSuffix,
		))
		if err != nil {
			return
		}

		err = sm.removeTempFile(sessionsFilePath)
		if err != nil {
			return err
		}
	} else { // Session file does not exist
		// Create a new file and write the session data into it
		_, err = db.Exec(fmt.Sprintf(itemsWriteToDbFileSQL, sessionsTblName, sessionsFilePath))
		if err != nil {
			return
		}
	}

	return
}

// RemoveSession attempts to remove the stored session with the given ID
func (sm *Manager) RemoveSession(sessionId string, fileEncKey []byte) error {
	sessionsFilePath, err := sm.getSessionsFilePath()
	if os.IsNotExist(err) {
		return errors.New(RemoveSessionNotStoredErrMsg)
	}
	if err != nil {
		return err
	}

	db, err := sm.OpenDb(fileEncKey)
	if err != nil {
		return err
	}
	defer db.Close()

	// Remove session
	_, err = db.Exec(fmt.Sprintf(removeItemSQL, sessionsFilePath, sessionsFilePath+tempFileSuffix), sessionId)
	if err != nil {
		return err
	}

	err = sm.removeTempFile(sessionsFilePath)
	if err != nil {
		return err
	}

	return nil
}

// PurgeAllSessions attempts to completely delete all stored sessions. It
// returns the number of sessions that were deleted.
func (sm *Manager) PurgeAllSessions(fileEncKey []byte) (numPurged uint, err error) {
	sessionsFilePath, err := sm.getSessionsFilePath()
	if err != nil {
		return 0, nil
	}

	db, err := sm.OpenDb(fileEncKey)
	if err != nil {
		return 0, err
	}
	defer db.Close()

	// Count number of sessions in the file
	row := db.QueryRow(fmt.Sprintf(countItemsSQL, sessionsFilePath))
	row.Scan(&numPurged)
	if err != nil {
		return 0, err
	}

	// Remove the file
	err = os.Remove(sessionsFilePath)
	if err != nil && !os.IsNotExist(err) {
		return 0, err
	}

	return
}

// PurgeExpiredSessions attempts to delete all expired stored sessions. It
// returns the number of sessions that were deleted.
func (sm *Manager) PurgeExpiredSessions(fileEncKey []byte) (uint, error) {
	sessionsFilePath, err := sm.getSessionsFilePath()
	if err != nil {
		return 0, nil
	}

	db, err := sm.OpenDb(fileEncKey)
	if err != nil {
		return 0, err
	}
	defer db.Close()

	// Count number of sessions in the file before removing anything
	row := db.QueryRow(fmt.Sprintf(countItemsSQL, sessionsFilePath))
	var totalSessions uint
	err = row.Scan(&totalSessions)
	if err != nil {
		return 0, err
	}

	// Remove expired sessions by exclusion
	row = db.QueryRow(
		fmt.Sprintf(removeExpiredItemsSQL, sessionsFilePath, sessionsFilePath+tempFileSuffix),
		time.Now().UTC(),
	)
	var numValid uint
	row.Scan(&numValid)

	err = sm.removeTempFile(sessionsFilePath)
	if err != nil {
		return 0, err
	}

	return (totalSessions - numValid), nil
}

// EstablishSession creates a new established session using the given dApp data
// after checking the given confirmation token, code and key
func (sm *Manager) EstablishSession(
	token string,
	code string,
	key *ecdh.PrivateKey,
	dappData *dc.DappData,
) (*Session, error) {
	// Check token
	if token == "" {
		return nil, errors.New(NoConfirmTokenGivenErrMsg)
	}

	// Decrypt token
	confirm, err := DecryptToken(token, key, sm.curve)
	if err != nil {
		return nil, err
	}

	// Check code
	if code != confirm.code {
		return nil, errors.New(WrongConfirmCodeErrMsg)
	}

	// Create the new established session
	now := time.Now()
	session := Session{
		key:           confirm.sessionKey,
		dappId:        confirm.dappId,
		dappData:      dappData,
		establishedAt: now,
		exp:           now.Add(sm.sessionLifetime),
	}

	return &session, nil
}

/*******************************************************************************
 * Managing confirmations
 ******************************************************************************/

// GenerateConfirmation creates a new confirmation by generating a new
// confirmation key pair for the dApp with the given ID
func (sm *Manager) GenerateConfirmation(dappId *ecdh.PublicKey) (confirm *Confirmation, err error) {
	// Check dApp ID
	if dappId == nil {
		return nil, errors.New(NoDappIdGivenErrMsg)
	}

	// Generate confirmation key pair
	confirmKey, err := sm.curve.GenerateKey(rand.Reader)
	if err != nil {
		return
	}

	// Generate session key pair
	sessionKey, err := sm.curve.GenerateKey(rand.Reader)
	if err != nil {
		return
	}

	// Generate confirmation code
	code, err := sm.generateConfirmationCode()
	if err != nil {
		return
	}

	confirm = NewConfirmation(
		dappId,
		sessionKey,
		confirmKey,
		code,
		time.Now().Add(sm.confirmLifetime),
	)

	return
}

// GetConfirmKey attempts to retrieve the stored confirmation key with the given
// ID (in base64) using the given file encryption key to decrypt the
// confirmation keystore file. Returns nil without an error if no confirmation
// with the given ID is found.
func (sm *Manager) GetConfirmKey(confirmId string, fileEncKey []byte) (*ecdh.PrivateKey, error) {
	confirmsFilePath, err := sm.getConfirmsFilePath()
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	db, err := sm.OpenDb(fileEncKey)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	// Retrieve confirmation key from database
	var retrievedId string
	var retrievedKeyBytes []byte
	row := db.QueryRow(fmt.Sprintf(findItemByIdSQL, confirmsFilePath), confirmId)
	err = row.Scan(&retrievedId, &retrievedKeyBytes)
	if err != nil {
		return nil, err
	}

	confirmKey, err := sm.curve.NewPrivateKey(retrievedKeyBytes)
	if err != nil {
		return nil, err
	}

	return confirmKey, nil
}

// GetAllConfirmKeys attempts to retrieve all stored confirmation keys using the
// given file encryption key to decrypt the confirmation keystore file.
func (sm *Manager) GetAllConfirmKeys(fileEncKey []byte) ([]*ecdh.PrivateKey, error) {
	confirmsFilePath, err := sm.getConfirmsFilePath()
	if os.IsNotExist(err) {
		return []*ecdh.PrivateKey{}, nil
	}
	if err != nil {
		return []*ecdh.PrivateKey{}, err
	}

	db, err := sm.OpenDb(fileEncKey)
	if err != nil {
		return []*ecdh.PrivateKey{}, err
	}
	defer db.Close()

	rows, err := db.Query(fmt.Sprintf(allItemsSQL, confirmsFilePath))
	if err != nil {
		return []*ecdh.PrivateKey{}, err
	}

	var retrievedKeys []*ecdh.PrivateKey

	// Convert each row into an ECDH private key
	for rows.Next() {
		var id string
		var keyBytes []byte

		err := rows.Scan(&id, &keyBytes)
		if err != nil {
			// An unexpected error occurred
			// Return the incomplete set along with the error
			return retrievedKeys, err
		}

		key, err := sm.curve.NewPrivateKey(keyBytes)
		if err != nil {
			// Return the incomplete set along with the error
			return retrievedKeys, err
		}

		retrievedKeys = append(retrievedKeys, key)
	}

	return retrievedKeys, nil
}

// StoreConfirmKey attempts to store the given confirmation key using the given
// file encryption key to access the confirmation keystore file
func (sm *Manager) StoreConfirmKey(key *ecdh.PrivateKey, fileEncKey []byte) error {
	if key == nil {
		return errors.New(NoConfirmKeyGivenErrMsg)
	}

	db, err := sm.OpenDb(fileEncKey)
	if err != nil {
		return err
	}
	defer db.Close()

	// Create temporary table in memory
	_, err = db.Exec(confirmsCreateTblSQL)
	if err != nil {
		return err
	}

	idB64 := base64.StdEncoding.EncodeToString(key.PublicKey().Bytes())

	// Insert confirmation key pair into temporary table
	_, err = db.Exec(confirmSimpleInsertSQL, idB64, key.Bytes())
	if err != nil {
		return err
	}

	confirmsFilePath, err := sm.getConfirmsFilePath()
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	// Write the confirmation key pair in the temporary table into the file
	if !os.IsNotExist(err) { // If confirmation keystore file exists
		// Check if confirmation key pair is already stored in the file
		sessionRow := db.QueryRow(fmt.Sprintf(findItemByIdSQL, confirmsFilePath), idB64)
		if sessionRow.Scan() != sql.ErrNoRows {
			return errors.New(ConfirmKeyExistsErrMsg)
		}

		// Combine the confirmation key within the file with the new session
		_, err = db.Exec(fmt.Sprintf(
			itemsAddToDbFileSQL, confirmsFilePath, confirmsTblName, confirmsFilePath+tempFileSuffix,
		))
		if err != nil {
			return err
		}

		err = sm.removeTempFile(confirmsFilePath)
		if err != nil {
			return err
		}
	} else { // Confirmation keystore file does not exist
		// Create a new file and write the confirmation key pair into it
		_, err = db.Exec(fmt.Sprintf(itemsWriteToDbFileSQL, confirmsTblName, confirmsFilePath))
		if err != nil {
			return err
		}
	}

	return nil
}

// RemoveConfirmKey attempts to remove the confirmation key with the given ID
// from the confirmation keystore
func (sm *Manager) RemoveConfirmKey(confirmId string, fileEncKey []byte) error {
	confirmsFilePath, err := sm.getConfirmsFilePath()
	if os.IsNotExist(err) {
		return errors.New(RemoveConfirmKeyNotStoredErrMsg)
	}
	if err != nil {
		return err
	}

	db, err := sm.OpenDb(fileEncKey)
	if err != nil {
		return err
	}
	defer db.Close()

	// Remove confirmation key
	_, err = db.Exec(fmt.Sprintf(removeItemSQL, confirmsFilePath, confirmsFilePath+tempFileSuffix), confirmId)
	if err != nil {
		return err
	}

	err = sm.removeTempFile(confirmsFilePath)
	if err != nil {
		return err
	}

	return nil
}

// PurgeConfirmKeystore attempts to delete the entire confirmation keystore. It
// returns the number of confirmation keys that were deleted.
func (sm *Manager) PurgeConfirmKeystore(fileEncKey []byte) (int, error) {
	// TODO: Complete this
	return 0, nil
}

/*******************************************************************************
 * Helpers
 ******************************************************************************/

// OpenDb is a helper function that opens the database connection for a database
// and sets the file encryption key that will be used to encrypt and decrypt
// database file(s). This function usually does not need to be called directly
// outside of testing. The file encryption key is only set within the database
// connection. The key is NOT checked if it is correct. Returns a database
// handle if there are no errors.
func (sm *Manager) OpenDb(fileEncKey []byte) (db *sql.DB, err error) {
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

// getSessionsFilePath returns the file path of the sessions database file. If
// the data directory does not exist, it tries to created it. It also checks if
// the file exists. If the file does not exist, an os.ErrNotExist is returned as
// the error.
func (sm *Manager) getSessionsFilePath() (filePath string, err error) {
	filePath = sm.dataDir + "/" + sm.sessionsFile

	// Ensure data directory exists
	err = os.Mkdir(sm.dataDir, dataDirPermissions)
	if err != nil && !os.IsExist(err) {
		return
	}

	// Check if session file exists
	_, err = os.Stat(filePath)
	if err != nil {
		return
	}

	return
}

// getConfirmsFilePath returns the file path of the confirmation keystore file.
// If the data directory does not exist, it tries to created it. It also checks
// if the file exists. If the file does not exist, an os.ErrNotExist is returned
// as the error.
func (sm *Manager) getConfirmsFilePath() (filePath string, err error) {
	filePath = sm.dataDir + "/" + sm.confirmsFile

	// Ensure data directory exists
	err = os.Mkdir(sm.dataDir, dataDirPermissions)
	if err != nil && !os.IsExist(err) {
		return
	}

	// Check if confirmation keystore file exists
	_, err = os.Stat(filePath)
	if err != nil {
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

// removeTempFile attempts to remove the temporary file used when modifying the
// file with the given file name.
func (sm *Manager) removeTempFile(originalFilename string) error {
	// Check if temporary file exists
	_, err := os.Stat(originalFilename + tempFileSuffix)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		// Could not access file some reason other than that it does not exist
		// (e.g. permissions, drive failure)
		return err
	}

	// Remove the original file
	err = os.Remove(originalFilename)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	// Rename the temporary file
	err = os.Rename(originalFilename+tempFileSuffix, originalFilename)
	if err != nil {
		return err
	}

	return nil
}

// generateConfirmationCode generates a confirmation using the confirmation code
// settings in the session manager
func (sm *Manager) generateConfirmationCode() (string, error) {
	var code string

	for range sm.confirmCodeCharset {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(sm.confirmCodeCharset))))
		if err != nil {
			// Return partial code if there is an error
			return code, err
		}
		code = code + string(DefaultConfirmCodeCharset[n.Uint64()])
	}

	return code, nil
}
