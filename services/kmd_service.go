package services

import (
	"crypto/ed25519"
	"duckysigner/kmd/config"
	"duckysigner/kmd/wallet"
	"duckysigner/kmd/wallet/driver"
	"encoding/base64"

	"github.com/algorand/go-algorand-sdk/v2/encoding/msgpack"
	"github.com/algorand/go-algorand-sdk/v2/mnemonic"
	"github.com/algorand/go-algorand-sdk/v2/types"
	"github.com/awnumar/memguard"
	logging "github.com/sirupsen/logrus"
)

// KMDService as a Wails binding allows for a Wails frontend to interact and
// manage KMD wallets
type KMDService struct {
	// Configuration for KMD
	Config config.KMDConfig
	// Path to the configuration file for KMD. The configuration file used only
	// if `Config` is not set
	ConfigPath string

	// If KMD configuration and drivers have been initialized
	kmdInitialized bool
}

// ListWallets gives the metadata of all available wallets
func (service *KMDService) ListWallets() ([]wallet.Metadata, error) {
	err := service.init()
	if err != nil {
		return nil, err
	}
	return driver.ListWalletMetadatas()
}

// CreateWallet creates a new SQLite wallet with the given walletName and password
func (service *KMDService) CreateWallet(walletName, password string) (newWalletData wallet.Metadata, err error) {
	err = service.init()
	if err != nil {
		return
	}

	// Generate new ID that will assigned to new wallet
	walletID, err := wallet.GenerateWalletID()
	if err != nil {
		return
	}

	// Get SQLite driver
	sqliteDriver, err := driver.FetchWalletDriver("sqlite")
	if err != nil {
		return
	}

	// Create wallet
	err = sqliteDriver.CreateWallet([]byte(walletName), walletID, []byte(password), types.MasterDerivationKey{})
	if err != nil {
		return
	}

	// Get the newly created wallet
	newWallet, err := sqliteDriver.FetchWallet(walletID)
	if err != nil {
		return
	}

	return newWallet.Metadata()
}

// GetWalletInfo attempts to get the information wallet with the given walletID
func (service *KMDService) GetWalletInfo(walletID string) (wallet.Metadata, error) {
	err := service.init()
	if err != nil {
		return wallet.Metadata{}, err
	}

	fetchedWallet, err := service.getWallet(walletID)
	if err != nil {
		return wallet.Metadata{}, err
	}

	return fetchedWallet.Metadata()
}

// RenameWallet renames the wallet with the given walletID to the given newName
func (service *KMDService) RenameWallet(walletID string, newName string, password string) error {
	err := service.init()
	if err != nil {
		return err
	}

	sqliteDriver, err := driver.FetchWalletDriver("sqlite")
	if err != nil {
		return err
	}

	return sqliteDriver.RenameWallet([]byte(newName), []byte(walletID), []byte(password))
}

// ExportWalletMnemonic exports the wallet with the given walletID by giving
// its mnemonic, which can be imported later
func (service *KMDService) ExportWalletMnemonic(walletID, password string) (string, error) {
	err := service.init()
	if err != nil {
		return "", err
	}

	// Get the wallet to be exported
	fetchedWallet, err := service.getWallet(walletID)
	if err != nil {
		return "", err
	}

	// Initialize wallet in order to temporarily and securely load master
	// derivation key (MDK) and master encryption key (MEK) into memory
	err = fetchedWallet.Init([]byte(password))
	if err != nil {
		return "", err
	}

	mdk, err := fetchedWallet.ExportMasterDerivationKey([]byte(password))
	if err != nil {
		return "", err
	}

	return mnemonic.FromMasterDerivationKey(mdk)
}

