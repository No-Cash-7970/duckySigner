package services

import (
	"duckysigner/kmd/config"
	"duckysigner/kmd/wallet"
	"duckysigner/kmd/wallet/driver"

	"github.com/algorand/go-algorand-sdk/v2/mnemonic"
	"github.com/algorand/go-algorand-sdk/v2/types"
	"github.com/algorand/go-algorand/logging"
)

// KMDService as a Wails binding allows for a Wails frontend to interact and
// manage KMD wallets
type KMDService struct {
	kmdInitialized bool
	Config         config.KMDConfig
	ConfigPath     string
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

	// TODO: Use some mechanism to prevent MDK in memory from reaching the disk (e.g. mlock, memory enclave)
	// DANGEROUS: Initialize wallet in order to temporarily load MDK into memory before export
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

	// TODO: Use some mechanism to prevent MDK in memory from reaching the disk (e.g. mlock, memory enclave)
	// DANGEROUS: Initialize wallet in order to temporarily load MDK into memory before export
	err = fetchedWallet.Init([]byte(password))
	if err != nil {
		return "", err
	}

	// Generate new public key using wallet master derivation key (MDK)
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

	// TODO: Use some mechanism to prevent MDK in memory from reaching the disk (e.g. mlock, memory enclave)
	// DANGEROUS: Initialize wallet in order to temporarily load MDK into memory before export
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

	// TODO: Use some mechanism to prevent MDK in memory from reaching the disk (e.g. mlock, memory enclave)
	// DANGEROUS: Initialize wallet in order to temporarily load MDK into memory before export
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
	logger := logging.NewLogger()
	logger.SetLevel(logging.Info)

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
