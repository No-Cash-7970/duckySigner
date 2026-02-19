package session

import (
	"encoding/json"
	"os"
	"time"

	"aidanwoods.dev/go-paseto"
)

const (
	// DefaultConfigFile is the default file name for the session configuration
	// file
	DefaultConfigFile = "session_config.paseto"
	// DefaultDataFile is the default file name for the database file where
	// established sessions and pending confirmations are stored
	DefaultDataFile = "sessions.duckdb"
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
	// DefaultConfirmCodeLen is the default length of a confirmation code
	DefaultConfirmCodeLen = 5
	// DefaultApprovalTimeout is the default length of time to wait for a user
	// to approve a dApp connect session. This must be less than the default
	// confirmation lifetime.
	DefaultApprovalTimeout = 5 * time.Minute
)

// SessionConfig is used to configure the session manager when creating a new
// session manager
type SessionConfig struct {
	// File name of the database file where the established sessions and pending
	// confirmations are stored
	DataFile string `json:"sessions_file,omitempty"`
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
	// The length of time to wait for a user to approve a dApp connect session.
	// This must be less than the confirmation lifetime.
	ApprovalTimeoutSecs uint64 `json:"approval_timeout_secs,omitempty"`

	// TODO: Create mutex lock to protect config from races
}

// ToFile creates the session configuration file by encapsulating the session
// configuration into a Paseto using the given encryption key and writes it to a
// file with the given file path. If the file exists, it is overwritten.
func (sc *SessionConfig) ToFile(filePath string, encryptKey []byte) error {
	// Convert session configuration to JSON
	sc.fillWithDefaults()
	configJsonBytes, err := json.Marshal(sc)
	if err != nil {
		return err
	}

	// Encapsulate settings into a local Paseto. This is primarily for
	// preventing tampering of the configuration file.
	configPaseto := paseto.NewToken()
	// Note: Placing the configuration data in the footer means the data is
	// Base64-encoded and authenticated, but NOT encrypted
	configPaseto.SetFooter(configJsonBytes)
	// Use the encryption key to create the token
	pasetoKey, err := paseto.V4SymmetricKeyFromBytes(encryptKey)
	if err != nil {
		return err
	}
	// Create token
	configPasetoString := configPaseto.V4Encrypt(pasetoKey, nil)

	// Write token to a file
	os.WriteFile(filePath+tempFileSuffix, []byte(configPasetoString), dataDirPermissions)
	err = removeTempFile(filePath)
	if err != nil {
		return err
	}

	return nil
}

// ConfigFromFile extracts the session configuration from the file with the
// given file path using the given encryption key
func ConfigFromFile(configFilePath string, encryptKey []byte) (config *SessionConfig, err error) {
	// Extract file contents
	fileBytes, err := os.ReadFile(configFilePath)
	if err != nil {
		return
	}

	// Decrypt, parse and validate the Paseto that was in the file
	pasetoKey, err := paseto.V4SymmetricKeyFromBytes(encryptKey)
	if err != nil {
		return
	}
	parser := paseto.NewParserWithoutExpiryCheck()
	// Note: The token is validated when it is parsed
	token, err := parser.ParseV4Local(pasetoKey, string(fileBytes), nil)
	if err != nil {
		return
	}

	// Parse the session configuration JSON within the token footer
	config = &SessionConfig{}
	json.Unmarshal(token.Footer(), config)

	return
}

// fillWithDefaults fills in all unset configuration fields with their
// respective default values
func (sc *SessionConfig) fillWithDefaults() {
	if sc.DataFile == "" {
		sc.DataFile = DefaultDataFile
	}

	if sc.DataDir == "" {
		sc.DataDir = DefaultDataDir
	}

	if sc.SessionLifetimeSecs == 0 {
		sc.SessionLifetimeSecs = uint64(DefaultSessionLifetime.Seconds())
	}

	if sc.ConfirmLifetimeSecs == 0 {
		sc.ConfirmLifetimeSecs = uint64(DefaultConfirmLifetime.Seconds())
	}

	if sc.ConfirmCodeCharset == "" {
		sc.ConfirmCodeCharset = DefaultConfirmCodeCharset
	}

	if sc.ConfirmCodeLen == 0 {
		sc.ConfirmCodeLen = DefaultConfirmCodeLen
	}
}
