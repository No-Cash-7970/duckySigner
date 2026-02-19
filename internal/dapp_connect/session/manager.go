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
	"path/filepath"
	"strings"
	"time"

	"github.com/duckdb/duckdb-go/v2"

	dc "duckysigner/internal/dapp_connect"
	"duckysigner/internal/tools"
)

const (
	// sessionsTblName is the name of the in-memory table used to temporarily
	// store new sessions
	sessionsTblName = "sessions"
	// confirmsTblName is the name of the in-memory table used to temporarily
	// store new confirmation keys
	confirmsTblName = "confirms"
)

// attachEncDuckDbSQL is the SQL statement for opening or creating an encrypted
// DuckDB file. The file encryption key is used for encrypting and decrypting
// all the encrypted Duck DB files. Requires the file encryption key in Base64.
// NOTE: These SQL statements will create the file if it does not exist.
const attachEncDuckDbSQL = `
LOAD httpfs; -- use OpenSSL library to increase speed
ATTACH '%s' AS db (ENCRYPTION_KEY '%s');
`

// findItemByIdSQL is the SQL statement for finding an item (e.g. session)
// within a database file by ID. Requires the database table name.
const findItemByIdSQL = "FROM db.%s WHERE id = ?"

// getAllItemsSQL is the SQL statement for getting all stored items (eg. sessions,
// confirmation keys). Requires the database table name.
const getAllItemsSQL = "FROM db.%s"

// removeAllItemsSQL is the SQL statement for deleting all stored items (eg. sessions,
// confirmation keys). Requires the database table name.
const removeAllItemsSQL = "TRUNCATE db.%s"

// removeItemSQL is the SQL statement for removing an item (e.g. session,
// confirmation key) from a database file. Requires the database table name.
const removeItemSQL = "DELETE FROM db.%s WHERE id=?"

// countItemsSQL is the SQL statement for counting the number of items stored in
// a database file.
const countItemsSQL = "SELECT count(1) FROM db.%s"

// sessionsSchema is the SQL statement for creating the tables in the session
// data file
const sessionsSchema = `
CREATE TABLE IF NOT EXISTS db.sessions (
    id VARCHAR PRIMARY KEY,
    key BLOB NOT NULL,
    expiry TIMESTAMP_S NOT NULL,
    est TIMESTAMP_S NOT NULL,
    dapp_id VARCHAR NOT NULL,
    dapp_name VARCHAR,
    dapp_url VARCHAR,
    dapp_desc VARCHAR,
    dapp_icon VARCHAR,
    addrs VARCHAR[],
);

CREATE TABLE IF NOT EXISTS db.confirms (
    id VARCHAR PRIMARY KEY,
    key BLOB NOT NULL
);
`

// sessionInsertSQL is the SQL statement for inserting a session into a table
const sessionInsertSQL = "INSERT INTO db.sessions VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"

// confirmInsertSQL is the SQL statement for inserting a confirmation key pair
// into a table
const confirmInsertSQL = "INSERT INTO db.confirms VALUES (?, ?)"

// removeExpiredSessionsSQL is the SQL statement for removing all expired items
// (e.g. sessions, confirmation keys) from a database file.
const removeExpiredSessionsSQL = "DELETE FROM db.sessions WHERE expiry < ?"

/*******************************************************************************
 * Manager
 ******************************************************************************/

// Manager is the dApp connect session manager
type Manager struct {
	// The ECDH curve to use for generating a keys or processing stored keys
	curve tools.ECDHCurve
	// File name of the database file where the established sessions and pending
	// confirmations are stored
	dataFile string
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
	// Amount of time to wait for user approval of a session
	approvalTimeout time.Duration
}

