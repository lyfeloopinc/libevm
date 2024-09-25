// Copyright 2024 the libevm authors.
//
// The libevm additions to go-ethereum are free software: you can redistribute
// them and/or modify them under the terms of the GNU Lesser General Public License
// as published by the Free Software Foundation, either version 3 of the License,
// or (at your option) any later version.
//
// The libevm additions are distributed in the hope that they will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU Lesser
// General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see
// <http://www.gnu.org/licenses/>.

package vm

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/libevm"
	"github.com/ethereum/go-ethereum/params"
)

// An Environment provides information about the context in which a precompiled
// contract or an instruction is being executed.
type Environment interface {
	ChainConfig() *params.ChainConfig
	Rules() params.Rules
	ReadOnly() bool
	// StateDB will be non-nil i.f.f !ReadOnly().
	StateDB() StateDB
	// ReadOnlyState will always be non-nil.
	ReadOnlyState() libevm.StateReader
	Addresses() *libevm.AddressContext

	BlockHeader() (types.Header, error)
	BlockNumber() *big.Int
	BlockTime() uint64
}

//
// ****** SECURITY ******
//
// If you are updating Environment to provide the ability to call back into
// another contract, you MUST revisit the evmCallArgs.forceReadOnly flag.
//
// It is possible that forceReadOnly is true but evm.interpreter.readOnly is
// false. This is safe for now, but may not be if recursive calling *from* a
// precompile is enabled.
//
// ****** SECURITY ******

var _ Environment = (*environment)(nil)

type environment struct {
	evm      *EVM
	readOnly bool
	addrs    libevm.AddressContext
}

func (e *environment) ChainConfig() *params.ChainConfig  { return e.evm.chainConfig }
func (e *environment) Rules() params.Rules               { return e.evm.chainRules }
func (e *environment) ReadOnly() bool                    { return e.readOnly }
func (e *environment) ReadOnlyState() libevm.StateReader { return e.evm.StateDB }
func (e *environment) Addresses() *libevm.AddressContext { return &e.addrs }
func (e *environment) BlockNumber() *big.Int             { return new(big.Int).Set(e.evm.Context.BlockNumber) }
func (e *environment) BlockTime() uint64                 { return e.evm.Context.Time }

func (e *environment) StateDB() StateDB {
	if e.ReadOnly() {
		return nil
	}
	return e.evm.StateDB
}

func (e *environment) BlockHeader() (types.Header, error) {
	hdr := e.evm.Context.Header
	if hdr == nil {
		// Although [core.NewEVMBlockContext] sets the field and is in the
		// typical hot path (e.g. miner), there are other ways to create a
		// [vm.BlockContext] (e.g. directly in tests) that may result in no
		// available header.
		return types.Header{}, fmt.Errorf("nil %T in current %T", hdr, e.evm.Context)
	}
	return *hdr, nil
}
