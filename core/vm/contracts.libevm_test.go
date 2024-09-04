package vm

import (
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/libevm"
	"github.com/ethereum/go-ethereum/libevm/hookstest"
	"github.com/ethereum/go-ethereum/params"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/rand"
)

// The original RunPrecompiledContract was migrated to being a method on
// [evmCallArgs]. We need to replace it for use by regular geth tests.
func RunPrecompiledContract(p PrecompiledContract, input []byte, suppliedGas uint64) (ret []byte, remainingGas uint64, err error) {
	return (*evmCallArgs)(nil).RunPrecompiledContract(p, input, suppliedGas)
}

type precompileStub struct {
	requiredGas uint64
	returnData  []byte
}

func (s *precompileStub) RequiredGas([]byte) uint64  { return s.requiredGas }
func (s *precompileStub) Run([]byte) ([]byte, error) { return s.returnData, nil }

func TestPrecompileOverride(t *testing.T) {
	type test struct {
		name        string
		addr        common.Address
		requiredGas uint64
		stubData    []byte
	}

	const gasLimit = uint64(1e7)

	tests := []test{
		{
			name:        "arbitrary values",
			addr:        common.Address{'p', 'r', 'e', 'c', 'o', 'm', 'p', 'i', 'l', 'e'},
			requiredGas: 314159,
			stubData:    []byte("the return data"),
		},
	}

	rng := rand.New(rand.NewSource(42))
	for _, addr := range PrecompiledAddressesCancun {
		tests = append(tests, test{
			name:        fmt.Sprintf("existing precompile %v", addr),
			addr:        addr,
			requiredGas: rng.Uint64n(gasLimit),
			stubData:    addr[:],
		})
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hooks := &hookstest.Stub{
				PrecompileOverrides: map[common.Address]libevm.PrecompiledContract{
					tt.addr: &precompileStub{
						requiredGas: tt.requiredGas,
						returnData:  tt.stubData,
					},
				},
			}
			params.TestOnlyClearRegisteredExtras()
			hooks.RegisterForRules()

			t.Run(fmt.Sprintf("%T.Call([overridden precompile address = %v])", &EVM{}, tt.addr), func(t *testing.T) {
				gotData, gotGasLeft, err := newEVM(t).Call(AccountRef{}, tt.addr, nil, gasLimit, uint256.NewInt(0))
				require.NoError(t, err)
				assert.Equal(t, tt.stubData, gotData, "contract's return data")
				assert.Equal(t, gasLimit-tt.requiredGas, gotGasLeft, "gas left")
			})
		})
	}
}

func TestCanCreateContract(t *testing.T) {
	// We need to prove end-to-end plumbing of contract-creation addresses,
	// state, and any returned error. We therefore condition an error on a state
	// value being set, and that error contains the addresses.
	makeErr := func(cc *libevm.ContractCreation) error {
		return fmt.Errorf("Origin: %v Caller: %v Contract: %v", cc.Origin, cc.Caller, cc.Contract)
	}
	slot := common.Hash(crypto.Keccak256([]byte("slot")))
	value := common.Hash(crypto.Keccak256([]byte("value")))
	hooks := &hookstest.Stub{
		CanCreateContractFn: func(cc *libevm.ContractCreation, s libevm.StateReader) error {
			if s.GetState(common.Address{}, slot).Cmp(value) != 0 {
				return makeErr(cc)
			}
			return nil
		},
	}
	params.TestOnlyClearRegisteredExtras()
	hooks.RegisterForRules()

	origin := common.Address{'o', 'r', 'i', 'g', 'i', 'n'}
	caller := common.Address{'c', 'a', 'l', 'l', 'e', 'r'}
	create := crypto.CreateAddress(caller, 0)
	var (
		code []byte
		salt [32]byte
	)
	create2 := crypto.CreateAddress2(caller, salt, crypto.Keccak256(code))

	tests := []struct {
		name                          string
		setState                      bool
		wantCreateErr, wantCreate2Err error
	}{
		{
			name:           "no state => return error",
			setState:       false,
			wantCreateErr:  makeErr(&libevm.ContractCreation{Origin: origin, Caller: caller, Contract: create}),
			wantCreate2Err: makeErr(&libevm.ContractCreation{Origin: origin, Caller: caller, Contract: create2}),
		},
		{
			name:     "state set => no error",
			setState: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evm := newEVM(t)
			evm.TxContext.Origin = origin

			if tt.setState {
				sdb := evm.StateDB.(*state.StateDB)
				sdb.SetState(common.Address{}, slot, value)
			}

			methods := []struct {
				name     string
				deploy   func() ([]byte, common.Address, uint64, error)
				wantErr  error
				wantAddr common.Address
			}{
				{
					name: "Create",
					deploy: func() ([]byte, common.Address, uint64, error) {
						return evm.Create(AccountRef(caller), code, 1e6, uint256.NewInt(0))
					},
					wantErr:  tt.wantCreateErr,
					wantAddr: create,
				},
				{
					name: "Create2",
					deploy: func() ([]byte, common.Address, uint64, error) {
						return evm.Create2(AccountRef(caller), code, 1e6, uint256.NewInt(0), new(uint256.Int).SetBytes(salt[:]))
					},
					wantErr:  tt.wantCreate2Err,
					wantAddr: create2,
				},
			}

			for _, m := range methods {
				t.Run(m.name, func(t *testing.T) {
					_, gotAddr, _, err := m.deploy()
					if want := m.wantErr; want == nil {
						require.NoError(t, err)
						assert.Equal(t, m.wantAddr, gotAddr)
					} else {
						require.EqualError(t, err, want.Error())
					}
				})
			}
		})
	}
}

func newEVM(t *testing.T) *EVM {
	t.Helper()

	sdb, err := state.New(common.Hash{}, state.NewDatabase(rawdb.NewMemoryDatabase()), nil)
	require.NoError(t, err, "state.New()")

	return NewEVM(
		BlockContext{
			CanTransfer: func(_ StateDB, _ common.Address, _ *uint256.Int) bool { return true },
			Transfer:    func(_ StateDB, _, _ common.Address, _ *uint256.Int) {},
		},
		TxContext{},
		sdb,
		&params.ChainConfig{},
		Config{},
	)
}
