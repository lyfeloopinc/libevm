package vm_test

import (
	"fmt"
	"math/big"
	"reflect"
	"testing"

	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/libevm/ethtest"
	"github.com/ethereum/go-ethereum/libevm/hookstest"
	"github.com/ethereum/go-ethereum/params"
)

type vmHooksStub struct {
	replacement *vm.JumpTable
	overridden  bool
}

var _ vm.Hooks = (*vmHooksStub)(nil)

// OverrideJumpTable overrides all non-nil operations from s.replacement .
func (s *vmHooksStub) OverrideJumpTable(_ params.Rules, jt *vm.JumpTable) *vm.JumpTable {
	s.overridden = true
	for op, instr := range s.replacement {
		if instr != nil {
			fmt.Println(op, instr)
			jt[op] = instr
		}
	}
	return jt
}

// An opRecorder is an instruction that records its inputs.
type opRecorder struct {
	stateVal common.Hash
}

func (op *opRecorder) execute(env *vm.OperationEnvironment, pc *uint64, interpreter *vm.EVMInterpreter, scope *vm.ScopeContext) ([]byte, error) {
	op.stateVal = env.StateDB.GetState(scope.Contract.Address(), common.Hash{})
	return nil, nil
}

func TestOverrideJumpTable(t *testing.T) {
	override := new(bool)
	hooks := &hookstest.Stub{
		OverrideJumpTableFn: func() bool {
			return *override
		},
	}
	hooks.Register(t)

	const (
		opcode          = 1
		gasLimit uint64 = 1e6
	)
	rng := ethtest.NewPseudoRand(142857)
	gasCost := 1 + rng.Uint64n(gasLimit)
	spy := &opRecorder{}

	vmHooks := &vmHooksStub{
		replacement: &vm.JumpTable{
			opcode: vm.OperationBuilder{
				Execute:     spy.execute,
				ConstantGas: gasCost,
				MemorySize: func(s *vm.Stack) (size uint64, overflow bool) {
					return 0, false
				},
			}.Build(),
		},
	}
	vm.RegisterHooks(vmHooks)

	t.Run("LookupInstructionSet", func(t *testing.T) {
		rules := new(params.ChainConfig).Rules(big.NewInt(0), false, 0)

		for _, b := range []bool{false, true} {
			vmHooks.overridden = false

			*override = b
			_, err := vm.LookupInstructionSet(rules)
			require.NoError(t, err)
			require.Equal(t, b, vmHooks.overridden, "vm.Hooks.OverrideJumpTable() called i.f.f. params.RulesHooks.OverrideJumpTable() returns true")
		}
	})

	t.Run("EVMInterpreter", func(t *testing.T) {
		// We don't need to test the non-override case in EVMInterpreter because
		// that uses code shared with LookupInstructionSet. Here we only care
		// that the op gets executed as expected.
		*override = true
		state, evm := ethtest.NewZeroEVM(t)

		contract := rng.Address()
		state.CreateAccount(contract)
		state.SetCode(contract, []byte{opcode})
		value := rng.Hash()
		state.SetState(contract, common.Hash{}, value)

		_, gasRemaining, err := evm.Call(vm.AccountRef(rng.Address()), contract, []byte{}, gasLimit, uint256.NewInt(0))
		require.NoError(t, err, "evm.Call([contract with overridden opcode])")
		assert.Equal(t, gasLimit-gasCost, gasRemaining, "gas remaining")
		assert.Equal(t, spy.stateVal, value, "StateDB propagated")
	})
}

func TestOperationFieldCount(t *testing.T) {
	// The libevm OperationBuilder assumes that the 6 struct fields are the only
	// ones.
	op := vm.OperationBuilder{}.Build()
	require.Equalf(t, 6, reflect.TypeOf(*op).NumField(), "number of fields in %T struct", *op)
}
