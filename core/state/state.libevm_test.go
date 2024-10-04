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

package state_test

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/state/snapshot"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethdb/memorydb"
	"github.com/ethereum/go-ethereum/libevm/ethtest"
	"github.com/ethereum/go-ethereum/triedb"
)

func TestGetSetExtra(t *testing.T) {
	types.TestOnlyClearRegisteredExtras()
	t.Cleanup(types.TestOnlyClearRegisteredExtras)
	payloads := types.RegisterExtras[[]byte]()

	rng := ethtest.NewPseudoRand(42)
	addr := rng.Address()
	nonce := rng.Uint64()
	balance := rng.Uint256()
	extra := rng.Bytes(8)

	views := newWithSnaps(t)
	stateDB := views.stateDB
	assert.Nilf(t, state.GetExtra(stateDB, payloads, addr), "state.GetExtra() returns zero-value %T if before account creation", extra)
	stateDB.CreateAccount(addr)
	stateDB.SetNonce(addr, nonce)
	stateDB.SetBalance(addr, balance)
	assert.Nilf(t, state.GetExtra(stateDB, payloads, addr), "state.GetExtra() returns zero-value %T if after account creation but before SetExtra()", extra)
	state.SetExtra(stateDB, payloads, addr, extra)
	assert.Equal(t, extra, state.GetExtra(stateDB, payloads, addr), "state.GetExtra() immediately after SetExtra()")

	root, err := stateDB.Commit(1, false) // arbitrary block number
	require.NoErrorf(t, err, "%T.Commit(1, false)", stateDB)
	require.NotEqualf(t, types.EmptyRootHash, root, "root hash returned by %T.Commit() is not the empty root", stateDB)

	t.Run(fmt.Sprintf("retrieve from %T", views.snaps), func(t *testing.T) {
		iter, err := views.snaps.AccountIterator(root, common.Hash{})
		require.NoErrorf(t, err, "%T.AccountIterator(...)", views.snaps)
		defer iter.Release()

		require.Truef(t, iter.Next(), "%T.Next() (i.e. at least one account)", iter)
		require.NoErrorf(t, iter.Error(), "%T.Error()", iter)

		t.Run("types.FullAccount()", func(t *testing.T) {
			got, err := types.FullAccount(iter.Account())
			require.NoErrorf(t, err, "types.FullAccount(%T.Account())", iter)

			want := &types.StateAccount{
				Nonce:    nonce,
				Balance:  balance,
				Root:     types.EmptyRootHash,
				CodeHash: types.EmptyCodeHash[:],
			}
			payloads.SetOnStateAccount(want, extra)

			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("types.FullAccount(%T.Account()) diff (-want +got):\n%s", iter, diff)
			}
		})

		require.Falsef(t, iter.Next(), "%T.Next() after first account (i.e. only one)", iter)
	})

	t.Run(fmt.Sprintf("retrieve from new %T", views.stateDB), func(t *testing.T) {
		stateDB, err := state.New(root, views.database, views.snaps)
		require.NoError(t, err, "state.New()")

		// triggers SlimAccount RLP decoding
		assert.Equalf(t, nonce, stateDB.GetNonce(addr), "%T.GetNonce()", stateDB)
		assert.Equalf(t, balance, stateDB.GetBalance(addr), "%T.GetBalance()", stateDB)
		assert.Equal(t, extra, state.GetExtra(stateDB, payloads, addr), "state.GetExtra()")
	})
}

// stateViews are different ways to access the same data.
type stateViews struct {
	stateDB  *state.StateDB
	snaps    *snapshot.Tree
	database state.Database
}

func newWithSnaps(t *testing.T) stateViews {
	t.Helper()
	empty := types.EmptyRootHash
	kvStore := memorydb.New()
	ethDB := rawdb.NewDatabase(kvStore)
	snaps, err := snapshot.New(
		snapshot.Config{
			CacheSize: 16, // Mb (arbitrary but non-zero)
		},
		kvStore,
		triedb.NewDatabase(ethDB, nil),
		empty,
	)
	require.NoError(t, err, "snapshot.New()")

	database := state.NewDatabase(ethDB)
	stateDB, err := state.New(empty, database, snaps)
	require.NoError(t, err, "state.New()")

	return stateViews{
		stateDB:  stateDB,
		snaps:    snaps,
		database: database,
	}
}
