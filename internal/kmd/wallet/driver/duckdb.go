// XXX: This driver is a modified version of a modified version of the sqlite
// driver in go-algorand. This is modified to uses DuckDB (with encryption)
// instead of SQLite.
// Modified SQLite driver: https://github.com/No-Cash-7970/duckySigner/blob/0f5365f1062a36c20af88379dd61d53693314b11/internal/kmd/wallet/driver/sqlite.go
// Original SQLite driver: https://github.com/algorand/go-algorand/tree/c2d7047585f6109d866ebaf9fca0ee7490b16c6a/daemon/kmd/wallet/driver/sqlite.go

package driver

import (
	"crypto/ed25519"
	"crypto/subtle"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"duckysigner/internal/kmd/config"
	kmdCrypto "duckysigner/internal/kmd/crypto"
	"duckysigner/internal/kmd/wallet"

	"github.com/algorand/go-algorand-sdk/v2/crypto"
	"github.com/algorand/go-algorand-sdk/v2/types"
	"github.com/algorand/go-codec/codec"
	"github.com/awnumar/memguard"
	_ "github.com/duckdb/duckdb-go/v2"
	"github.com/onsi/ginkgo/v2"
	logging "github.com/sirupsen/logrus"
)

const (
	duckDbWalletDriverName      = "duckdb"
	duckDbWalletDriverVersion   = 1
	duckDbWalletsDirName        = "duckdb_wallets"
	duckDbWalletsDirPermissions = 0700
	duckDbMaxWalletNameLen      = 64
	duckDbMaxWalletIDLen        = 64
	duckDbIntOverflow           = 1 << 63
	duckDbWalletHasMnemonicUX   = false
	duckDbWalletHasMasterKey    = true
	// DuckDbWalletMetadataFile is the name of the metadata file is in each
	// directory for a wallet
	DuckDbWalletMetadataFile = "metadata.json"
	// DuckDbWalletAcctsFile is the name of the file that contains the keys and
	// data of the wallet's accounts
	DuckDbWalletAcctsFile = "accounts.duckdb"
)

var duckDbWalletSupportedTxs = []types.TxType{
	types.PaymentTx,
	types.KeyRegistrationTx,
	types.ApplicationCallTx,
	types.AssetConfigTx,
	types.AssetFreezeTx,
	types.AssetTransferTx,
}
var disallowedDuckDbFilenameRegex = regexp.MustCompile("[^a-zA-Z0-9_-]*")

// attachEncDuckDbSQL is the SQL statement for opening or creating an encrypted
// DuckDB file. The file encryption key is used for encrypting and decrypting
// all the encrypted Duck DB files. Requires the file encryption key in Base64.
// NOTE: These SQL statements will create the file if it does not exist.
const attachEncDuckDbSQL = `
LOAD httpfs; -- use OpenSSL library to increase speed
ATTACH '%s' AS db (ENCRYPTION_KEY '%s');
`

var duckDbAcctsSchema = `
CREATE TABLE IF NOT EXISTS db.info (
	mdk BLOB PRIMARY KEY,
	max_key_idx INT NOT NULL
);

CREATE TYPE db.acct_type AS ENUM ('kmd_hd', 'standalone', 'ledger', 'msig', 'watch');

CREATE TABLE IF NOT EXISTS db.addresses (
	addr BLOB PRIMARY KEY,
    pos_num INT,
    type acct_type NOT NULL,
    name TEXT,
    rekeyed BOOL NOT NULL
);

CREATE TABLE IF NOT EXISTS db.keys (
	address BLOB PRIMARY KEY,
	key BLOB NOT NULL,
	key_idx INT
);

CREATE TABLE IF NOT EXISTS db.msig_addrs (
	address BLOB PRIMARY KEY,
	version INT NOT NULL,
	threshold INT NOT NULL,
	pks BLOB NOT NULL
);
`

// DuckDbWalletDriver is the default wallet driver used by kmd. Keys are stored
// as authenticated-encrypted blobs in a sqlite 3 database.
type DuckDbWalletDriver struct {
	globalCfg config.KMDConfig
	duckDbCfg config.DuckDbWalletDriverConfig
}

// DuckDbWallet represents a particular DuckDbWallet under the
// DuckDbWalletDriver
type DuckDbWallet struct {
	masterEncryptionKey  *memguard.Enclave
	masterDerivationKey  *memguard.Enclave
	walletPasswordSalt   [saltLen]byte
	walletPasswordHash   types.Digest
	walletPasswordHashed bool
	cfg                  config.DuckDbWalletDriverConfig
	id                   string
	// The parent directory of the wallet where wallets are stored
	walletsPath string
	// If the wallet has been initialized
	initialized bool
}

