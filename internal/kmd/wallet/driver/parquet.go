// XXX: This driver is a modified version of a modified version of the sqlite driver in go-algorand. This is modified to use DuckDB instead of SQLite.
// Modified SQLite driver: https://github.com/No-Cash-7970/duckySigner/blob/0f5365f1062a36c20af88379dd61d53693314b11/internal/kmd/wallet/driver/sqlite.go
// Original SQLite driver: https://github.com/algorand/go-algorand/tree/c2d7047585f6109d866ebaf9fca0ee7490b16c6a/daemon/kmd/wallet/driver/sqlite.go

package driver

import (
	"crypto/ed25519"
	"crypto/subtle"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"

	"duckysigner/internal/kmd/config"
	kmdCrypto "duckysigner/internal/kmd/crypto"
	"duckysigner/internal/kmd/wallet"

	"github.com/algorand/go-algorand-sdk/v2/crypto"
	"github.com/algorand/go-algorand-sdk/v2/types"
	"github.com/algorand/go-codec/codec"
	"github.com/algorand/go-deadlock"
	"github.com/awnumar/memguard"
	"github.com/jmoiron/sqlx"
	_ "github.com/marcboeker/go-duckdb/v2"
	logging "github.com/sirupsen/logrus"
	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

const (
	parquetWalletDriverName      = "parquet"
	parquetWalletDriverVersion   = 1
	parquetWalletsDirName        = "parquet_wallets"
	parquetWalletsDirPermissions = 0700
	parquetWalletDBOptions       = "_pragma=secure_delete(1)&_txlock=exclusive"
	parquetMaxWalletNameLen      = 64
	parquetMaxWalletIDLen        = 64
	parquetIntOverflow           = 1 << 63
	parquetWalletHasMnemonicUX   = false
	parquetWalletHasMasterKey    = true
	// ParquetMetadatasFile is the file name of the file that is supposed to
	// contain a list of the metadatas of all wallets
	ParquetMetadatasFile = "metadatas.parquet"
	// ParquetWalletMetadataFile is the name of the metadata file is in each
	// directory for a wallet
	ParquetWalletMetadataFile = "metadata.json"

	// tempFileSuffix is the suffix used to create a temporary file. The
	// temporary file name is the original file name + this suffix.
	tempFileSuffix = ".new"
)

var parquetWalletSupportedTxs = []types.TxType{
	types.PaymentTx,
	types.KeyRegistrationTx,
	types.ApplicationCallTx,
	types.AssetConfigTx,
	types.AssetFreezeTx,
	types.AssetTransferTx,
}
var disallowedParquetFilenameRegex = regexp.MustCompile("[^a-zA-Z0-9_-]*")
var parquetFilenameRegex = regexp.MustCompile(`^.*\.parquet$`)

var parquetCreateMetadatasTblSchema = `
CREATE TABLE IF NOT EXISTS metadatas (
	driver_name TEXT NOT NULL,
	driver_version INT NOT NULL,
	wallet_id TEXT NOT NULL UNIQUE,
	wallet_name TEXT NOT NULL,
	mep_encrypted BLOB NOT NULL,
	mdk_encrypted BLOB NOT NULL,
	max_key_idx_encrypted BLOB NOT NULL
);
`

var parquetWalletSchema = `
CREATE TABLE IF NOT EXISTS metadata (
	driver_name TEXT NOT NULL,
	driver_version INT NOT NULL,
	wallet_id TEXT NOT NULL UNIQUE,
	wallet_name TEXT NOT NULL,
	mep_encrypted BLOB NOT NULL,
	mdk_encrypted BLOB NOT NULL,
	max_key_idx_encrypted BLOB NOT NULL
);

CREATE TABLE IF NOT EXISTS keys (
	address BLOB PRIMARY KEY,
	secret_key_encrypted BLOB NOT NULL,
	key_idx INT
);

CREATE TABLE IF NOT EXISTS msig_addrs (
	address BLOB PRIMARY KEY,
	version INT NOT NULL,
	threshold INT NOT NULL,
	pks BLOB NOT NULL
);
`

// ParquetWalletDriver is the default wallet driver used by kmd. Keys are stored
// as authenticated-encrypted blobs in a sqlite 3 database.
type ParquetWalletDriver struct {
	globalCfg  config.KMDConfig
	parquetCfg config.ParquetWalletDriverConfig
	mux        *deadlock.Mutex
}

// ParquetWallet represents a particular ParquetWallet under the
// ParquetWalletDriver
type ParquetWallet struct {
	masterEncryptionKey  *memguard.Enclave
	masterDerivationKey  *memguard.Enclave
	walletPasswordSalt   [saltLen]byte
	walletPasswordHash   types.Digest
	walletPasswordHashed bool
	dbPath               string
	cfg                  config.ParquetWalletDriverConfig
	id                   string
	walletsPath          string
}

type ParquetWalletMetadata struct {
	DriverName         string `json:"driver_name"`
	DriverVersion      int    `json:"driver_version"`
	WalletId           string `json:"wallet_id"`
	WalletName         string `json:"wallet_name"`
	MEPEncrypted       []byte `json:"mep_encrypted"`
	MDKEncrypted       []byte `json:"mdk_encrypted"`
	MaxKeyIdxEncrypted []byte `json:"max_key_idx_encrypted"`
}

/*******************************************************************************
 * Wallet Driver
 ******************************************************************************/

// InitWithConfig accepts a driver configuration so that the Parquet driver
// knows where to read and write its wallet databases
func (parqwd *ParquetWalletDriver) InitWithConfig(cfg config.KMDConfig, log *logging.Logger) error {
	parqwd.globalCfg = cfg
	parqwd.parquetCfg = cfg.DriverConfig.ParquetWalletDriverConfig

	// Make sure the scrypt params are reasonable
	if !parqwd.parquetCfg.UnsafeScrypt {
		if parqwd.parquetCfg.ScryptParams.ScryptN < minScryptN {
			return fmt.Errorf("slow scrypt N must be at least %d", minScryptN)
		}
		if parqwd.parquetCfg.ScryptParams.ScryptR < minScryptR {
			return fmt.Errorf("slow scrypt R must be at least %d", minScryptR)
		}
		if parqwd.parquetCfg.ScryptParams.ScryptP < minScryptP {
			return fmt.Errorf("slow scrypt P must be at least %d", minScryptP)
		}
	}

	// Make the wallets directory if it doesn't already exist
	err := parqwd.maybeMakeWalletsDir()
	if err != nil {
		return err
	}

	// Initialize lock. When creating a new wallet, this lock protects us from
	// creating another with the same name or ID
	parqwd.mux = &deadlock.Mutex{}

	return nil
}

// ListWalletMetadata returns the metadatas stored in the generated "metadatas"
// parquet file. If the metadatas parquet file does not exist, the file is
// generated by extracting the metadata of each wallet in the wallets directory.
func (parqwd *ParquetWalletDriver) ListWalletMetadatas() ([]wallet.Metadata, error) {
	// Do not list if this wallet driver is disabled
	if parqwd.parquetCfg.Disable {
		return []wallet.Metadata{}, nil
	}

	metadatasPath := parqwd.walletsDir() + "/" + ParquetMetadatasFile

	// Check if metadatas file exists
	_, statErr := os.Stat(metadatasPath)

	// Open database
	db, err := sqlx.Connect("duckdb", "")
	if err != nil {
		return []wallet.Metadata{}, err
	}
	defer db.Close()

	metadatas := []wallet.Metadata{}

	if statErr == nil { // The metadatas file exists
		// Return all metadatas in the metadatas file
		rows, err := db.Query(fmt.Sprintf("FROM read_parquet('%s')", metadatasPath))
		if err != nil {
			return metadatas, err
		}

		for rows.Next() {
			retrievedMetadata := ParquetWalletMetadata{}
			err := rows.Scan(
				&retrievedMetadata.DriverName,
				&retrievedMetadata.DriverVersion,
				&retrievedMetadata.WalletId,
				&retrievedMetadata.WalletName,
				&retrievedMetadata.MEPEncrypted,
				&retrievedMetadata.MDKEncrypted,
				&retrievedMetadata.MaxKeyIdxEncrypted,
			)
			if err != nil {
				return metadatas, nil
			}
			metadatas = append(metadatas, wallet.Metadata{
				ID:                    []byte(retrievedMetadata.WalletId),
				Name:                  []byte(retrievedMetadata.WalletName),
				DriverName:            retrievedMetadata.DriverName,
				DriverVersion:         uint32(retrievedMetadata.DriverVersion),
				SupportsMnemonicUX:    parquetWalletHasMnemonicUX,
				SupportsMasterKey:     parquetWalletHasMasterKey,
				SupportedTransactions: parquetWalletSupportedTxs,
			})
		}

		return metadatas, nil
	}

	if os.IsNotExist(statErr) { // The metadatas file does not exists
		// Get a list of the paths that may be wallets
		paths, err := parqwd.potentialWalletPaths()
		if err != nil {
			return metadatas, err
		}

		// Run the schema for creating the metadatas schema in temporary database
		_, err = db.Exec(parquetCreateMetadatasTblSchema)
		if err != nil {
			return metadatas, err
		}

		if len(paths) == 0 {
			return metadatas, err
		}

		for _, path := range paths {
			// Get metadata from path (if possible)
			walletMetadata, err := parquetWalletMetadataFromPath(path)
			if err != nil {
				continue
			}

			// Add metadata to temporary metadatas table
			_, err = db.Exec(
				"INSERT INTO metadatas (driver_name, driver_version, wallet_id, wallet_name, mep_encrypted, mdk_encrypted, max_key_idx_encrypted) VALUES(?, ?, ?, ?, ?, ?, ?)",
				parquetWalletDriverName,
				parquetWalletDriverVersion,
				walletMetadata.WalletId,
				walletMetadata.WalletName,
				walletMetadata.MEPEncrypted,
				walletMetadata.MDKEncrypted,
				walletMetadata.MaxKeyIdxEncrypted,
			)
			if err != nil {
				return metadatas, err
			}

			metadatas = append(metadatas, wallet.Metadata{
				ID:                    []byte(walletMetadata.WalletId),
				Name:                  []byte(walletMetadata.WalletName),
				DriverName:            walletMetadata.DriverName,
				DriverVersion:         uint32(walletMetadata.DriverVersion),
				SupportsMnemonicUX:    parquetWalletHasMnemonicUX,
				SupportsMasterKey:     parquetWalletHasMasterKey,
				SupportedTransactions: parquetWalletSupportedTxs,
			})
		}

		// Create write temporary metadatas table to file
		_, err = db.Exec(fmt.Sprintf("COPY metadatas TO '%s' (FORMAT parquet)", metadatasPath))
		if err != nil {
			return metadatas, err
		}

		return metadatas, nil
	}

	// If we are at this point, something went wrong with checking the metadatas file
	return metadatas, statErr
}

// CreateWallet creates a wallet with the given name and ID that will be
// protected with the given password and Master Derivation Key (MDK). Providing
// the MDK is optional. If an MDK is provided, then one will be generated.
func (parqwd *ParquetWalletDriver) CreateWallet(name []byte, id []byte, pw []byte, mdk types.MasterDerivationKey) error {
	if len(name) > parquetMaxWalletNameLen {
		return errNameTooLong
	}

	if len(id) > parquetMaxWalletIDLen {
		return errIDTooLong
	}

	walletPath := parqwd.idToPath(id)

	// Create directory for new wallet
	err := os.Mkdir(walletPath, parquetWalletsDirPermissions)
	if err != nil {

		return err
	}

	// Generate the master encryption password, used to encrypt the master
	// derivation key, generated keys, and imported keys
	var masterKey [masterKeyLen]byte
	err = fillRandomBytes(masterKey[:])
	if err != nil {
		return err
	}

	// If we were passed a blank master derivation key, generate one here
	masterDerivationKey := mdk
	if masterDerivationKey == (types.MasterDerivationKey{}) {
		err = fillRandomBytes(masterDerivationKey[:])
		if err != nil {
			return err
		}
	}

	// Encrypt the master encryption password using the user's password (which
	// may be blank)
	encryptedMEPBlob, err := encryptBlobWithPasswordBlankOK(masterKey[:], PTMasterKey, pw, &parqwd.parquetCfg.ScryptParams)
	if err != nil {
		return err
	}

	// Encrypt the master derivation key using the master encryption password
	// (which may not be blank)
	encryptedMDKBlob, err := encryptBlobWithKey(masterDerivationKey[:], PTMasterDerivationKey, masterKey[:])
	if err != nil {
		return err
	}

	// Encrypt the max key index using the master encryption password. We encrypt
	// this for integrity reasons, so that someone with access to the file can't
	// make the index enormous.
	maxKeyIdx := 0
	encryptedIdxBlob, err := encryptBlobWithKey(msgpackEncode(maxKeyIdx), PTMaxKeyIdx, masterKey[:])
	if err != nil {
		return err
	}

	metadata := ParquetWalletMetadata{
		DriverName:         parquetWalletDriverName,
		DriverVersion:      parquetWalletDriverVersion,
		WalletId:           string(id),
		WalletName:         string(name),
		MEPEncrypted:       encryptedMEPBlob,
		MDKEncrypted:       encryptedMDKBlob,
		MaxKeyIdxEncrypted: encryptedIdxBlob,
	}

	// Add metadata of new wallet metadatas file (if it exists)
	parqwd.addParquetWalletMetadata(&metadata)

	// Create metadata.json file in new wallet directory
	metadataJson, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	err = os.WriteFile(walletPath+"/"+ParquetWalletMetadataFile, metadataJson, parquetWalletsDirPermissions)
	if err != nil {
		return err
	}

	return nil
}

// FetchWallet looks up a wallet by ID and returns it. The wallet returned by
// this function in uninitialized and will need to be initialized before using
// most of the wallet function.
func (parqwd *ParquetWalletDriver) FetchWallet(id []byte) (wallet.Wallet, error) {
	if len(id) == 0 {
		return &ParquetWallet{}, fmt.Errorf("no ID is given")
	}

	metadatasPath := parqwd.walletsDir() + "/" + ParquetMetadatasFile

	// Check if metadatas file exists
	_, err := os.Stat(metadatasPath)
	if err != nil {
		return &ParquetWallet{}, errWalletNotFound
	}

	// Open database
	db, err := sqlx.Connect("duckdb", "")
	if err != nil {
		return &ParquetWallet{}, err
	}
	defer db.Close()

	// Get wallet metadata from database
	row := db.QueryRow(
		fmt.Sprintf("FROM read_parquet('%s') where wallet_id = ?", metadatasPath),
		id,
	)
	retrievedMetadata := ParquetWalletMetadata{}
	err = row.Scan(
		&retrievedMetadata.DriverName,
		&retrievedMetadata.DriverVersion,
		&retrievedMetadata.WalletId,
		&retrievedMetadata.WalletName,
		&retrievedMetadata.MEPEncrypted,
		&retrievedMetadata.MDKEncrypted,
		&retrievedMetadata.MaxKeyIdxEncrypted,
	)
	if err == sql.ErrNoRows {
		return &ParquetWallet{}, errWalletNotFound
	}
	if err != nil {
		// Some unexpected error occurred
		return &ParquetWallet{}, err
	}

	// Fill in the wallet details
	return &ParquetWallet{
		id:          string(id),
		walletsPath: parqwd.walletsDir(),
		cfg:         parqwd.parquetCfg,
	}, nil
}

// RenameWallet renames the wallet with the given id to newName. The given
// password is ignored, so the wallet can be successfully renamed if password is
// incorrect. The password can be left empty.
func (parqwd *ParquetWalletDriver) RenameWallet(newName []byte, id []byte, pw []byte) error {
	if len(id) == 0 {
		return fmt.Errorf("no ID is given")
	}

	if len(newName) > parquetMaxWalletNameLen {
		return errNameTooLong
	}

	walletMetadataPath := parqwd.walletsDir() + "/" + string(id) + "/" + ParquetWalletMetadataFile

	// Load wallet's metadata file
	metadataFileContents, err := os.ReadFile(walletMetadataPath)
	if os.IsNotExist(err) {
		// The directory is not a valid wallet
		return errWalletNotFound
	}
	if err != nil {
		// An unexpected error occurred
		return err
	}
	metadata := ParquetWalletMetadata{}
	err = json.Unmarshal(metadataFileContents, &metadata)
	if err != nil {
		return err
	}

	// Update the name within the metadata file
	metadata.WalletName = string(newName)
	updatedMetadataJson, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	err = os.WriteFile(walletMetadataPath+tempFileSuffix, updatedMetadataJson, parquetWalletsDirPermissions)
	if err != nil {
		return err
	}
	err = removeTempFile(walletMetadataPath)
	if err != nil {
		return err
	}

	metadatasPath := parqwd.walletsDir() + "/" + ParquetMetadatasFile

	// Check if metadatas file exists
	_, err = os.Stat(metadatasPath)
	if err != nil {
		return errWalletNotFound
	}

	// Open database
	db, err := sqlx.Connect("duckdb", "")
	if err != nil {
		return err
	}
	defer db.Close()

	// Run the schema for creating the metadatas schema in temporary in-memory database
	_, err = db.Exec(parquetCreateMetadatasTblSchema)
	if err != nil {
		return err
	}

	// Load metadatas file into temporary in-memory table
	_, err = db.Exec(fmt.Sprintf("INSERT INTO metadatas FROM read_parquet('%s')", metadatasPath))
	if err != nil {
		return err
	}

	// Update wallet name
	_, err = db.Exec("UPDATE metadatas SET wallet_name=? WHERE wallet_id=?", newName, id)
	if err != nil {
		return err
	}

	// Replace metadatas file
	_, err = db.Exec(fmt.Sprintf("COPY metadatas TO '%s' (FORMAT parquet)", metadatasPath+tempFileSuffix))
	if err != nil {
		return err
	}
	err = removeTempFile(metadatasPath)
	if err != nil {
		return err
	}

	return nil
}

/*******************************************************************************
 * Wallet
 ******************************************************************************/

// Metadata builds a wallet.Metadata from the wallet's metadata in the metadatas
// file
func (pqw *ParquetWallet) Metadata() (wallet.Metadata, error) {
	metadatasPath := pqw.walletsPath + "/" + ParquetMetadatasFile

	// Check if metadatas file exists
	_, err := os.Stat(metadatasPath)
	if err != nil {
		return wallet.Metadata{}, errWalletNotFound
	}

	// Open database
	db, err := sqlx.Connect("duckdb", "")
	if err != nil {
		return wallet.Metadata{}, err
	}
	defer db.Close()

	// Get wallet metadata from database
	row := db.QueryRow(
		fmt.Sprintf("FROM read_parquet('%s') where wallet_id = ?", metadatasPath),
		pqw.id,
	)
	retrievedMetadata := ParquetWalletMetadata{}
	err = row.Scan(
		&retrievedMetadata.DriverName,
		&retrievedMetadata.DriverVersion,
		&retrievedMetadata.WalletId,
		&retrievedMetadata.WalletName,
		&retrievedMetadata.MEPEncrypted,
		&retrievedMetadata.MDKEncrypted,
		&retrievedMetadata.MaxKeyIdxEncrypted,
	)
	if err == sql.ErrNoRows {
		return wallet.Metadata{}, errWalletNotFound
	}
	if err != nil {
		// Some unexpected error occurred
		return wallet.Metadata{}, err
	}

	return wallet.Metadata{
		ID:                    []byte(retrievedMetadata.WalletId),
		Name:                  []byte(retrievedMetadata.WalletName),
		DriverName:            retrievedMetadata.DriverName,
		DriverVersion:         uint32(retrievedMetadata.DriverVersion),
		SupportsMnemonicUX:    parquetWalletHasMnemonicUX,
		SupportsMasterKey:     parquetWalletHasMasterKey,
		SupportedTransactions: parquetWalletSupportedTxs,
	}, nil
}

// Init attempts to decrypt the master encrypt password and master derivation
// key, and store them in memory for subsequent operations
func (pqw *ParquetWallet) Init(pw []byte) error {
	// Decrypt the master password
	masterEncryptionKey, err := pqw.decryptAndGetMasterKey(pw)
	if err != nil {
		return err
	}

	// Decrypt the master derivation key
	masterDerivationKey, err := pqw.decryptAndGetMasterDerivationKey(masterEncryptionKey)
	if err != nil {
		return err
	}

	// Initialize wallet
	pqw.masterEncryptionKey = memguard.NewEnclave(masterEncryptionKey)
	pqw.masterDerivationKey = memguard.NewEnclave(masterDerivationKey)
	err = fillRandomBytes(pqw.walletPasswordSalt[:])
	if err != nil {
		return err
	}
	pqw.walletPasswordHash = fastHashWithSalt(pw, pqw.walletPasswordSalt[:])
	pqw.walletPasswordHashed = true
	return nil
}

// CheckPassword checks that the database can be decrypted with the password.
// It's the same as Init but doesn't store the decrypted key
func (pqw *ParquetWallet) CheckPassword(pw []byte) error {
	if pqw.walletPasswordHashed {
		// Check against pre-computed password hash
		pwhash := fastHashWithSalt(pw, pqw.walletPasswordSalt[:])
		if subtle.ConstantTimeCompare(pwhash[:], pqw.walletPasswordHash[:]) == 1 {
			return nil
		}
		return errDecrypt
	}

	_, err := pqw.decryptAndGetMasterKey(pw)
	return err
}

// ListKeys lists all the addresses in the wallet
func (pqw *ParquetWallet) ListKeys() (addrs []types.Digest, err error) {
	// Connect to the database
	db, err := sqlx.Connect("sqlite", parquetDbConnectionURL(pqw.dbPath))
	if err != nil {
		return nil, errDatabaseConnect
	}
	defer db.Close()

	var addrByteSlices [][]byte
	var tmp types.Digest
	// We can't select directly into a types.Digest array, unfortunately.
	// Instead, we select into a slice of byte slices, and then convert each of
	// those slices into a types.Digest.
	err = db.Select(&addrByteSlices, "SELECT address FROM keys")
	if err != nil {
		err = errDatabase
		return
	}

	for _, byteSlice := range addrByteSlices {
		copy(tmp[:], byteSlice)
		addrs = append(addrs, tmp)
	}

	return
}

// ExportMasterDerivationKey decrypts the encrypted MDK and returns it
func (pqw *ParquetWallet) ExportMasterDerivationKey(pw []byte) (mdk types.MasterDerivationKey, err error) {
	// Check the password
	err = pqw.CheckPassword(pw)
	if err != nil {
		return
	}

	// Decrypt the master derivation key stored in enclave into a local copy
	mdkBuf, err := pqw.masterDerivationKey.Open()
	if err != nil {
		return
	}
	defer mdkBuf.Destroy() // Destroy the copy when we return

	// Copy master derivation key into the result
	copy(mdk[:], mdkBuf.Bytes())
	return
}

// ImportKey imports a keypair into the wallet, deriving the public key from
// the passed secret key
func (pqw *ParquetWallet) ImportKey(rawSK ed25519.PrivateKey) (addr types.Digest, err error) {
	// Extract the seed from the secret key so that we don't trust the public part
	seed := rawSK.Seed()

	// Convert the seed to an sk/pk pair
	sk := ed25519.NewKeyFromSeed(seed[:])
	pk := sk.Public().(ed25519.PublicKey)

	// Decrypt the master encryption key stored in enclave into a local copy
	mekBuf, err := pqw.masterEncryptionKey.Open()
	if err != nil {
		return
	}
	defer mekBuf.Destroy() // Destroy the copy when we return

	// Encrypt the encoded secret key
	skEncrypted, err := encryptBlobWithKey(msgpackEncode(sk), PTSecretKey, mekBuf.Bytes())
	if err != nil {
		return
	}

	// Connect to the database
	db, err := sqlx.Connect("sqlite", parquetDbConnectionURL(pqw.dbPath))
	if err != nil {
		err = errDatabaseConnect
		return
	}
	defer db.Close()

	// Insert the pk, e(sk) into the database
	addr = publicKeyToAddress(pk)
	_, err = db.Exec("INSERT INTO keys (address, secret_key_encrypted) VALUES(?, ?)", addr[:], skEncrypted)
	err = checkParquetDBError(err)
	if err != nil {
		return
	}
	return
}

// ExportKey fetches the encrypted private key using the public key, decrypts
// it, verifies that it matches the passed public key, and returns it
func (pqw *ParquetWallet) ExportKey(addr types.Digest, pw []byte) (sk ed25519.PrivateKey, err error) {
	// Check the password
	err = pqw.CheckPassword(pw)
	if err != nil {
		return
	}

	// Export the key
	return pqw.fetchSecretKey(addr)
}

// GenerateKey generates a key from system entropy and imports it
func (pqw *ParquetWallet) GenerateKey(displayMnemonic bool) (addr types.Digest, err error) {
	// Connect to the database
	db, err := sqlx.Connect("sqlite", parquetDbConnectionURL(pqw.dbPath))
	if err != nil {
		err = errDatabaseConnect
		return
	}
	defer db.Close()

	// The sqlite wallet has SupportsMnemonicUX = false, meaning we don't know how to
	// show mnemonics to the user
	if displayMnemonic {
		err = errNoMnemonicUX
		return
	}

	// Begin an exclusive database transaction (we set _txlock=exclusive on the
	// database connection string)
	tx, err := db.Beginx()
	if err != nil {
		err = errDatabase
		return
	}

	// Generate and insert the next key
	addr, err = pqw.generateKeyTxLocked(tx)
	if err != nil {
		// Rollback in case any part of the tx failed
		tx.Rollback()
		return
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		err = errDatabase
		return
	}

	return addr, nil
}

// DeleteKey deletes the key corresponding to the passed public key from the wallet
func (pqw *ParquetWallet) DeleteKey(addr types.Digest, pw []byte) (err error) {
	// Check the password
	err = pqw.CheckPassword(pw)
	if err != nil {
		return
	}

	// Connect to the database
	db, err := sqlx.Connect("sqlite", parquetDbConnectionURL(pqw.dbPath))
	if err != nil {
		return errDatabaseConnect
	}
	defer db.Close()

	// Delete the key
	_, err = db.Exec("DELETE FROM keys WHERE address=?", addr[:])
	if err != nil {
		err = errDatabase
	}
	return
}

// ImportMultisigAddr imports a multisig address, taking in version, threshold,
// and public keys
func (pqw *ParquetWallet) ImportMultisigAddr(version, threshold uint8, pks []ed25519.PublicKey) (addr types.Digest, err error) {
	addr, err = kmdCrypto.MultisigAddrGen(version, threshold, pks)
	if err != nil {
		return
	}

	// Connect to the database
	db, err := sqlx.Connect("sqlite", parquetDbConnectionURL(pqw.dbPath))
	if err != nil {
		err = errDatabaseConnect
		return
	}
	defer db.Close()

	_, err = db.Exec("INSERT INTO msig_addrs (address, version, threshold, pks) VALUES (?, ?, ?, ?)", addr[:], version, threshold, msgpackEncode(pks))
	err = checkParquetDBError(err)
	if err != nil {
		return
	}
	return
}

// LookupMultisigPreimage exports the preimage of a multisig address: version,
// threshold, public keys
func (pqw *ParquetWallet) LookupMultisigPreimage(addr types.Digest) (version, threshold uint8, pks []ed25519.PublicKey, err error) {
	// Connect to the database
	db, err := sqlx.Connect("sqlite", parquetDbConnectionURL(pqw.dbPath))
	if err != nil {
		err = errDatabaseConnect
		return
	}
	defer db.Close()

	var pksCandidate []ed25519.PublicKey
	var versionCandidate, thresholdCandidate int
	var pksBlob []byte

	row := db.QueryRow("SELECT version, threshold, pks FROM msig_addrs WHERE address=?", addr[:])
	err = row.Scan(&versionCandidate, &thresholdCandidate, &pksBlob)
	if err != nil {
		err = errMsigDataNotFound
		return
	}

	// Decode the candidate
	err = msgpackDecode(pksBlob, &pksCandidate)
	if err != nil {
		return
	}

	// Sanity check: make sure the preimage is correct
	addr2, err := kmdCrypto.MultisigAddrGen(uint8(versionCandidate), uint8(thresholdCandidate), pksCandidate)
	if addr2 != addr {
		err = errTampering
		return
	}

	version = uint8(versionCandidate)
	threshold = uint8(thresholdCandidate)
	pks = pksCandidate
	return
}

// DeleteMultisigAddr deletes the multisig address and preimage from the database
func (pqw *ParquetWallet) DeleteMultisigAddr(addr types.Digest, pw []byte) (err error) {
	// Check the password
	err = pqw.CheckPassword(pw)
	if err != nil {
		return
	}

	// Connect to the database
	db, err := sqlx.Connect("sqlite", parquetDbConnectionURL(pqw.dbPath))
	if err != nil {
		return errDatabaseConnect
	}
	defer db.Close()

	_, err = db.Exec("DELETE FROM msig_addrs WHERE address=?", addr[:])
	if err != nil {
		err = errDatabase
	}
	return
}

// ListMultisigAddrs lists the multisig addresses whose preimages we know
func (pqw *ParquetWallet) ListMultisigAddrs() (addrs []types.Digest, err error) {
	// Connect to the database
	db, err := sqlx.Connect("sqlite", parquetDbConnectionURL(pqw.dbPath))
	if err != nil {
		err = errDatabaseConnect
		return
	}
	defer db.Close()

	var addrByteSlices [][]byte
	var tmp types.Digest
	err = db.Select(&addrByteSlices, "SELECT address FROM msig_addrs")
	if err != nil {
		err = errDatabase
		return
	}
	for _, addr := range addrByteSlices {
		copy(tmp[:], addr)
		addrs = append(addrs, tmp)
	}
	return
}

// SignTransaction signs the passed transaction with the private key whose public key is provided, or
// if the provided public key is zero, inferring the required private key from the transaction itself
func (pqw *ParquetWallet) SignTransaction(tx types.Transaction, pk ed25519.PublicKey, pw []byte) (stx []byte, err error) {
	// Check the password
	err = pqw.CheckPassword(pw)
	if err != nil {
		return
	}

	// Fetch the required key
	var sk ed25519.PrivateKey
	if (slices.Equal(pk, ed25519.PublicKey{})) {
		sk, err = pqw.fetchSecretKey(types.Digest(tx.Sender))
	} else {
		sk, err = pqw.fetchSecretKey(types.Digest(pk))
	}
	if err != nil {
		return
	}

	// Sign the transaction with the required key
	_, stx, err = crypto.SignTransaction(sk, tx)

	return
}

// SignProgram signs the passed data for the src address
func (pqw *ParquetWallet) SignProgram(data []byte, src types.Digest, pw []byte) (sprog []byte, err error) {
	// Check the password
	err = pqw.CheckPassword(pw)
	if err != nil {
		return
	}

	// Fetch the required key
	sk, err := pqw.fetchSecretKey(types.Digest(src))
	if err != nil {
		return
	}

	sprog, err = crypto.SignBytes(sk, data)

	return
}

// MultisigSignTransaction starts a multisig signature or adds a signature to a
// partially signed multisig transaction signature of the passed transaction
// using the key
func (pqw *ParquetWallet) MultisigSignTransaction(tx types.Transaction, pk ed25519.PublicKey, partial types.MultisigSig, pw []byte, signer types.Digest) (sig types.MultisigSig, err error) {
	// Check the password
	err = pqw.CheckPassword(pw)
	if err != nil {
		return
	}

	if partial.Version == 0 && partial.Threshold == 0 && len(partial.Subsigs) == 0 {
		// We weren't given a partial multisig, so create a new one

		// Look up the preimage in the database
		var pks []ed25519.PublicKey
		var version, threshold uint8
		version, threshold, pks, err = pqw.LookupMultisigPreimage(types.Digest(tx.Sender))
		if err != nil {
			return
		}

		// Fetch the required secret key (the secret key for the given public
		// key)
		var sk ed25519.PrivateKey
		sk, err = pqw.fetchSecretKey(publicKeyToAddress(pk))
		if err != nil {
			return
		}

		// Create multisig account for signing the transaction
		ma := crypto.MultisigAccount{
			Version:   version,
			Threshold: threshold,
			Pks:       pks,
		}

		// Sign transaction
		var stx []byte // Signed transaction bytes
		_, stx, err = crypto.SignMultisigTransaction(sk, ma, tx)
		if err != nil {
			return
		}

		// Create new a partial multisig

		subsigs := make([]types.MultisigSubsig, len(pks))

		// Convert signed transaction bytes from a []byte to a
		// [ed25519.SignatureSize]byte subsignature so it can be placed into the
		// collection of subsigs
		subsig := [ed25519.SignatureSize]byte{}
		copy(subsig[:], stx)

		// Insert the subsig into the collection of subsigs. The index of the
		// subsig within the subsigs slice must match the index of the
		// corresponding public key within the public keys slice
		for i, multisigPk := range pks {
			if pk.Equal(multisigPk) {
				subsigs[i] = types.MultisigSubsig{Key: pk, Sig: subsig}
			}
		}

		sig = types.MultisigSig{
			Version:   version,
			Threshold: threshold,
			Subsigs:   subsigs,
		}

		return
	}

	// We were given a partial multisig, so add to it

	// Convert partial multisig to a partial multisig "account"
	var ma crypto.MultisigAccount
	ma, err = crypto.MultisigAccountFromSig(partial)
	if err != nil {
		return
	}

	// Check preimage matches tx src address
	var addr types.Address
	addr, err = ma.Address()
	if err != nil {
		return
	}

	// Convert from `Address` to `Digest`
	multisigAddr := types.Digest(addr)

	// Check that the multisig address equals to either sender or signer
	if multisigAddr != types.Digest(tx.Sender) && multisigAddr != signer {
		err = errMsigWrongAddr
		return
	}

	// Check that key is one of the ones in the preimage
	err = errMsigWrongKey
	for _, subsig := range partial.Subsigs {
		if pk.Equal(subsig.Key) {
			err = nil
			break
		}
	}
	if err != nil {
		return
	}

	// Fetch the required secret key
	sk, err := pqw.fetchSecretKey(publicKeyToAddress(pk))
	if err != nil {
		return
	}

	// Sign transaction
	var stx []byte // Signed transaction bytes
	_, stx, err = crypto.SignMultisigTransaction(sk, ma, tx)
	if err != nil {
		return
	}

	// Create new partial multisig with new signature added

	subsigs := partial.Subsigs

	// Convert signed transaction bytes from a []byte to a
	// [ed25519.SignatureSize]byte subsignature so it can be placed into the
	// collection of subsigs
	subsig := [ed25519.SignatureSize]byte{}
	copy(subsig[:], stx)

	// Insert the subsig into the collection of subsigs. The index of the
	// subsig within the subsigs slice must match the index of the
	// corresponding public key within the public keys slice
	for i, multisigPk := range ma.Pks {
		if pk.Equal(multisigPk) {
			subsigs[i] = types.MultisigSubsig{Key: pk, Sig: subsig}
		}
	}

	sig = types.MultisigSig{
		Version:   partial.Version,
		Threshold: partial.Threshold,
		Subsigs:   subsigs,
	}

	return
}

// MultisigSignProgram starts a multisig signature or adds a signature to a
// partially signed multisig transaction signature of the passed transaction
// using the key
func (pqw *ParquetWallet) MultisigSignProgram(data []byte, src types.Digest, pk ed25519.PublicKey, partial types.MultisigSig, pw []byte) (sig types.MultisigSig, err error) {
	// Check the password
	err = pqw.CheckPassword(pw)
	if err != nil {
		return
	}

	if partial.Version == 0 && partial.Threshold == 0 && len(partial.Subsigs) == 0 {
		// We weren't given a partial multisig, so create a new one

		// Look up the preimage in the database
		var pks []ed25519.PublicKey
		var version, threshold uint8
		version, threshold, pks, err = pqw.LookupMultisigPreimage(src)
		if err != nil {
			return
		}

		// Fetch the required secret key
		var sk ed25519.PrivateKey
		sk, err = pqw.fetchSecretKey(publicKeyToAddress(pk))
		if err != nil {
			return
		}

		// Sign program
		var sprog []byte // Signed program bytes
		sprog, err = crypto.SignBytes(sk, data)
		if err != nil {
			return
		}

		// Create new a partial multisig

		subsigs := make([]types.MultisigSubsig, len(pks))

		// Convert signed program bytes from a []byte to a
		// [ed25519.SignatureSize]byte subsignature so it can be placed into the
		// collection of subsigs
		subsig := [ed25519.SignatureSize]byte{}
		copy(subsig[:], sprog)

		// Insert the subsig into the collection of subsigs. The index of the
		// subsig within the subsigs slice must match the index of the
		// corresponding public key within the public keys slice
		for i, multisigPk := range pks {
			if pk.Equal(multisigPk) {
				subsigs[i] = types.MultisigSubsig{Key: pk, Sig: subsig}
			}
		}

		sig = types.MultisigSig{
			Version:   version,
			Threshold: threshold,
			Subsigs:   subsigs,
		}
		return
	}

	// We were given a partial multisig, so add to it

	// Convert partial multisig to a partial multisig "account"
	var ma crypto.MultisigAccount
	ma, err = crypto.MultisigAccountFromSig(partial)
	if err != nil {
		return
	}

	// Check preimage matches tx src address
	var addr types.Address
	addr, err = ma.Address()
	if err != nil {
		return
	}
	if types.Digest(addr) != src {
		err = errMsigWrongAddr
		return
	}

	// Check that key is one of the ones in the preimage
	err = errMsigWrongKey
	for _, subsig := range partial.Subsigs {
		if pk.Equal(subsig.Key) {
			err = nil
			break
		}
	}
	if err != nil {
		return
	}

	// Fetch the required secret key
	sk, err := pqw.fetchSecretKey(publicKeyToAddress(pk))
	if err != nil {
		return
	}

	// Sign program
	var sprog []byte // Signed program bytes
	sprog, err = crypto.SignBytes(sk, data)
	if err != nil {
		return
	}

	// Create new partial multisig with new signature merged into it

	subsigs := partial.Subsigs

	// Convert signed program bytes from a []byte to a
	// [ed25519.SignatureSize]byte subsignature so it can be placed into the
	// collection of subsigs
	subsig := [ed25519.SignatureSize]byte{}
	copy(subsig[:], sprog)

	// Insert the subsig into the collection of subsigs. The index of the
	// subsig within the subsigs slice must match the index of the
	// corresponding public key within the public keys slice
	for i, multisigPk := range ma.Pks {
		if pk.Equal(multisigPk) {
			subsigs[i] = types.MultisigSubsig{Key: pk, Sig: subsig}
		}
	}

	sig = types.MultisigSig{
		Version:   partial.Version,
		Threshold: partial.Threshold,
		Subsigs:   subsigs,
	}

	return
}

/*******************************************************************************
 * Helpers
 ******************************************************************************/

// Initialize the codec
func init() {
	codecHandle = new(codec.MsgpackHandle)
	codecHandle.ErrorIfNoField = true
	codecHandle.ErrorIfNoArrayExpand = true
	codecHandle.Canonical = true
	codecHandle.RecursiveEmptyCheck = true
	codecHandle.WriteExt = true
	codecHandle.PositiveIntUnsigned = true
}

// parquetDbConnectionURL takes a path to a Parquet database on the filesystem and
// constructs a proper connection URL from it with feature flags included
func parquetDbConnectionURL(path string) string {
	// Set flags on the database connection. For all options see:
	// https://pkg.go.dev/modernc.org/sqlite#Driver.Open
	return fmt.Sprintf("file:%s?%s", path, parquetWalletDBOptions)
}

// parquetWalletMetadataFromPath accepts path to a directory for a wallet and
// returns a Metadata struct with information about it
func parquetWalletMetadataFromPath(walletPath string) (metadata ParquetWalletMetadata, err error) {
	// Read wallet's metadata.json
	metadataFileContents, err := os.ReadFile(walletPath + "/" + ParquetWalletMetadataFile)
	if err != nil {
		return
	}

	// Parse JSON contents
	metadata = ParquetWalletMetadata{}
	err = json.Unmarshal(metadataFileContents, &metadata)
	if err != nil {
		return
	}

	return
}

// checkParquetDBError inspects an error from the database and interprets it as a
// "duplicate key" error, a generic database error, or a nil error
func checkParquetDBError(err error) error {
	if err != nil {
		if err.Error() == sqlite.ErrorCodeString[sqlite3.SQLITE_CONSTRAINT] {
			// If it was a constraint error, that means we already have the key.
			return errKeyExists
		}
		// Otherwise, return a generic database error
		return errDatabase
	}
	return nil
}

// removeTempFile attempts to remove the temporary file used when modifying the
// file with the given file name.
func removeTempFile(originalFilename string) error {
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

/********** Wallet Driver Helpers **********/

// potentialWalletPaths lists paths to plausible wallet directories in the
// wallets directory. This means things that are directories and contain a
// `metadata.json` file.
func (parqwd *ParquetWalletDriver) potentialWalletPaths() (paths []string, err error) {
	// List all files and folders in the wallets directory
	wDir := parqwd.walletsDir()
	files, err := os.ReadDir(wDir)
	if err != nil {
		return
	}

	for _, f := range files {
		// Skip files
		if !f.IsDir() {
			continue
		}

		// Skip directories that don't have a metadata.json file
		_, err := os.Stat(wDir + "/" + f.Name() + "/" + ParquetWalletMetadataFile)
		if err != nil {
			continue
		}

		paths = append(paths, filepath.Join(wDir, f.Name()))
	}

	return
}

// maybeMakeWalletsDir tries to create the wallets directory if it doesn't
// already exist
func (parqwd *ParquetWalletDriver) maybeMakeWalletsDir() error {
	wDir := parqwd.walletsDir()
	err := os.Mkdir(wDir, parquetWalletsDirPermissions)
	if err != nil && !os.IsExist(err) {
		return fmt.Errorf("couldn't create wallets directory at %s: %v", wDir, err)
	}
	return nil
}

// walletsDir returns the wallet directory specified in the config, if there
// is one, otherwise it returns a subdirectory of the global kmd data dir
func (parqwd *ParquetWalletDriver) walletsDir() string {
	if parqwd.parquetCfg.WalletsDir != "" {
		return parqwd.parquetCfg.WalletsDir
	}
	return filepath.Join(parqwd.globalCfg.DataDir, parquetWalletsDirName)
}

// idToPath turns a wallet id into a path to the corresponding wallet directory
// to create
func (parqwd *ParquetWalletDriver) idToPath(id []byte) string {
	// wallet ID should already be safe, but filter it just in case
	safeID := disallowedParquetFilenameRegex.ReplaceAll(id, []byte(""))
	// The directory name for the wallet should be the wallet ID
	return filepath.Join(parqwd.walletsDir(), string(safeID))
}

// addParquetWalletMetadata adds the given data to the metadatas file for the
// Parquet wallet driver
func (parqwd *ParquetWalletDriver) addParquetWalletMetadata(metadata *ParquetWalletMetadata) error {
	// Create the temporary database
	db, err := sqlx.Connect("duckdb", "")
	if err != nil {
		return err
	}
	defer db.Close()

	// Run the schema for creating the metadatas schema in temporary database
	_, err = db.Exec(parquetCreateMetadatasTblSchema)
	if err != nil {
		return err
	}

	// Store the metadata row in the temporary database
	_, err = db.Exec(
		"INSERT INTO metadatas (driver_name, driver_version, wallet_id, wallet_name, mep_encrypted, mdk_encrypted, max_key_idx_encrypted) VALUES(?, ?, ?, ?, ?, ?, ?)",
		parquetWalletDriverName,
		parquetWalletDriverVersion,
		metadata.WalletId,
		metadata.WalletName,
		metadata.MEPEncrypted,
		metadata.MDKEncrypted,
		metadata.MaxKeyIdxEncrypted,
	)
	if err != nil {
		return err
	}

	metadatasPath := parqwd.walletsDir() + "/" + ParquetMetadatasFile

	// Check if metadatas file exists
	_, err = os.Stat(metadatasPath)
	if err != nil && !os.IsNotExist(err) {
		// Some unexpected error
		return err
	}

	// Write the metadata in the temporary table into the file
	if !os.IsNotExist(err) { // If metadatas file exists
		var retrievedWalletId string

		// Check if metadata is already stored in the file
		row := db.QueryRow(
			fmt.Sprintf("SELECT wallet_id from read_parquet('%s') WHERE wallet_id = ?", metadatasPath),
			metadata.WalletId,
		)
		scanErr := row.Scan(&retrievedWalletId)
		if scanErr != sql.ErrNoRows {
			return errSameID
		}

		// Combine the metadatas within the file with the new metadata
		_, err = db.Exec(fmt.Sprintf(
			"COPY ((FROM read_parquet('%s') UNION FROM metadatas) ORDER BY wallet_id) TO '%s' (FORMAT parquet)",
			metadatasPath, metadatasPath+tempFileSuffix,
		))
		if err != nil {
			return err
		}

		err = removeTempFile(metadatasPath)
		if err != nil {
			return err
		}
	} else { // Metadatas file does not exist
		// Create a new file and write the metadata into it
		_, err = db.Exec(fmt.Sprintf("COPY metadatas TO '%s' (FORMAT parquet)", metadatasPath))
		if err != nil {
			return err
		}
	}

	return nil
}

/********** Wallet Helpers **********/

// decryptAndGetMasterKey fetches the master key from the metadata table and
// attempts to decrypt it with the passed password
func (pqw *ParquetWallet) decryptAndGetMasterKey(pw []byte) ([]byte, error) {
	// Connect to the database
	db, err := sqlx.Connect("sqlite", parquetDbConnectionURL(pqw.dbPath))
	if err != nil {
		return nil, errDatabaseConnect
	}
	defer db.Close()

	var encryptedMEPBlob []byte
	err = db.Get(&encryptedMEPBlob, "SELECT mep_encrypted FROM metadata LIMIT 1")
	if err != nil {
		return nil, errDatabase
	}

	mep, err := decryptBlobWithPassword(encryptedMEPBlob, PTMasterKey, pw)
	if err != nil {
		return nil, err
	}

	return mep, nil
}

// decryptAndGetMasterDerivationKey fetches the mdk from the metadata table and
// attempts to decrypt it with the master password
func (pqw *ParquetWallet) decryptAndGetMasterDerivationKey(pw []byte) ([]byte, error) {
	// Connect to the database
	db, err := sqlx.Connect("sqlite", parquetDbConnectionURL(pqw.dbPath))
	if err != nil {
		return nil, errDatabaseConnect
	}
	defer db.Close()

	var encryptedMDKBlob []byte
	err = db.Get(&encryptedMDKBlob, "SELECT mdk_encrypted FROM metadata LIMIT 1")
	if err != nil {
		return nil, errDatabase
	}

	mdk, err := decryptBlobWithPassword(encryptedMDKBlob, PTMasterDerivationKey, pw)
	if err != nil {
		return nil, err
	}

	return mdk, nil
}

// fetchSecretKey retrieves the private key for a given public key
func (pqw *ParquetWallet) fetchSecretKey(addr types.Digest) (sk ed25519.PrivateKey, err error) {
	// Connect to the database
	db, err := sqlx.Connect("sqlite", parquetDbConnectionURL(pqw.dbPath))
	if err != nil {
		err = errDatabaseConnect
		return
	}
	defer db.Close()

	var skCandidate ed25519.PrivateKey
	var blob []byte

	// Fetch the encrypted secret key from the database
	err = db.Get(&blob, "SELECT secret_key_encrypted FROM keys WHERE address=?", addr[:])
	if err != nil {
		err = errKeyNotFound
		return
	}

	// Decrypt the master encryption key stored in enclave into a local copy
	mekBuf, err := pqw.masterEncryptionKey.Open()
	if err != nil {
		return
	}
	defer mekBuf.Destroy() // Destroy the copy when we return

	// Decrypt the secret key
	skEncoded, err := decryptBlobWithPassword(blob, PTSecretKey, mekBuf.Bytes())
	if err != nil {
		return
	}

	// Decode the secret key candidate
	err = msgpackDecode(skEncoded, &skCandidate)
	if err != nil {
		return
	}

	// Extract the public key from the candidate secret key
	derivedPK := skCandidate.Public().(ed25519.PublicKey)

	// Convert the derived public key to an address
	derivedAddr := publicKeyToAddress(derivedPK)

	// Ensure the derived address matches the one we used to look the key up
	if addr != derivedAddr {
		err = errTampering
		return
	}

	// The candidate looks good, return it
	sk = skCandidate
	return
}

// generateKeyTxLocked is a helper for GenerateKey that accepts a locked tx,
// computes the next key that should be generated, inserts it, and returns
// its address
func (pqw *ParquetWallet) generateKeyTxLocked(tx *sqlx.Tx) (addr types.Digest, err error) {
	// Fetch the encrypted highest index
	var encryptedHighestIndexBlob []byte
	err = tx.Get(&encryptedHighestIndexBlob, "SELECT max_key_idx_encrypted FROM metadata LIMIT 1")
	if err != nil {
		err = errDatabase
		return
	}

	// Decrypt the master encryption key stored in enclave into a local copy
	mekBuf, err := pqw.masterEncryptionKey.Open()
	if err != nil {
		return
	}
	defer mekBuf.Destroy() // Destroy the copy when we return

	// Decrypt the highest index
	highestIndexBlob, err := decryptBlobWithPassword(encryptedHighestIndexBlob, PTMaxKeyIdx, mekBuf.Bytes())
	if err != nil {
		return
	}

	// Decode the highest index
	var highestIndex uint64
	err = msgpackDecode(highestIndexBlob, &highestIndex)
	if err != nil {
		return
	}

	// nextIndex is the index of the next key we should generate
	nextIndex := highestIndex + 1

	var genPK ed25519.PublicKey
	var genSK ed25519.PrivateKey
	var genAddr types.Digest

	// We may have to bump nextIndex if the user has manually imported the next
	// key we were going to generate (thus we didn't see it in the search for the
	// highest-derived key above)
	for {
		// Honestly, if you could get 2**63 - 1 keys into this database, I'd be impressed
		if nextIndex == parquetIntOverflow {
			err = errTooManyKeys
			return
		}

		// Decrypt the key stored in enclave into a local copy
		var mdkBuf *memguard.LockedBuffer
		mdkBuf, err = pqw.masterDerivationKey.Open()
		if err != nil {
			return addr, err
		}
		defer mdkBuf.Destroy() // Destroy the copy when we return

		// Compute the secret key and public key for nextIndex
		genPK, genSK, err = extractKeyWithIndex(mdkBuf.Bytes(), nextIndex)
		if err != nil {
			return
		}

		// Convert the public key into an address
		genAddr = publicKeyToAddress(genPK)

		// Check that we don't already have this PK in the database
		var cnt int
		err = tx.Get(&cnt, "SELECT COUNT(1) FROM keys WHERE address=?", genAddr[:])
		if err != nil {
			err = errDatabase
			return
		}

		if cnt == 0 {
			// Good, key didn't exist. Break from loop
			break
		}

		// Uh oh, user already imported this key manually. Bump nextIndex
		nextIndex++
	}

	// Encrypt the encoded secret key
	skEncrypted, err := encryptBlobWithKey(msgpackEncode(genSK), PTSecretKey, mekBuf.Bytes())
	if err != nil {
		return
	}

	// Insert the key into the database
	_, err = tx.Exec("INSERT INTO keys (address, secret_key_encrypted, key_idx) VALUES(?, ?, ?)", genAddr[:], skEncrypted, nextIndex)
	if err != nil {
		return
	}

	// Encrypt the new max key index
	encryptedIdxBlob, err := encryptBlobWithKey(msgpackEncode(nextIndex), PTMaxKeyIdx, mekBuf.Bytes())
	if err != nil {
		return
	}

	// Update the metadata row
	_, err = tx.Exec("UPDATE metadata SET max_key_idx_encrypted = ?", encryptedIdxBlob)
	if err != nil {
		return
	}

	// Return the generated public key
	return genAddr, nil
}