// NewManager creates a new session manager using the given configuration for
// creating sessions. The given ECDH curve will be used by this new session
// manager to generate session keys. If the given session configuration is nil,
// the default configuration will be used.
func NewManager(curve tools.ECDHCurve, sessionConfig *SessionConfig) *Manager {
	var (
		sessionLife        time.Duration
		confirmLife        time.Duration
		dataDir            string
		dataFile           string
		confirmCodeCharset string
		confirmCodeLen     uint
		approvalTimeout    time.Duration
	)

	// Use default config if no config was given
	if sessionConfig == nil {
		sessionLife = DefaultSessionLifetime
		confirmLife = DefaultConfirmLifetime
		dataDir = DefaultDataDir
		dataFile = DefaultDataFile
		confirmCodeCharset = DefaultConfirmCodeCharset
		confirmCodeLen = DefaultConfirmCodeLen
		approvalTimeout = DefaultApprovalTimeout
	} else {
		sessionLife = time.Duration(sessionConfig.SessionLifetimeSecs) * time.Second
		confirmLife = time.Duration(sessionConfig.ConfirmLifetimeSecs) * time.Second
		dataDir = sessionConfig.DataDir
		dataFile = sessionConfig.DataFile
		confirmCodeCharset = sessionConfig.ConfirmCodeCharset
		confirmCodeLen = sessionConfig.ConfirmCodeLen
		approvalTimeout = time.Duration(sessionConfig.ApprovalTimeoutSecs) * time.Second

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

		// If no file name for the data file was given
		if dataFile == "" {
			dataFile = DefaultDataFile
		}

		// If no confirmation code character set was given
		if confirmCodeCharset == "" {
			confirmCodeCharset = DefaultConfirmCodeCharset
		}

		// If no confirmation code length was given
		if confirmCodeLen == 0 {
			confirmCodeLen = DefaultConfirmCodeLen
		}

		// If no approval timeout was given
		if approvalTimeout == time.Duration(0) {
			approvalTimeout = DefaultApprovalTimeout
		}
	}

	// The each part of the directory path must be escaped to prevent the
	// directory name from being used for SQL injection
	dataDirParts := strings.Split(filepath.FromSlash(dataDir), string(filepath.Separator))
	var escapedDataDirParts []string
	for _, part := range dataDirParts {
		escapedDataDirParts = append(escapedDataDirParts, url.PathEscape(part))
	}

	return &Manager{
		curve:              curve,
		sessionLifetime:    sessionLife,
		confirmLifetime:    confirmLife,
		dataDir:            filepath.Join(escapedDataDirParts...),
		dataFile:           url.PathEscape(dataFile),
		confirmCodeCharset: confirmCodeCharset,
		confirmCodeLen:     confirmCodeLen,
		approvalTimeout:    approvalTimeout,
	}
}

// DataFile returns the file name of the database file for storing
// established sessions
func (sm *Manager) DataFile() string {
	return sm.dataFile
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

// ApprovalTimeout returns the length of time to wait for the approval of a
// session
func (sm *Manager) ApprovalTimeout() time.Duration {
	return sm.approvalTimeout
}

/*******************************************************************************
 * Managing sessions
 ******************************************************************************/

// GenerateSession creates a new unestablished session (with no established-at
// date-time) by generating a new session key pair for the dApp with the given
// ID with the given dApp data
func (sm *Manager) GenerateSession(dappId *ecdh.PublicKey, dappData *dc.DappData, addrs []string) (session *Session, err error) {
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
		addrs:    addrs,
	}

	return
}

// GetSession attempts to retrieve the stored session with the given ID (in
// base64) using the given file encryption key to decrypt the session data file.
// Returns nil without an error if no session with the given ID is found.
func (sm *Manager) GetSession(sessionId string, fileEncKey []byte) (*Session, error) {
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
		retrievedDappIcon        string
		retrievedAddrs           duckdb.Composite[[]string]
	)
	sessionRow := db.QueryRow(fmt.Sprintf(findItemByIdSQL, sessionsTblName), sessionId)
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
		&retrievedAddrs,
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
		retrievedAddrs.Get(),
	)
}

// GetAllSessions attempts to retrieve all stored sessions using the given file
// encryption key to decrypt the session data file.
func (sm *Manager) GetAllSessions(fileEncKey []byte) (retrievedSessions []*Session, err error) {
	db, err := sm.OpenDb(fileEncKey)
	if err != nil {
		return
	}
	defer db.Close()

	sessionsRows, err := db.Query(fmt.Sprintf(getAllItemsSQL, sessionsTblName))
	if err != nil {
		return
	}

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
			retrievedDappIcon        string
			retrievedAddrs           duckdb.Composite[[]string]
		)

		err = sessionsRows.Scan(
			&retrievedSessionId,
			&retrievedSessionKeyBytes,
			&retrievedExp,
			&retrievedEst,
			&retrievedDappIdB64,
			&retrievedDappName,
			&retrievedDappURL,
			&retrievedDappDesc,
			&retrievedDappIcon,
			&retrievedAddrs,
		)
		if err != nil {
			// An unexpected error occurred
			// Return the incomplete set along with the error
			return
		}

		session, convertErr := sm.rowToSession(
			retrievedSessionKeyBytes,
			retrievedExp,
			retrievedEst,
			retrievedDappIdB64,
			retrievedDappName,
			retrievedDappURL,
			retrievedDappDesc,
			retrievedDappIcon,
			retrievedAddrs.Get(),
		)
		if convertErr != nil {
			// Return the incomplete set along with the error
			return
		}

		retrievedSessions = append(retrievedSessions, session)
	}

	return retrievedSessions, nil
}