type DuckDbWalletMetadata struct {
	DriverName    string `json:"driver_name"`
	DriverVersion int    `json:"driver_version"`
	WalletId      string `json:"wallet_id"`
	WalletName    string `json:"wallet_name"`
	MEPEncrypted  []byte `json:"mep_encrypted"`
}

/*******************************************************************************
 * Wallet Driver
 ******************************************************************************/

// InitWithConfig accepts a driver configuration so that the DuckDb driver
// knows where to read and write its wallet databases
func (ddbwd *DuckDbWalletDriver) InitWithConfig(cfg config.KMDConfig, log *logging.Logger) error {
	ddbwd.globalCfg = cfg
	ddbwd.duckDbCfg = cfg.DriverConfig.DuckDbWalletDriverConfig

	// Make sure the scrypt params are reasonable
	if !ddbwd.duckDbCfg.UnsafeScrypt {
		if ddbwd.duckDbCfg.ScryptParams.ScryptN < minScryptN {
			return fmt.Errorf("slow scrypt N must be at least %d", minScryptN)
		}
		if ddbwd.duckDbCfg.ScryptParams.ScryptR < minScryptR {
			return fmt.Errorf("slow scrypt R must be at least %d", minScryptR)
		}
		if ddbwd.duckDbCfg.ScryptParams.ScryptP < minScryptP {
			return fmt.Errorf("slow scrypt P must be at least %d", minScryptP)
		}
	}

	// Make the wallets directory if it doesn't already exist
	err := ddbwd.maybeMakeWalletsDir()
	if err != nil {
		return err
	}

	return nil
}

// ListWalletMetadatas opens everything that looks like a wallet in the
// walletsDir() and tries to extract its metadata. It does not fail if it
// is unable to read metadata of one of the "wallets".
func (ddbwd *DuckDbWalletDriver) ListWalletMetadatas() (metadatas []wallet.Metadata, err error) {
	// Do not list if this wallet driver is disabled
	if ddbwd.duckDbCfg.Disable {
		return []wallet.Metadata{}, nil
	}

	paths, err := ddbwd.potentialWalletPaths()
	if err != nil {
		return
	}
	for _, path := range paths {
		// Get metadata from path (if possible)
		walletMetadata, err := duckDbWalletMetadataFromPath(path)
		if err != nil {
			continue
		}

		metadatas = append(metadatas, wallet.Metadata{
			ID:                    []byte(walletMetadata.WalletId),
			Name:                  []byte(walletMetadata.WalletName),
			DriverName:            walletMetadata.DriverName,
			DriverVersion:         uint32(walletMetadata.DriverVersion),
			SupportsMnemonicUX:    duckDbWalletHasMnemonicUX,
			SupportsMasterKey:     duckDbWalletHasMasterKey,
			SupportedTransactions: duckDbWalletSupportedTxs,
		})
	}
	return metadatas, nil
}

// CreateWallet creates a wallet with the given name and ID that will be
// protected with the given password and Master Derivation Key (MDK). Providing
// the MDK is optional. If an MDK is not provided, then one will be generated.
func (ddbwd *DuckDbWalletDriver) CreateWallet(name []byte, id []byte, pw []byte, mdk types.MasterDerivationKey) error {
	if len(name) > duckDbMaxWalletNameLen {
		return errNameTooLong
	}

	if len(id) > duckDbMaxWalletIDLen {
		return errIDTooLong
	}

	walletPath := ddbwd.idToPath(id)

	// Create directory for new wallet
	err := os.Mkdir(walletPath, duckDbWalletsDirPermissions)
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
	encryptedMEPBlob, err := encryptBlobWithPasswordBlankOK(masterKey[:], PTMasterKey, pw, &ddbwd.duckDbCfg.ScryptParams)
	if err != nil {
		return err
	}

	// // Encrypt the master derivation key using the master encryption password
	// // (which may not be blank)
	// encryptedMDKBlob, err := encryptBlobWithKey(masterDerivationKey[:], PTMasterDerivationKey, masterKey[:])
	// if err != nil {
	// 	return err
	// }

	metadata := DuckDbWalletMetadata{
		DriverName:    duckDbWalletDriverName,
		DriverVersion: duckDbWalletDriverVersion,
		WalletId:      string(id),
		WalletName:    string(name),
		MEPEncrypted:  encryptedMEPBlob,
	}

	// Create metadata.json file in new wallet directory
	metadataJson, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	err = os.WriteFile(
		filepath.Join(walletPath, DuckDbWalletMetadataFile),
		metadataJson,
		duckDbWalletsDirPermissions,
	)
	if err != nil {
		return err
	}

	// Open database
	db, err := sql.Open("duckdb", "")
	if err != nil {
		return err
	}
	defer db.Close()

	// Encrypt accounts database file using master encryption password
	// NOTE: This SQL statement creates the file if it does not exist
	_, err = db.Exec(fmt.Sprintf(attachEncDuckDbSQL,
		filepath.Join(walletPath, DuckDbWalletAcctsFile),
		base64.StdEncoding.EncodeToString(masterKey[:]),
	))
	if err != nil {
		return err
	}

	ginkgo.GinkgoWriter.Println("DB key:", base64.StdEncoding.EncodeToString(masterDerivationKey[:]))

	// Run the schema for creating the tables in accounts database
	_, err = db.Exec(duckDbAcctsSchema)
	if err != nil {
		return err
	}

	// Add max index to info table
	_, err = db.Exec("INSERT INTO db.info (mdk, max_key_idx) VALUES(?, ?)", masterDerivationKey[:], 0)
	if err != nil {
		return err
	}

	return nil
}

