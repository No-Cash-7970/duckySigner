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

// XXX: Modified from https://github.com/algorand/go-algorand/tree/c2d7047585f6109d866ebaf9fca0ee7490b16c6a/daemon/kmd/wallet/driver/util.go

package driver

import (
	"crypto/ed25519"
	"os"

	"github.com/algorand/go-algorand-sdk/v2/types"
)

func publicKeyToAddress(pk ed25519.PublicKey) (addr types.Digest) {
	copy(addr[:], pk[:])
	return
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
