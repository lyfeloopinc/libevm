package vm_test

import (
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/libevm"
	"github.com/ethereum/go-ethereum/libevm/ethtest"
	"github.com/ethereum/go-ethereum/libevm/hookstest"
	"github.com/ethereum/go-ethereum/params"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/rand"
)

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
	for _, addr := range vm.PrecompiledAddressesCancun {
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
			hooks.RegisterForRules(t)

			t.Run(fmt.Sprintf("%T.Call([overridden precompile address = %v])", &vm.EVM{}, tt.addr), func(t *testing.T) {
				_, evm := ethtest.NewZeroEVM(t)
				gotData, gotGasLeft, err := evm.Call(vm.AccountRef{}, tt.addr, nil, gasLimit, uint256.NewInt(0))
				require.NoError(t, err)
				assert.Equal(t, tt.stubData, gotData, "contract's return data")
				assert.Equal(t, gasLimit-tt.requiredGas, gotGasLeft, "gas left")
			})
		})
	}
}

func TestNewStatefulPrecompile(t *testing.T) {
	var (
		caller, precompile common.Address
		input              = make([]byte, 8)
		slot, value        common.Hash
	)
	rng := rand.New(rand.NewSource(314159))
	rng.Read(caller[:])
	rng.Read(precompile[:])
	rng.Read(input[:])
	rng.Read(slot[:])
	rng.Read(value[:])

	const gasLimit = 1e6
	gasCost := rng.Uint64n(gasLimit)

	makeOutput := func(caller, self common.Address, input []byte, stateVal common.Hash) []byte {
		return []byte(fmt.Sprintf(
			"Caller: %v Precompile: %v State: %v Input: %#x",
			caller, self, stateVal, input,
		))
	}
	hooks := &hookstest.Stub{
		PrecompileOverrides: map[common.Address]libevm.PrecompiledContract{
			precompile: vm.NewStatefulPrecompile(
				func(state vm.StateDB, _ *params.Rules, caller, self common.Address, input []byte) ([]byte, error) {
					return makeOutput(caller, self, input, state.GetState(precompile, slot)), nil
				},
				func(b []byte) uint64 {
					return gasCost
				},
			),
		},
	}
	hooks.RegisterForRules(t)

	state, evm := ethtest.NewZeroEVM(t)
	state.SetState(precompile, slot, value)
	wantReturnData := makeOutput(caller, precompile, input, value)
	wantGasLeft := gasLimit - gasCost

	gotReturnData, gotGasLeft, err := evm.Call(vm.AccountRef(caller), precompile, input, gasLimit, uint256.NewInt(0))
	require.NoError(t, err)
	assert.Equal(t, wantReturnData, gotReturnData)
	assert.Equal(t, wantGasLeft, gotGasLeft)
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
	hooks.RegisterForRules(t)

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
			stateDB, evm := ethtest.NewZeroEVM(t)
			evm.TxContext.Origin = origin

			if tt.setState {
				stateDB.SetState(common.Address{}, slot, value)
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
						return evm.Create(vm.AccountRef(caller), code, 1e6, uint256.NewInt(0))
					},
					wantErr:  tt.wantCreateErr,
					wantAddr: create,
				},
				{
					name: "Create2",
					deploy: func() ([]byte, common.Address, uint64, error) {
						return evm.Create2(vm.AccountRef(caller), code, 1e6, uint256.NewInt(0), new(uint256.Int).SetBytes(salt[:]))
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
