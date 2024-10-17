package services

import (
	"duckysigner/kmd/config"
	"duckysigner/kmd/wallet"
	"duckysigner/kmd/wallet/driver"
	"sync"
	"time"

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
	// A singleton within a KMD service instance for the session for the wallet
	// that is currently open
	session *WalletSession
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

	// Initialize wallet in order to securely load master derivation key (MDK)
	// and master encryption key (MEK) into memory
	err = fetchedWallet.Init([]byte(password))

	service.sessionMutex.Lock()
	service.session = &WalletSession{
		wallet:     &fetchedWallet,
		password:   memguard.NewEnclave([]byte(password)),
		expiration: time.Now().Add(time.Duration(service.Config.SessionLifetimeSecs) * time.Second),
	}
	service.sessionMutex.Unlock()

	return
}

// Session is a getter that gets the current wallet session
func (service *KMDService) Session() (ws *WalletSession) {
	service.sessionMutex.RLock()
	ws = service.session
	service.sessionMutex.RUnlock()

	return
}

// EndSession ends and removes the current wallet session
func (service *KMDService) EndSession() {
	service.sessionMutex.Lock()
	service.session = &WalletSession{}
	service.sessionMutex.Unlock()
}

// RenewSession extends the current valid session by setting the expiration
// date-time to the current date-time plus the session lifetime specified in the
// KMD configuration. The expiration is always set to new expiration, even in
// the case of the new expiration being earlier than the old expiration.
func (service *KMDService) RenewSession() error {
	err := service.session.Check()
	if err != nil {
		return nil
	}

	service.session.expiration = time.Now().Add(time.Duration(service.Config.SessionLifetimeSecs) * time.Second)

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

	sqliteDriver, err := driver.FetchWalletDriver("sqlite")
	if err != nil {
		return err
	}

	return sqliteDriver.RenameWallet([]byte(newName), []byte(walletID), []byte(password))
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

// CatchInterrupt starts an interrupt handler that will clean up before exiting.
// This should be run right after creating a new service and before using it,
// which would often be in the beginning of a main() function.
func (service *KMDService) CatchInterrupt() {
	// Start interrupt handler for cleaning up the memory enclaves and locked
	// buffers
	memguard.CatchInterrupt()
}

// CleanUp end memory enclave sessions and release resources used by this KMD service
func (service *KMDService) CleanUp() {
	// Purge the MemGuard session
	defer memguard.Purge()
}

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