// FetchWallet looks up a wallet by ID and returns it. The wallet returned by
// this function is uninitialized and will need to be initialized before using
// most of the wallet function.
func (ddbwd *DuckDbWalletDriver) FetchWallet(id []byte) (wallet.Wallet, error) {
	if len(id) == 0 {
		return &DuckDbWallet{}, fmt.Errorf("no ID is given")
	}

	// Check wallet exists by checking existence of metadata and accounts files

	walletsDir := ddbwd.walletsDir()
	metadataPath := filepath.Join(walletsDir, string(id), DuckDbWalletMetadataFile)
	acctsPath := filepath.Join(walletsDir, string(id), DuckDbWalletAcctsFile)

	// Check if metadatas file exists
	_, err := os.Stat(metadataPath)
	if err != nil {
		return &DuckDbWallet{}, errWalletNotFound
	}

	// Check if accounts database file exists
	_, err = os.Stat(acctsPath)
	if err != nil {
		return &DuckDbWallet{}, errWalletNotFound
	}

	// Fill in the wallet details
	return &DuckDbWallet{
		id:          string(id),
		walletsPath: walletsDir,
		cfg:         ddbwd.duckDbCfg,
	}, nil
}

// RenameWallet renames the wallet with the given id to newName. The given
// password is ignored, so the wallet can be successfully renamed if password is
// incorrect. The password can be left empty.
func (ddbwd *DuckDbWalletDriver) RenameWallet(newName []byte, id []byte, pw []byte) error {
	if len(id) == 0 {
		return fmt.Errorf("no ID is given")
	}

	if len(newName) > duckDbMaxWalletNameLen {
		return errNameTooLong
	}

	walletMetadataPath := filepath.Join(ddbwd.walletsDir(), string(id), DuckDbWalletMetadataFile)

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
	metadata := DuckDbWalletMetadata{}
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
	err = os.WriteFile(walletMetadataPath+tempFileSuffix, updatedMetadataJson, duckDbWalletsDirPermissions)
	if err != nil {
		return err
	}
	err = removeTempFile(walletMetadataPath)
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
func (ddbw *DuckDbWallet) Metadata() (wallet.Metadata, error) {
	walletPath := filepath.Join(ddbw.walletsPath, ddbw.id)

	retrievedMetadata, err := duckDbWalletMetadataFromPath(walletPath)
	if err != nil {
		return wallet.Metadata{}, err
	}

	return wallet.Metadata{
		ID:                    []byte(retrievedMetadata.WalletId),
		Name:                  []byte(retrievedMetadata.WalletName),
		DriverName:            retrievedMetadata.DriverName,
		DriverVersion:         uint32(retrievedMetadata.DriverVersion),
		SupportsMnemonicUX:    duckDbWalletHasMnemonicUX,
		SupportsMasterKey:     duckDbWalletHasMasterKey,
		SupportedTransactions: duckDbWalletSupportedTxs,
	}, nil
}

// Init attempts to decrypt the master encrypt password and master derivation
// key, and store them in memory for subsequent operations
func (ddbw *DuckDbWallet) Init(pw []byte) error {
	// Decrypt the master password
	masterEncryptionKey, err := ddbw.DecryptAndGetMasterKey(pw)
	if err != nil {
		return err
	}

	// Decrypt the master derivation key
	masterDerivationKey, err := ddbw.decryptAndGetMasterDerivationKey(pw)
	if err != nil {
		return err
	}

	// Initialize wallet
	ddbw.masterEncryptionKey = memguard.NewEnclave(masterEncryptionKey)
	ddbw.masterDerivationKey = memguard.NewEnclave(masterDerivationKey)
	err = fillRandomBytes(ddbw.walletPasswordSalt[:])
	if err != nil {
		return err
	}
	ddbw.walletPasswordHash = fastHashWithSalt(pw, ddbw.walletPasswordSalt[:])
	ddbw.walletPasswordHashed = true

	ddbw.initialized = true

	return nil
}

// CheckPassword checks that the database can be decrypted with the password.
// It's the same as Init but doesn't store the decrypted key
func (ddbw *DuckDbWallet) CheckPassword(pw []byte) error {
	if ddbw.walletPasswordHashed {
		// Check against pre-computed password hash
		pwhash := fastHashWithSalt(pw, ddbw.walletPasswordSalt[:])
		if subtle.ConstantTimeCompare(pwhash[:], ddbw.walletPasswordHash[:]) == 1 {
			return nil
		}
		return errDecrypt
	}

	_, err := ddbw.DecryptAndGetMasterKey(pw)
	return err
}

// CheckAddrInWallet checks if the account with the given address is stored in
// the wallet
func (ddbw *DuckDbWallet) CheckAddrInWallet(addr string) (bool, error) {
	if !ddbw.initialized {
		return false, fmt.Errorf("wallet not initialized")
	}

	decodedAddr, err := types.DecodeAddress(addr)
	if err != nil {
		return false, err
	}

	// Attempt to fetch key
	_, err = ddbw.fetchSecretKey(types.Digest(decodedAddr))
	if err == errKeyNotFound {
		return false, nil
	}
	if err != nil {
		// Some unexpected error occurred
		return false, err
	}

	return true, nil
}

// ListKeys lists all the addresses in the wallet
func (ddbw *DuckDbWallet) ListKeys() (addrs []types.Digest, err error) {
	if !ddbw.initialized {
		return addrs, fmt.Errorf("wallet not initialized")
	}

	// Open database
	db, err := ddbw.openAccountsDb()
	if err != nil {
		return
	}
	defer db.Close()

	// Get all stored addresses
	rows, err := db.Query("SELECT address FROM db.keys")
	if err != nil {
		return
	}

	for rows.Next() {
		// We can't select directly into a types.Digest array, unfortunately.
		// Instead, we select into a slice of byte slices, and then convert each
		// of those slices into a types.Digest.

		var retrievedAddr []byte
		var tmp types.Digest

		err = rows.Scan(&retrievedAddr)
		if err != nil {
			return
		}

		copy(tmp[:], retrievedAddr)
		addrs = append(addrs, tmp)
	}

	return
}

// ExportMasterDerivationKey decrypts the encrypted MDK and returns it
func (ddbw *DuckDbWallet) ExportMasterDerivationKey(pw []byte) (mdk types.MasterDerivationKey, err error) {
	if !ddbw.initialized {
		return mdk, fmt.Errorf("wallet not initialized")
	}

	// Check the password
	err = ddbw.CheckPassword(pw)
	if err != nil {
		return
	}

	// Decrypt the master derivation key stored in enclave into a local copy
	mdkBuf, err := ddbw.masterDerivationKey.Open()
	if err != nil {
		return
	}
	defer mdkBuf.Destroy() // Destroy the copy when we return

	// Copy master derivation key into the result
	copy(mdk[:], mdkBuf.Bytes())

	return
}

// ImportKey imports a key pair into the wallet, deriving the public key from
// the passed secret key
func (ddbw *DuckDbWallet) ImportKey(rawSK ed25519.PrivateKey) (addr types.Digest, err error) {
	if !ddbw.initialized {
		return addr, fmt.Errorf("wallet not initialized")
	}

	// Extract the seed from the secret key so that we don't trust the public part
	seed := rawSK.Seed()

	// Convert the seed to an sk/pk pair
	sk := ed25519.NewKeyFromSeed(seed[:])
	pk := sk.Public().(ed25519.PublicKey)
	addr = publicKeyToAddress(pk)

	// Open database
	db, err := ddbw.openAccountsDb()
	if err != nil {
		return
	}
	defer db.Close()

	// Insert the pk, e(sk) into the temporary database
	_, err = db.Exec("INSERT INTO db.keys (address, key) VALUES(?, ?)", addr[:], sk.Seed())
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "constraint") {
			// If it was a constraint error, that means we already have the key.
			err = errKeyExists
			return
		}
		// Otherwise, return a generic database error
		err = errDatabase
		return
	}

	return
}