// StoreSession attempts to store the given session using the given file
// encryption key to access the session data file
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

	// Insert session into table
	_, err = db.Exec(sessionInsertSQL,
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
		dappData.Icon,
		session.addrs,
	)
	if err != nil {
		// If the session already exists
		if strings.Contains(strings.ToLower(err.Error()), "constraint") {
			err = errors.New(SessionExistsErrMsg)
			return
		}

		// Unexpected error
		return
	}

	return
}

// RemoveSession attempts to remove the stored session with the given ID and the
// given file encryption key to access the session data file
func (sm *Manager) RemoveSession(sessionId string, fileEncKey []byte) error {
	db, err := sm.OpenDb(fileEncKey)
	if err != nil {
		return err
	}
	defer db.Close()

	// Remove session
	_, err = db.Exec(fmt.Sprintf(removeItemSQL, sessionsTblName), sessionId)
	if err != nil {
		return err
	}

	return nil
}

// PurgeAllSessions attempts to completely delete all stored sessions with the
// given ID and the given file encryption key to access the session data file.
// Returns the number of sessions that were deleted.
func (sm *Manager) PurgeAllSessions(fileEncKey []byte) (numPurged uint, err error) {
	db, err := sm.OpenDb(fileEncKey)
	if err != nil {
		return 0, err
	}
	defer db.Close()

	// Remove all sessions
	row := db.QueryRow(fmt.Sprintf(removeAllItemsSQL, sessionsTblName))
	err = row.Scan(&numPurged)

	return
}

// PurgeExpiredSessions attempts to delete all expired stored sessions with the
// given ID and the given file encryption key to access the session data file.
// Returns the number of sessions that were deleted.
func (sm *Manager) PurgeExpiredSessions(fileEncKey []byte) (numPurged uint, err error) {
	db, err := sm.OpenDb(fileEncKey)
	if err != nil {
		return 0, err
	}
	defer db.Close()

	// Remove all expired sessions
	row := db.QueryRow(removeExpiredSessionsSQL, time.Now().UTC())
	err = row.Scan(&numPurged)

	return
}

// EstablishSession creates a new established session using the given dApp data
// and connect addresses after checking the given confirmation token, code and
// key. NOTE: The established session is not saved into the data file. Use
// `StoreSession()` to save the session.
func (sm *Manager) EstablishSession(
	token string,
	code string,
	key *ecdh.PrivateKey,
	dappData *dc.DappData,
	connectAddrs []string,
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
		addrs:         connectAddrs,
	}

	return &session, nil
}

