package vm_test

import (
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/libevm/ethtest"
	"github.com/ethereum/go-ethereum/libevm/hookstest"
	"github.com/ethereum/go-ethereum/params"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type vmHooksStub struct {
	replacement *vm.JumpTable
}

func (s *vmHooksStub) OverrideJumpTable(_ params.Rules, jt *vm.JumpTable) *vm.JumpTable {
	for op, instr := range s.replacement {
		if instr != nil {
			fmt.Println(op, instr)
			jt[op] = instr
		}
	}
	return jt
}

func TestOverrideJumpTable(t *testing.T) {
	hooks := &hookstest.Stub{
		OverrideJumpTableFlag: true,
	}
	hooks.Register(t)

	var called bool
	const opcode = 1

	vmHooks := &vmHooksStub{
		replacement: &vm.JumpTable{
			opcode: vm.NewOperation(
				func(pc *uint64, interpreter *vm.EVMInterpreter, callContext *vm.ScopeContext) ([]byte, error) {
					called = true
					return nil, nil
				},
				10, nil,
				0, 0,
				func(s *vm.Stack) (size uint64, overflow bool) {
					return 0, false
				},
			),
		},
	}
	vm.RegisterHooks(vmHooks)

	state, evm := ethtest.NewZeroEVM(t)

	rng := ethtest.NewPseudoRand(142857)
	contract := rng.Address()
	state.CreateAccount(contract)
	state.SetCode(contract, []byte{opcode})

	_, _, err := evm.Call(vm.AccountRef(rng.Address()), contract, []byte{}, 1e6, uint256.NewInt(0))
	require.NoError(t, err)
	assert.True(t, called)
}