// ImportWalletMnemonic creates a new wallet with the given walletName and
// password using the given walletMnemonic
func (service *KMDService) ImportWalletMnemonic(walletMnemonic, walletName, password string) (newWalletData wallet.Metadata, err error) {
	err = service.init()
	if err != nil {
		return
	}

	// Generate new ID that will assigned to new wallet
	walletID, err := wallet.GenerateWalletID()
	if err != nil {
		return
	}

	// Get SQLite driver
	sqliteDriver, err := driver.FetchWalletDriver("sqlite")
	if err != nil {
		return
	}

	// Convert given mnemonic to a master derivation key (MDK)
	mdk, err := mnemonic.ToMasterDerivationKey(walletMnemonic)
	if err != nil {
		return
	}

	// Create new wallet using imported wallet MDK
	err = sqliteDriver.CreateWallet([]byte(walletName), walletID, []byte(password), mdk)
	if err != nil {
		return
	}

	// Get the newly created wallet
	newWallet, err := sqliteDriver.FetchWallet(walletID)
	if err != nil {
		return
	}

	return newWallet.Metadata()
}

// ListAccountsInWallet lists the addresses of all accounts within the wallet
// with the given walletID
func (service *KMDService) ListAccountsInWallet(walletID string) (acctAddrs []string, err error) {
	err = service.init()
	if err != nil {
		return
	}

	// Get the wallet to be exported
	fetchedWallet, err := service.getWallet(walletID)
	if err != nil {
		return
	}

	// Get the public keys of accounts stored in wallet
	pks, err := fetchedWallet.ListKeys()

	// Convert the list of public keys to a list of addresses
	for _, pk := range pks {
		acctAddrs = append(acctAddrs, types.Address(pk).String())
	}

	return
}

// GenerateWalletAccount generates an account for the wallet with the given
// walletID using its master derivation key (MDK)
func (service *KMDService) GenerateWalletAccount(walletID, password string) (string, error) {
	err := service.init()
	if err != nil {
		return "", err
	}

	// Get the wallet to be exported
	fetchedWallet, err := service.getWallet(walletID)
	if err != nil {
		return "", err
	}

	// Initialize wallet in order to temporarily and securely load master
	// derivation key (MDK) and master encryption key (MEK) into memory
	err = fetchedWallet.Init([]byte(password))
	if err != nil {
		return "", err
	}

	// Generate new public key using wallet MDK
	pk, err := fetchedWallet.GenerateKey(false)
	if err != nil {
		return "", err
	}

	return types.Address(pk).String(), nil
}

// CheckWalletPassword tests if the given password is the correct password for
// the wallet with the given wallet ID
func (service *KMDService) CheckWalletPassword(walletID, password string) error {
	err := service.init()
	if err != nil {
		return err
	}

	// Get wallet
	fetchedWallet, err := service.getWallet(walletID)
	if err != nil {
		return err
	}

	err = fetchedWallet.CheckPassword([]byte(password))
	if err != nil {
		return err
	}

	return nil
}

// RemoveAccountFromWallet removes the account with the given acctAddr from the
// wallet with the given walletID
func (service *KMDService) RemoveAccountFromWallet(acctAddr, walletID, password string) error {
	err := service.init()
	if err != nil {
		return err
	}

	// Get the wallet to be exported
	fetchedWallet, err := service.getWallet(walletID)
	if err != nil {
		return err
	}

	// Decode the given account address so it can be converted to a public key digest
	decodedAcctAddr, err := types.DecodeAddress(acctAddr)
	if err != nil {
		return err
	}

	return fetchedWallet.DeleteKey(types.Digest(decodedAcctAddr), []byte(password))
}

// ImportAccountIntoWallet imports the account with the given acctMnemonic into
// the wallet with the given walletID
func (service *KMDService) ImportAccountIntoWallet(acctMnemonic, walletID, password string) (acctAddr string, err error) {
	err = service.init()
	if err != nil {
		return
	}

	// Get the wallet to be exported
	fetchedWallet, err := service.getWallet(walletID)
	if err != nil {
		return
	}

	// Initialize wallet in order to temporarily and securely load master
	// derivation key (MDK) and master encryption key (MEK) into memory
	err = fetchedWallet.Init([]byte(password))
	if err != nil {
		return
	}

	// Convert mnemonic to private key
	sk, err := mnemonic.ToPrivateKey(acctMnemonic)
	if err != nil {
		return
	}

	// Import key into wallet
	pk, err := fetchedWallet.ImportKey(sk)
	if err != nil {
		return
	}

	acctAddr = types.Address(pk).String()

	return
}

