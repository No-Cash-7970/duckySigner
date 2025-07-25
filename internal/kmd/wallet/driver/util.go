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

	"github.com/algorand/go-algorand-sdk/v2/types"
)

func publicKeyToAddress(pk ed25519.PublicKey) (addr types.Digest) {
	copy(addr[:], pk[:])
	return
}