// EstablishSessionWithConfirm creates a new established session using the given
// dApp data and connect addresses after checking the given confirmation code
// (given by the wallet user) using the data within the given confirmation.
// NOTE: The established session is not saved into the data file. Use
// `StoreSession()` to save the session.
func (sm *Manager) EstablishSessionWithConfirm(
	confirm *Confirmation,
	codeFromUser string,
	dappData *dc.DappData,
	connectAddrs []string,
) (*Session, error) {
	// Check confirmation
	if confirm == nil {
		return nil, errors.New(NoConfirmGivenErrMsg)
	}

	// Check code
	if codeFromUser != confirm.code {
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
		addrs:         connectAddrs,
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
// ID (in base64) using the given file encryption key to decrypt the session
// data file. Returns nil without an error if no confirmation with the given ID
// is found.
func (sm *Manager) GetConfirmKey(confirmId string, fileEncKey []byte) (*ecdh.PrivateKey, error) {
	db, err := sm.OpenDb(fileEncKey)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	// Retrieve confirmation key from database
	var retrievedId string
	var retrievedKeyBytes []byte
	row := db.QueryRow(fmt.Sprintf(findItemByIdSQL, confirmsTblName), confirmId)
	err = row.Scan(&retrievedId, &retrievedKeyBytes)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		// An unexpected error occurred
		return nil, err
	}

	confirmKey, err := sm.curve.NewPrivateKey(retrievedKeyBytes)
	if err != nil {
		return nil, err
	}

	return confirmKey, nil
}

// GetAllConfirmKeys attempts to retrieve all stored confirmation keys using the
// given file encryption key to decrypt the session data file
func (sm *Manager) GetAllConfirmKeys(fileEncKey []byte) (retrievedKeys []*ecdh.PrivateKey, err error) {
	db, err := sm.OpenDb(fileEncKey)
	if err != nil {
		return
	}
	defer db.Close()

	rows, err := db.Query(fmt.Sprintf(getAllItemsSQL, confirmsTblName))
	if err != nil {
		return
	}

	// Convert each row into an ECDH private key
	for rows.Next() {
		var id string
		var keyBytes []byte

		err = rows.Scan(&id, &keyBytes)
		if err != nil {
			// An unexpected error occurred
			// Return the incomplete set along with the error
			return
		}

		key, newKeyErr := sm.curve.NewPrivateKey(keyBytes)
		if newKeyErr != nil {
			// Return the incomplete set along with the error
			return retrievedKeys, newKeyErr
		}

		retrievedKeys = append(retrievedKeys, key)
	}

	return retrievedKeys, nil
}

// StoreConfirmKey attempts to store the given confirmation key using the given
// file encryption key to access the session data file
func (sm *Manager) StoreConfirmKey(key *ecdh.PrivateKey, fileEncKey []byte) error {
	if key == nil {
		return errors.New(NoConfirmKeyGivenErrMsg)
	}

	db, err := sm.OpenDb(fileEncKey)
	if err != nil {
		return err
	}
	defer db.Close()

	idB64 := base64.StdEncoding.EncodeToString(key.PublicKey().Bytes())

	// Insert confirmation key pair into table
	_, err = db.Exec(confirmInsertSQL, idB64, key.Bytes())
	if err != nil {
		return err
	}

	return nil
}

// RemoveConfirmKey attempts to remove the confirmation key with the given ID
// from the session data file using the given file encryption key to access the
// session data file
func (sm *Manager) RemoveConfirmKey(confirmId string, fileEncKey []byte) error {
	db, err := sm.OpenDb(fileEncKey)
	if err != nil {
		return err
	}
	defer db.Close()

	// Remove confirmation
	_, err = db.Exec(fmt.Sprintf(removeItemSQL, confirmsTblName), confirmId)
	if err != nil {
		return err
	}

	return nil
}

// PurgeConfirmKeystore attempts to delete the entire confirmation keystore. It
// returns the number of confirmation keys that were deleted.
func (sm *Manager) PurgeConfirmKeystore(fileEncKey []byte) (numPurged uint, err error) {
	db, err := sm.OpenDb(fileEncKey)
	if err != nil {
		return 0, err
	}
	defer db.Close()

	// Remove all confirmations
	row := db.QueryRow(fmt.Sprintf(removeAllItemsSQL, confirmsTblName))
	err = row.Scan(&numPurged)

	return
}

/*******************************************************************************
 * Helpers
 ******************************************************************************/

// OpenDb is a helper function that opens the database connection for a database
// using the given file encryption key. This function usually does not need to
// be called directly outside of testing.
func (sm *Manager) OpenDb(fileEncKey []byte) (db *sql.DB, err error) {
	// TODO: Add mutex lock to protect against data races

	// Open DuckDB in in-memory mode
	db, err = sql.Open("duckdb", "")
	if err != nil {
		return
	}

	dataFilePath, err := sm.getDataFilePath()
	if err != nil {
		return
	}

	// Open and decrypt accounts database file
	_, err = db.Exec(fmt.Sprintf(attachEncDuckDbSQL,
		dataFilePath,
		base64.StdEncoding.EncodeToString(fileEncKey),
	))
	if err != nil {
		return
	}

	// Create tables if they do not exist
	_, err = db.Exec(sessionsSchema)
	if err != nil {
		return
	}

	return
}

// getDataFilePath returns the file path of the data file. If the data
// directory does not exist, it tries to created it. It DOES NOT check if the
// data file exists.
func (sm *Manager) getDataFilePath() (filePath string, err error) {
	filePath = filepath.Join(sm.dataDir, sm.dataFile)

	// Ensure data directory exists
	err = os.Mkdir(sm.dataDir, dataDirPermissions)
	if err != nil && !os.IsExist(err) {
		// Unexpected error
		return
	}

	return filePath, nil
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
	dappIcon string,
	addrs []string,
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
			Icon:        dappIcon,
		},
		addrs: addrs,
	}, nil
}

// generateConfirmationCode generates a confirmation using the confirmation code
// settings in the session manager
func (sm *Manager) generateConfirmationCode() (string, error) {
	var code []byte

	for i := 0; i < int(sm.confirmCodeLen); i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(sm.confirmCodeCharset))))
		if err != nil {
			// Return partial code if there is an error
			return string(code), err
		}
		code = append(code, sm.confirmCodeCharset[n.Uint64()])
	}

	return string(code), nil
}
