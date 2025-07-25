package wallet_session

import (
	"crypto/ed25519"
	"duckysigner/internal/kmd/wallet"
	"encoding/base64"
	"errors"
	"time"

	"github.com/algorand/go-algorand-sdk/v2/encoding/msgpack"
	"github.com/algorand/go-algorand-sdk/v2/mnemonic"
	"github.com/algorand/go-algorand-sdk/v2/types"
	"github.com/awnumar/memguard"
)

// WalletSession contains data about a session for an open wallet
type WalletSession struct {
	// The open and active Wallet for this session
	Wallet *wallet.Wallet
	// The wallet's Password (securely stored in a memory enclave)
	Password *memguard.Enclave
	// The date-time when this wallet session expires
	expiration time.Time
}

// Check returns an error if the session is no longer valid
func (session *WalletSession) Check() error {
	if time.Now().After(session.expiration) {
		return errors.New("wallet session is expired")
	}

	return nil
}

// Expiration returns the date-time when the session expires
func (session *WalletSession) Expiration() time.Time {
	return session.expiration
}

// SetExpiration sets the date-time when the session expires
func (session *WalletSession) SetExpiration(exp time.Time) {
	session.expiration = exp
}

// GetWalletInfo returns the information of the session wallet
func (session *WalletSession) GetWalletInfo() (wallet.Metadata, error) {
	if err := session.Check(); err != nil {
		return wallet.Metadata{}, err
	}

	return (*session.Wallet).Metadata()
}

// ExportWallet exports the session wallet by returning its 25-word mnemonic.
// The password is always required for security reasons.
func (session *WalletSession) ExportWallet(password string) (string, error) {
	if err := session.Check(); err != nil {
		return "", err
	}

	mdk, err := (*session.Wallet).ExportMasterDerivationKey([]byte(password))
	if err != nil {
		return "", err
	}

	return mnemonic.FromMasterDerivationKey(mdk)
}

// ListAccounts lists the addresses of all accounts within the session wallet
func (session *WalletSession) ListAccounts() (acctAddrs []string, err error) {
	if err = session.Check(); err != nil {
		return
	}

	// Get the public keys of accounts stored in wallet
	pks, err := (*session.Wallet).ListKeys()

	// Convert the list of public keys to a list of addresses
	for _, pk := range pks {
		acctAddrs = append(acctAddrs, types.Address(pk).String())
	}

	return
}

// GenerateAccount generates an account for the session wallet using its master
// derivation key (MDK). Returns the address of the generated account.
func (session *WalletSession) GenerateAccount() (string, error) {
	if err := session.Check(); err != nil {
		return "", err
	}

	// Generate new public key using wallet MDK
	pk, err := (*session.Wallet).GenerateKey(false)
	if err != nil {
		return "", err
	}

	return types.Address(pk).String(), nil
}

// ImportAccount imports the account with the given acctMnemonic into the
// session wallet. Returns the address of the imported account if the import was
// successful.
func (session *WalletSession) ImportAccount(acctMnemonic string) (string, error) {
	if err := session.Check(); err != nil {
		return "", err
	}

	// Convert mnemonic to private key
	sk, err := mnemonic.ToPrivateKey(acctMnemonic)
	if err != nil {
		return "", err
	}

	// Import key into wallet
	pk, err := (*session.Wallet).ImportKey(sk)
	if err != nil {
		return "", err
	}

	return types.Address(pk).String(), nil
}

// ExportAccount exports the account with the given acctAddr into the session
// wallet returning the account's 25-word mnemonic. The password is always
// required for security reasons.
func (session *WalletSession) ExportAccount(acctAddr, password string) (string, error) {
	if err := session.Check(); err != nil {
		return "", err
	}

	// Decode the given account address so it can be converted to a public key digest
	decodedAcctAddr, err := types.DecodeAddress(acctAddr)
	if err != nil {
		return "", err
	}

	sk, err := (*session.Wallet).ExportKey(types.Digest(decodedAcctAddr), []byte(password))
	if err != nil {
		return "", err
	}

	return mnemonic.FromPrivateKey(sk)
}

// RemoveAccount removes the account with the given acctAddr from the session
// wallet
func (session *WalletSession) RemoveAccount(acctAddr string) (err error) {
	if err = session.Check(); err != nil {
		return
	}

	// Decode the given account address so it can be converted to a public key digest
	decodedAcctAddr, err := types.DecodeAddress(acctAddr)

	// Retrieve password from memory enclave
	pwBuf, err := session.Password.Open()
	defer pwBuf.Destroy()

	return (*session.Wallet).DeleteKey(types.Digest(decodedAcctAddr), pwBuf.Bytes())
}

// SignTransaction signs the given Base64-encoded transaction. If specified, the
// account with the given address will be used to sign the transaction if it is
// in the session wallet. Otherwise, the account that is the "sender" of the
// transaction will be used to sign the transaction if it is in the wallet.
// Returns the signed transaction as a Base64 string if successful.
func (session *WalletSession) SignTransaction(txB64, acctAddr string) (stxB64 string, err error) {
	if err = session.Check(); err != nil {
		return
	}

	// Decode the Base64 transaction to bytes
	txnBytes, err := base64.StdEncoding.DecodeString(txB64)
	if err != nil {
		return
	}

	// Decode the transaction bytes to Transaction struct
	tx := types.Transaction{}
	if err = msgpack.Decode(txnBytes, &tx); err != nil {
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

	// Retrieve password from memory enclave
	pwBuf, err := session.Password.Open()
	if err != nil {
		return
	}
	defer pwBuf.Destroy()

	stxBytes, err := (*session.Wallet).SignTransaction(tx, acctPk, pwBuf.Bytes())
	if err != nil {
		return
	}

	// Sign transaction
	return base64.StdEncoding.EncodeToString(stxBytes), nil
}

// TODO: ChangePassword(password string) error
