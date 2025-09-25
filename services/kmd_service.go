package services

import (
	"path/filepath"
	"sync"
	"time"

	"github.com/algorand/go-algorand-sdk/v2/mnemonic"
	"github.com/algorand/go-algorand-sdk/v2/types"
	"github.com/awnumar/memguard"
	logging "github.com/sirupsen/logrus"

	"duckysigner/internal/kmd/config"
	"duckysigner/internal/kmd/wallet"
	"duckysigner/internal/kmd/wallet/driver"
	ws "duckysigner/internal/wallet_session"
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
	// A singleton within a KMD service instance for the session for the wallet
	// that is currently open
	session *ws.WalletSession
	// Prevents possible data races when starting, ending or reading the same
	// session in different processes.
	sessionMutex sync.RWMutex
}

// StartSession starts a new wallet session by opening the wallet with the given
// ID and decrypting it using the given password. If a session already exists,
// it will be overwritten with this new wallet session.
func (service *KMDService) StartSession(walletID, password string) (err error) {
	err = service.init()
	if err != nil {
		return
	}

	fetchedWallet, err := service.getWallet(walletID)
	if err != nil {
		return
	}

	var walletDir string
	// Calculate the path of the wallet directory
	if service.Config.DriverConfig.ParquetWalletDriverConfig.WalletsDir != "" {
		walletDir = filepath.Join(service.Config.DriverConfig.ParquetWalletDriverConfig.WalletsDir, walletID)
	} else {
		walletDir = filepath.Join(service.Config.DataDir, "parquet_wallets", walletID)
	}

	// Initialize wallet in order to securely load master derivation key (MDK)
	// and master encryption key (MEK) into memory
	err = fetchedWallet.Init([]byte(password))

	service.sessionMutex.Lock()
	service.session = &ws.WalletSession{
		Wallet:   &fetchedWallet,
		Password: memguard.NewEnclave([]byte(password)),
		FilePath: walletDir,
	}
	service.session.SetExpiration(
		time.Now().Add(time.Duration(service.Config.SessionLifetimeSecs) * time.Second),
	)
	service.sessionMutex.Unlock()

	return
}

// Session is a getter that gets the current wallet session
//
// FOR THE BACKEND ONLY
func (service *KMDService) Session() (wSession *ws.WalletSession) {
	service.sessionMutex.RLock()
	wSession = service.session
	service.sessionMutex.RUnlock()

	return
}

// EndSession ends and removes the current wallet session
func (service *KMDService) EndSession() {
	service.sessionMutex.Lock()
	service.session = nil
	service.sessionMutex.Unlock()
}

// SessionIsForWallet gives whether the current session is for the wallet
// with the given ID
func (service *KMDService) SessionIsForWallet(walletID string) (bool, error) {
	if service.session == nil { // There is no session
		return false, nil
	}

	sessionWalletInfo, err := service.session.GetWalletInfo()
	if err != nil {
		return false, err
	}

	return string(sessionWalletInfo.ID) == walletID, nil
}

// RenewSession extends the current valid session by setting the expiration
// date-time to the current date-time plus the session lifetime specified in the
// KMD configuration. The expiration is always set to new expiration, even in
// the case of the new expiration being earlier than the old expiration.
func (service *KMDService) RenewSession() error {
	err := service.session.Check()
	if err != nil {
		return err
	}

	service.session.SetExpiration(
		time.Now().Add(time.Duration(service.Config.SessionLifetimeSecs) * time.Second),
	)

	return nil
}

// ListWallets gives the metadata of all available wallets
func (service *KMDService) ListWallets() ([]wallet.Metadata, error) {
	err := service.init()
	if err != nil {
		return nil, err
	}
	return driver.ListWalletMetadatas()
}

