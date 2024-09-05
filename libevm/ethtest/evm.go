package ethtest

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/require"
)

func NewZeroEVM(tb testing.TB) (*state.StateDB, *vm.EVM) {
	tb.Helper()

	sdb, err := state.New(common.Hash{}, state.NewDatabase(rawdb.NewMemoryDatabase()), nil)
	require.NoError(tb, err, "state.New()")

	return sdb, vm.NewEVM(
		vm.BlockContext{
			CanTransfer: func(_ vm.StateDB, _ common.Address, _ *uint256.Int) bool { return true },
			Transfer:    func(_ vm.StateDB, _, _ common.Address, _ *uint256.Int) {},
		},
		vm.TxContext{},
		sdb,
		&params.ChainConfig{},
		vm.Config{},
	)
}
