package session

import "time"

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

func (sc *SessionConfig) ToFile(fileEncKey []byte) error {
	// TODO: Put config JSON into a Paseto and write it to a file
	return nil
}

func ConfigFromFile() (*SessionConfig, error) {
	// TODO: Read config from file
	// TODO: Parse config Paseto
	// TODO: Parse config
	return nil, nil
}