// CreateWallet creates a new Parquet wallet with the given walletName and
// password
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

	// Get Parquet driver
	pqDriver, err := driver.FetchWalletDriver("parquet")
	if err != nil {
		return
	}

	// Create wallet
	err = pqDriver.CreateWallet([]byte(walletName), walletID, []byte(password), types.MasterDerivationKey{})
	if err != nil {
		return
	}

	// Get the newly created wallet
	newWallet, err := pqDriver.FetchWallet(walletID)
	if err != nil {
		return
	}

	return newWallet.Metadata()
}

// GetWalletInfo gets the information of the wallet with the given walletID
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

	pqDriver, err := driver.FetchWalletDriver("parquet")
	if err != nil {
		return err
	}

	return pqDriver.RenameWallet([]byte(newName), []byte(walletID), []byte(password))
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

	// Get Parquet driver
	pqDriver, err := driver.FetchWalletDriver("parquet")
	if err != nil {
		return
	}

	// Convert given mnemonic to a master derivation key (MDK)
	mdk, err := mnemonic.ToMasterDerivationKey(walletMnemonic)
	if err != nil {
		return
	}

	// Create new wallet using imported wallet MDK
	err = pqDriver.CreateWallet([]byte(walletName), walletID, []byte(password), mdk)
	if err != nil {
		return
	}

	// Get the newly created wallet
	newWallet, err := pqDriver.FetchWallet(walletID)
	if err != nil {
		return
	}

	return newWallet.Metadata()
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

// func (service *KMDService) RemoveWallet(walletID string, password string) error {
// 	err := service.init()
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// CleanUp end memory enclave sessions and release resources used by this KMD
// service
//
// FOR THE BACKEND ONLY
func (service *KMDService) CleanUp() {
	// Purge the MemGuard session
	defer memguard.Purge()
	// XXX: May do other stuff (e.g. end db session properly) to clean up in the future
}

// init initializes KMD configuration and drivers for use if it has not been
// initialized
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
	pqDriver, err := driver.FetchWalletDriver("parquet")
	if err != nil {
		return nil, err
	}

	return pqDriver.FetchWallet([]byte(walletID))
}

/*
 * The following are shortcuts to make session functions accessible to the frontend
 */

// SessionCheck returns an error if the session is no longer valid
//
// FOR THE FRONTEND ONLY
func (service *KMDService) SessionCheck() error {
	return service.session.Check()
}

// SessionListAccounts lists the addresses of all accounts within the session
// wallet
//
// FOR THE FRONTEND ONLY
func (service *KMDService) SessionListAccounts() ([]string, error) {
	return service.session.ListAccounts()
}

// SessionGenerateAccount generates an account for the session wallet using its
// master derivation key (MDK)
//
// FOR THE FRONTEND ONLY
func (service *KMDService) SessionGenerateAccount() (string, error) {
	return service.session.GenerateAccount()
}

// SessionExportWallet exports the session wallet by returning its 25-word
// mnemonic. The password is always required for security reasons.
//
// FOR THE FRONTEND ONLY
func (service *KMDService) SessionExportWallet(password string) (string, error) {
	return service.session.ExportWallet(password)
}

// SessionImportAccount exports the account with the given acctAddr into the
// session wallet returning the account's 25-word mnemonic.
//
// FOR THE FRONTEND ONLY
func (service *KMDService) SessionImportAccount(acctMnemonic string) (string, error) {
	return service.session.ImportAccount(acctMnemonic)
}

// SessionExportAccount imports the account with the given acctMnemonic into the
// session wallet
//
// FOR THE FRONTEND ONLY
func (service *KMDService) SessionExportAccount(acctAddr, password string) (string, error) {
	return service.session.ExportAccount(acctAddr, password)
}

// SessionRemoveAccount removes the account with the given acctAddr from the
// session wallet
//
// FOR THE FRONTEND ONLY
func (service *KMDService) SessionRemoveAccount(acctAddr string) (err error) {
	return service.session.RemoveAccount(acctAddr)
}

// SessionSignTransaction signs the given Base64-encoded transaction
//
// FOR THE FRONTEND ONLY
func (service *KMDService) SessionSignTransaction(txB64, acctAddr string) (string, error) {
	return service.session.SignTransaction(txB64, acctAddr)
}
