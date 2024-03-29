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

// XXX: Modified from https://github.com/algorand/go-algorand/tree/c2d7047585f6109d866ebaf9fca0ee7490b16c6a/crypto/multisig.go

package crypto2

import (
	"crypto/ed25519"

	"github.com/algorand/go-algorand-sdk/v2/crypto"
	"github.com/algorand/go-algorand-sdk/v2/types"
)

// MultisigAddrGen identifies the exact group, version,
// and devices (Public keys) that it requires to sign
// Hash("MultisigAddr" || version uint8 || threshold uint8 || PK1 || PK2 || ...)
func MultisigAddrGen(version, threshold uint8, pks []ed25519.PublicKey) (addr types.Digest, err error) {
	// Create multisig account data using multisig settings
	ma := crypto.MultisigAccount{
		Version:   version,
		Threshold: threshold,
		Pks:       pks,
	}

	// Use multisig account to get multisig address
	var maddr types.Address
	maddr, err = ma.Address()
	if err != nil {
		return
	}

	// Convert address to a digest
	addr = types.Digest(maddr[:])

	return
}
