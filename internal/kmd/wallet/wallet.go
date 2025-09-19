// Copyright (C) 2019-2024 Algorand, Inc.
// This file is part of go-algorand
//
// go-algorand is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// go-algorand is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with go-algorand.  If not, see <https://www.gnu.org/licenses/>.

// XXX: Modified from https://github.com/algorand/go-algorand/tree/c2d7047585f6109d866ebaf9fca0ee7490b16c6a/daemon/kmd/wallet/wallet.go

package wallet

import (
	"crypto/ed25519"
	"crypto/rand"
	"fmt"

	"github.com/algorand/go-algorand-sdk/v2/types"
)

const (
	walletIDBytes = 16
)

// Wallet represents the interface that any wallet technology must satisfy in
// order to be used with KMD. Wallets start in a locked state until they are
// initialized with Init.
type Wallet interface {
	Init(pw []byte) error
	CheckPassword(pw []byte) error
	ExportMasterDerivationKey(pw []byte) (types.MasterDerivationKey, error)

	Metadata() (Metadata, error)

	ListKeys() ([]types.Digest, error)

	ImportKey(sk ed25519.PrivateKey) (types.Digest, error)
	ExportKey(pk types.Digest, pw []byte) (ed25519.PrivateKey, error)
	GenerateKey(displayMnemonic bool) (types.Digest, error)
	DeleteKey(pk types.Digest, pw []byte) error

	ImportMultisigAddr(version, threshold uint8, pks []ed25519.PublicKey) (types.Digest, error)
	LookupMultisigPreimage(types.Digest) (version, threshold uint8, pks []ed25519.PublicKey, err error)
	ListMultisigAddrs() (addrs []types.Digest, err error)
	DeleteMultisigAddr(addr types.Digest, pw []byte) error

	SignTransaction(tx types.Transaction, pk ed25519.PublicKey, pw []byte) ([]byte, error)

	MultisigSignTransaction(tx types.Transaction, pk ed25519.PublicKey, partial types.MultisigSig, pw []byte, signer types.Digest) (types.MultisigSig, error)

	SignProgram(program []byte, src types.Digest, pw []byte) ([]byte, error)
	MultisigSignProgram(program []byte, src types.Digest, pk ed25519.PublicKey, partial types.MultisigSig, pw []byte) (types.MultisigSig, error)

	DecryptAndGetMasterKey(pw []byte) ([]byte, error)
}

// Metadata represents high-level information about a wallet, like its name, id
// and what operations it supports
type Metadata struct {
	ID                    []byte
	Name                  []byte
	DriverName            string
	DriverVersion         uint32
	SupportsMnemonicUX    bool
	SupportsMasterKey     bool
	SupportedTransactions []types.TxType
}

// GenerateWalletID generates a random hex wallet ID
func GenerateWalletID() ([]byte, error) {
	bytes := make([]byte, walletIDBytes)
	_, err := rand.Read(bytes)
	if err != nil {
		return []byte(""), err
	}
	return []byte(fmt.Sprintf("%x", bytes)), nil
}