// ExportKey fetches the encrypted private key using the public key, decrypts
// it, verifies that it matches the passed public key, and returns it
func (ddbw *DuckDbWallet) ExportKey(addr types.Digest, pw []byte) (ed25519.PrivateKey, error) {
	if !ddbw.initialized {
		return nil, fmt.Errorf("wallet not initialized")
	}

	// Check the password
	err := ddbw.CheckPassword(pw)
	if err != nil {
		return nil, err
	}

	// Export the key
	return ddbw.fetchSecretKey(addr)
}

// GenerateKey generates a key from system entropy and imports it
func (ddbw *DuckDbWallet) GenerateKey(displayMnemonic bool) (addr types.Digest, err error) {
	if !ddbw.initialized {
		return addr, fmt.Errorf("wallet not initialized")
	}

	// The DuckDB wallet has SupportsMnemonicUX = false, meaning we don't know
	// how to show mnemonics to the user
	if displayMnemonic {
		err = errNoMnemonicUX
		return
	}

	// Open database
	db, err := ddbw.openAccountsDb()
	if err != nil {
		return
	}
	defer db.Close()

	// Fetch highest index from the accounts database
	var highestIndex uint64
	row := db.QueryRow("SELECT max_key_idx FROM db.info LIMIT 1")
	err = row.Scan(&highestIndex)
	if err != nil {
		return addr, err
	}

	// nextIndex he index of the next key we should generate
	nextIndex := highestIndex + 1

	var genPK ed25519.PublicKey
	var genSK ed25519.PrivateKey

	// We may have to bump nextIndex if the user has manually imported the next
	// key we were going to generate (thus we didn't see it in the search for the
	// highest-derived key above)
	for {
		// Honestly, if you could get 2**63 - 1 keys into this database, I'd be impressed
		if nextIndex == duckDbIntOverflow {
			err = errTooManyKeys
			return
		}

		// Decrypt the master derivation key stored in enclave into a local copy
		mdkBuf, mdkOpenErr := ddbw.masterDerivationKey.Open()
		if mdkOpenErr != nil {
			return addr, mdkOpenErr
		}
		defer mdkBuf.Destroy() // Destroy the copy when we return

		// Compute the secret key and public key for nextIndex
		genPK, genSK, err = extractKeyWithIndex(mdkBuf.Bytes(), nextIndex)
		if err != nil {
			return
		}

		// Convert the public key into an address
		addr = publicKeyToAddress(genPK)

		// Check that we don't already have this PK in the database
		var cnt int
		row := db.QueryRow("SELECT COUNT(1) FROM db.keys WHERE address = ? LIMIT 1", addr[:])
		err = row.Scan(&cnt)

		if cnt == 0 {
			// Good, key didn't exist. Break from loop
			break
		}

		// Uh oh, user already imported this key manually. Bump nextIndex
		nextIndex++
	}

	// Add new key into keys table
	_, err = db.Exec(
		"INSERT INTO db.keys (address, key, key_idx) VALUES(?, ?, ?)",
		addr[:], genSK.Seed(), nextIndex,
	)
	if err != nil {
		return
	}

	// Update `info` table with new max index (nextIndex)
	_, err = db.Exec("UPDATE db.info SET max_key_idx = ? ", nextIndex)
	if err != nil {
		return
	}

	return addr, nil
}

