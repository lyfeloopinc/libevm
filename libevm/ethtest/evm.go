// Package ethtest provides utility functions for use in testing
// Ethereum-related functionality.
package ethtest

import (
	"testing"

	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/state/snapshot"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/ethdb/memorydb"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/triedb"
	"github.com/stretchr/testify/require"
)

// NewZeroEVM returns a new EVM backed by a [rawdb.NewMemoryDatabase]; all other
// arguments to [vm.NewEVM] are the zero values of their respective types,
// except for the use of [core.CanTransfer] and [core.Transfer] instead of nil
// functions.
func NewZeroEVM(tb testing.TB) (*state.StateDB, *vm.EVM) {
	tb.Helper()

	empty := types.EmptyRootHash
	kvStore := memorydb.New()
	ethDB := rawdb.NewDatabase(kvStore)
	tdb := triedb.NewDatabase(ethDB, &triedb.Config{})
	snaps, err := snapshot.New(
		snapshot.Config{
			CacheSize: 16, // Mb (arbitrary but non-zero)
		},
		kvStore, tdb, empty,
	)
	require.NoError(tb, err, "snapshot.New()")

	sdb, err := state.New(empty, state.NewDatabase(tdb, snaps))
	require.NoError(tb, err, "state.New()")

	return sdb, vm.NewEVM(
		vm.BlockContext{
			CanTransfer: core.CanTransfer,
			Transfer:    core.Transfer,
		},
		vm.TxContext{},
		sdb,
		&params.ChainConfig{},
		vm.Config{},
	)
}