// ExportAccountInWallet exports the account with the given acctAddr within the
// wallet with the given walletID by giving the mnemonic for the account
func (service *KMDService) ExportAccountInWallet(acctAddr, walletID, password string) (string, error) {
	err := service.init()
	if err != nil {
		return "", err
	}

	// Get the wallet to be exported
	fetchedWallet, err := service.getWallet(walletID)
	if err != nil {
		return "", err
	}

	// Initialize wallet in order to temporarily and securely load master
	// derivation key (MDK) and master encryption key (MEK) into memory
	err = fetchedWallet.Init([]byte(password))
	if err != nil {
		return "", err
	}

	// Decode the given account address so it can be converted to a public key digest
	decodedAcctAddr, err := types.DecodeAddress(acctAddr)
	if err != nil {
		return "", err
	}

	sk, err := fetchedWallet.ExportKey(types.Digest(decodedAcctAddr), []byte(password))
	if err != nil {
		return "", err
	}

	return mnemonic.FromPrivateKey(sk)
}

// SignTransaction signs the given Base64-encoded transaction using the account
// with the given public key, which should be within the wallet with the given
// ID. Returns the signed transaction as a Base64 string.
func (service *KMDService) SignTransaction(walletID, password string, txB64 string, acctAddr string) (stxB64 string, err error) {
	// Decode the Base64 transaction to bytes
	txnBytes, err := base64.StdEncoding.DecodeString(txB64)
	if err != nil {
		return
	}

	// Decode the transaction bytes to Transaction struct
	tx := types.Transaction{}
	err = msgpack.Decode(txnBytes, &tx)
	if err != nil {
		return
	}

	// Initialize KMD
	err = service.init()
	if err != nil {
		return
	}

	// Get the wallet
	fetchedWallet, err := service.getWallet(walletID)
	if err != nil {
		return
	}

	// Initialize wallet in order to temporarily and securely load master
	// derivation key (MDK) and master encryption key (MEK) into memory
	err = fetchedWallet.Init([]byte(password))
	if err != nil {
		return
	}

	// Convert account address (if given) to public key
	var acctPk []byte
	if acctAddr == "" {
		acctPk = ed25519.PublicKey{}
	} else {
		decodedAcctAddr, err2 := types.DecodeAddress(acctAddr)
		if err2 != nil {
			return "", err2
		}
		acctPk = decodedAcctAddr[:]
	}

	stxBytes, err := fetchedWallet.SignTransaction(tx, acctPk, []byte(password))
	if err != nil {
		return
	}

	// Sign transaction
	return base64.StdEncoding.EncodeToString(stxBytes), nil
}

// Start an interrupt handler that will clean up before exiting. This should be
// run right after creating a new service and before using it, which would often
// be in the beginning of a main() function.
func (service *KMDService) CatchInterrupt() {
	// Start interrupt handler for cleaning up the memory enclaves and locked
	// buffers
	memguard.CatchInterrupt()
}

// End sessions and release resources used by this KMD service
func (service *KMDService) CleanUp() {
	// Purge the MemGuard session
	defer memguard.Purge()
}

// func (service *KMDService) RemoveWallet(walletID string, password string) error {
// 	err := service.init()
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// init initializes KMD configuration and drivers for use if it has not been initialized
func (service *KMDService) init() (err error) {
	if service.kmdInitialized {
		return
	}

	kmdConfig := service.Config

	// Load or generate config if no config is set
	if kmdConfig.DriverConfig == (config.DriverConfig{}) {
		kmdConfig, err = config.LoadKMDConfig(service.ConfigPath)
		if err != nil {
			return
		}
	}

	// Create new logger because initializing wallet drivers requires it
	logger := logging.New()
	logger.SetLevel(logging.InfoLevel)

	// Initialize and fetch the wallet drivers
	err = driver.InitWalletDrivers(kmdConfig, logger)
	if err != nil {
		return
	}

	service.kmdInitialized = true

	return
}

// getWallet gets the wallet with the given walletID
func (service *KMDService) getWallet(walletID string) (wallet.Wallet, error) {
	sqliteDriver, err := driver.FetchWalletDriver("sqlite")
	if err != nil {
		return nil, err
	}

	return sqliteDriver.FetchWallet([]byte(walletID))
}