// DeleteKey deletes the key corresponding to the passed public key from the
// wallet
func (ddbw *DuckDbWallet) DeleteKey(addr types.Digest, pw []byte) (err error) {
	if !ddbw.initialized {
		return fmt.Errorf("wallet not initialized")
	}

	// Check the password
	err = ddbw.CheckPassword(pw)
	if err != nil {
		return
	}

	// Open database
	db, err := ddbw.openAccountsDb()
	if err != nil {
		return
	}
	defer db.Close()

	// Delete the key
	_, err = db.Exec("DELETE FROM db.keys WHERE address=?", addr[:])
	if err != nil {
		err = errDatabase
		return
	}

	return
}

// ImportMultisigAddr imports a multisig address, taking in version, threshold,
// and public keys
func (ddbw *DuckDbWallet) ImportMultisigAddr(version, threshold uint8, pks []ed25519.PublicKey) (addr types.Digest, err error) {
	if !ddbw.initialized {
		return addr, fmt.Errorf("wallet not initialized")
	}

	addr, err = kmdCrypto.MultisigAddrGen(version, threshold, pks)
	if err != nil {
		return
	}

	// Open database
	db, err := ddbw.openAccountsDb()
	if err != nil {
		return
	}
	defer db.Close()

	// Check if the multisig address is already stored in the file
	var cnt int
	row := db.QueryRow("SELECT COUNT(1) FROM db.msig_addrs WHERE address=? LIMIT 1", addr[:])
	err = row.Scan(&cnt)
	if err != nil {
		return
	}
	if cnt != 0 {
		return addr, fmt.Errorf("multisignature address already exists in wallet")
	}

	// Insert multisig address into database
	_, err = db.Exec("INSERT INTO db.msig_addrs (address, version, threshold, pks) VALUES (?, ?, ?, ?)",
		addr[:], version, threshold, msgpackEncode(pks))
	if err != nil {
		return
	}

	return
}

// LookupMultisigPreimage exports the preimage of a multisig address: version,
// threshold, public keys
func (ddbw *DuckDbWallet) LookupMultisigPreimage(addr types.Digest) (version, threshold uint8, pks []ed25519.PublicKey, err error) {
	// Open database
	db, err := ddbw.openAccountsDb()
	if err != nil {
		return
	}
	defer db.Close()

	var pksCandidate []ed25519.PublicKey
	var versionCandidate, thresholdCandidate int
	var pksBlob []byte

	row := db.QueryRow(
		"SELECT version, threshold, pks FROM db.msig_addrs WHERE address=? LIMIT 1",
		addr[:],
	)
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

// DeleteMultisigAddr deletes the multisig address and preimage from the wallet
func (ddbw *DuckDbWallet) DeleteMultisigAddr(addr types.Digest, pw []byte) (err error) {
	if !ddbw.initialized {
		return fmt.Errorf("wallet not initialized")
	}

	// Check the password
	err = ddbw.CheckPassword(pw)
	if err != nil {
		return
	}

	// Open database
	db, err := ddbw.openAccountsDb()
	if err != nil {
		return
	}
	defer db.Close()

	// Delete the key
	_, err = db.Exec("DELETE FROM db.msig_addrs WHERE address=?", addr[:])
	if err != nil {
		return
	}

	return
}

// ListMultisigAddrs lists the multisig addresses whose preimages we know
func (ddbw *DuckDbWallet) ListMultisigAddrs() (addrs []types.Digest, err error) {
	if !ddbw.initialized {
		return addrs, fmt.Errorf("wallet not initialized")
	}

	// Open database
	db, err := ddbw.openAccountsDb()
	if err != nil {
		return
	}
	defer db.Close()

	// Get all stored addresses
	rows, err := db.Query("SELECT address FROM db.msig_addrs")
	if err != nil {
		return
	}

	for rows.Next() {
		// We can't select directly into a types.Digest array, unfortunately.
		// Instead, we select into a slice of byte slices, and then convert each
		// of those slices into a types.Digest.

		var retrievedAddr []byte
		var tmp types.Digest

		err = rows.Scan(&retrievedAddr)
		if err != nil {
			return
		}

		copy(tmp[:], retrievedAddr)
		addrs = append(addrs, tmp)
	}

	return
}

// SignTransaction signs the passed transaction with the private key whose
// public key is provided, or if the provided public key is zero, inferring the
// required private key from the transaction itself
func (ddbw *DuckDbWallet) SignTransaction(tx types.Transaction, pk ed25519.PublicKey, pw []byte) (stx []byte, err error) {
	if !ddbw.initialized {
		return stx, fmt.Errorf("wallet not initialized")
	}

	// Check the password
	err = ddbw.CheckPassword(pw)
	if err != nil {
		return
	}

	// Fetch the required key
	var sk ed25519.PrivateKey
	if (slices.Equal(pk, ed25519.PublicKey{})) {
		sk, err = ddbw.fetchSecretKey(types.Digest(tx.Sender))
	} else {
		sk, err = ddbw.fetchSecretKey(types.Digest(pk))
	}
	if err != nil {
		return
	}

	// Sign the transaction with the required key
	_, stx, err = crypto.SignTransaction(sk, tx)

	return
}

// SignProgram signs the passed data for the src address
func (ddbw *DuckDbWallet) SignProgram(data []byte, src types.Digest, pw []byte) (sprog []byte, err error) {
	if !ddbw.initialized {
		return sprog, fmt.Errorf("wallet not initialized")
	}

	// Check the password
	err = ddbw.CheckPassword(pw)
	if err != nil {
		return
	}

	// Fetch the required key
	sk, err := ddbw.fetchSecretKey(types.Digest(src))
	if err != nil {
		return
	}

	sprog, err = crypto.SignBytes(sk, data)

	return
}

// MultisigSignTransaction starts a multisig signature or adds a signature to a
// partially signed multisig transaction signature of the passed transaction
// using the key
func (ddbw *DuckDbWallet) MultisigSignTransaction(tx types.Transaction, pk ed25519.PublicKey, partial types.MultisigSig, pw []byte, signer types.Digest) (sig types.MultisigSig, err error) {
	if !ddbw.initialized {
		return sig, fmt.Errorf("wallet not initialized")
	}

	// Check the password
	err = ddbw.CheckPassword(pw)
	if err != nil {
		return
	}

	if partial.Version == 0 && partial.Threshold == 0 && len(partial.Subsigs) == 0 {
		// We weren't given a partial multisig, so create a new one

		// Look up the preimage in the database
		var pks []ed25519.PublicKey
		var version, threshold uint8
		version, threshold, pks, err = ddbw.LookupMultisigPreimage(types.Digest(tx.Sender))
		if err != nil {
			return
		}

		// Fetch the required secret key (the secret key for the given public
		// key)
		var sk ed25519.PrivateKey
		sk, err = ddbw.fetchSecretKey(publicKeyToAddress(pk))
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
	sk, err := ddbw.fetchSecretKey(publicKeyToAddress(pk))
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
func (ddbw *DuckDbWallet) MultisigSignProgram(data []byte, src types.Digest, pk ed25519.PublicKey, partial types.MultisigSig, pw []byte) (sig types.MultisigSig, err error) {
	if !ddbw.initialized {
		return sig, fmt.Errorf("wallet not initialized")
	}

	// Check the password
	err = ddbw.CheckPassword(pw)
	if err != nil {
		return
	}

	if partial.Version == 0 && partial.Threshold == 0 && len(partial.Subsigs) == 0 {
		// We weren't given a partial multisig, so create a new one

		// Look up the preimage in the database
		var pks []ed25519.PublicKey
		var version, threshold uint8
		version, threshold, pks, err = ddbw.LookupMultisigPreimage(src)
		if err != nil {
			return
		}

		// Fetch the required secret key
		var sk ed25519.PrivateKey
		sk, err = ddbw.fetchSecretKey(publicKeyToAddress(pk))
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
	sk, err := ddbw.fetchSecretKey(publicKeyToAddress(pk))
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

// duckDbWalletMetadataFromPath accepts path to a directory for a wallet and
// returns a Metadata struct with information about it
func duckDbWalletMetadataFromPath(walletPath string) (metadata DuckDbWalletMetadata, err error) {
	// Read wallet's metadata.json
	metadataFileContents, err := os.ReadFile(filepath.Join(walletPath, DuckDbWalletMetadataFile))
	if err != nil {
		return
	}

	// Parse JSON contents
	metadata = DuckDbWalletMetadata{}
	err = json.Unmarshal(metadataFileContents, &metadata)
	if err != nil {
		return
	}

	return
}

/********** Wallet Driver Helpers **********/

// potentialWalletPaths lists paths to plausible wallet directories in the
// wallets directory. This means things that are directories and contain a
// `metadata.json` file.
func (ddbwd *DuckDbWalletDriver) potentialWalletPaths() (paths []string, err error) {
	// List all files and folders in the wallets directory
	wDir := ddbwd.walletsDir()
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
		_, err := os.Stat(filepath.Join(wDir, f.Name(), DuckDbWalletMetadataFile))
		if err != nil {
			continue
		}

		paths = append(paths, filepath.Join(wDir, f.Name()))
	}

	return
}

// maybeMakeWalletsDir tries to create the wallets directory if it doesn't
// already exist
func (ddbwd *DuckDbWalletDriver) maybeMakeWalletsDir() error {
	wDir := ddbwd.walletsDir()
	err := os.Mkdir(wDir, duckDbWalletsDirPermissions)
	if err != nil && !os.IsExist(err) {
		return fmt.Errorf("couldn't create wallets directory at %s: %v", wDir, err)
	}
	return nil
}

// walletsDir returns the wallet directory specified in the config, if there
// is one, otherwise it returns a subdirectory of the global kmd data dir
func (ddbwd *DuckDbWalletDriver) walletsDir() string {
	var walletDir string

	// Use the default if no wallet directory has been specified in the Parquet
	// wallet configuration
	if ddbwd.duckDbCfg.WalletsDir != "" {
		walletDir = filepath.FromSlash(ddbwd.duckDbCfg.WalletsDir)
	} else {
		walletDir = filepath.Join(filepath.FromSlash(ddbwd.globalCfg.DataDir), duckDbWalletsDirName)
	}

	// The each part of the directory path must be escaped to prevent the
	// directory name from being used for SQL injection
	dataDirParts := strings.Split(filepath.FromSlash(walletDir), string(filepath.Separator))
	var escapedDataDirParts []string
	for _, part := range dataDirParts {
		escapedDataDirParts = append(escapedDataDirParts, url.PathEscape(part))
	}

	return filepath.Join(escapedDataDirParts...)
}

// idToPath turns a wallet id into a path to the corresponding wallet directory
// to create
func (ddbwd *DuckDbWalletDriver) idToPath(id []byte) string {
	// wallet ID should already be safe, but filter it just in case
	safeID := disallowedDuckDbFilenameRegex.ReplaceAll(id, []byte(""))
	// The directory name for the wallet should be the wallet ID
	return filepath.Join(ddbwd.walletsDir(), string(safeID))
}

/********** Wallet Helpers **********/

// DecryptAndGetMasterKey fetches the master key (MEP) from wallet's metadata file
func (ddbw *DuckDbWallet) DecryptAndGetMasterKey(pw []byte) ([]byte, error) {
	retrievedMetadata, err := duckDbWalletMetadataFromPath(
		filepath.Join(ddbw.walletsPath, string(ddbw.id)),
	)
	if err != nil {
		return nil, err
	}

	mep, err := decryptBlobWithPassword(retrievedMetadata.MEPEncrypted, PTMasterKey, pw)
	if err != nil {
		return nil, err
	}

	return mep, nil
}

// decryptAndGetMasterDerivationKey fetches the MDK from the accounts database
// that is encrypted with the master encryption key
func (ddbw *DuckDbWallet) decryptAndGetMasterDerivationKey(pw []byte) ([]byte, error) {
	// Decrypt the password to get the master (encryption) key
	mek, err := ddbw.DecryptAndGetMasterKey(pw)
	if err != nil {
		return nil, err
	}

	// Open database
	db, err := sql.Open("duckdb", "")
	if err != nil {
		return nil, err
	}
	defer db.Close()

	acctsPath := filepath.Join(ddbw.walletsPath, ddbw.id, DuckDbWalletAcctsFile)

	// Open and decrypt accounts database file
	_, err = db.Exec(fmt.Sprintf(attachEncDuckDbSQL, acctsPath,
		base64.StdEncoding.EncodeToString(mek),
	))
	if err != nil {
		return nil, err
	}

	// Retrieve MDK from `info` table
	var mdk []byte
	row := db.QueryRow("SELECT mdk FROM db.info LIMIT 1")
	err = row.Scan(&mdk)
	if err != nil {
		return nil, err
	}

	return mdk, nil
}

// fetchSecretKey retrieves the private key for a given public key
func (ddbw *DuckDbWallet) fetchSecretKey(addr types.Digest) (sk ed25519.PrivateKey, err error) {
	keysPath := filepath.Join(ddbw.walletsPath, ddbw.id, DuckDbWalletAcctsFile)

	// Check if keys file exists
	_, err = os.Stat(keysPath)
	if err != nil {
		return nil, errKeyNotFound
	}

	// Open database
	db, err := ddbw.openAccountsDb()
	if err != nil {
		return
	}
	defer db.Close()

	var keyBlob []byte

	// Fetch the encrypted secret key from the database
	row := db.QueryRow("SELECT key FROM db.keys WHERE address=?", addr[:])
	err = row.Scan(&keyBlob)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errKeyNotFound
		}
		// Some unexpected error occurred
		return nil, err
	}

	skCandidate := ed25519.NewKeyFromSeed(keyBlob)
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

// openAccountsDb opens the encrypted accounts database file using the master
// encryption key
func (ddbw *DuckDbWallet) openAccountsDb() (db *sql.DB, err error) {
	// TODO: Add mutex lock to protect against data races

	// Open database
	db, err = sql.Open("duckdb", "")
	if err != nil {
		return
	}

	// Decrypt the master encryption key stored in enclave into a local copy
	mekBuf, err := ddbw.masterEncryptionKey.Open()
	if err != nil {
		return
	}
	defer mekBuf.Destroy() // Destroy the copy when we return

	acctsPath := filepath.Join(ddbw.walletsPath, ddbw.id, DuckDbWalletAcctsFile)

	// Open and decrypt accounts database file
	_, err = db.Exec(fmt.Sprintf(attachEncDuckDbSQL,
		acctsPath,
		base64.StdEncoding.EncodeToString(mekBuf.Bytes()),
	))
	if err != nil {
		return
	}

	return
}
