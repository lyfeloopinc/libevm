package vm

import "github.com/ethereum/go-ethereum/params"

func NewOperation(
	execute func(pc *uint64, interpreter *EVMInterpreter, callContext *ScopeContext) ([]byte, error),
	constantGas uint64,
	dynamicGas func(e *EVM, c *Contract, s *Stack, m *Memory, u uint64) (uint64, error),
	minStack, maxStack int,
	memorySize func(s *Stack) (size uint64, overflow bool),
) *operation {
	return &operation{
		execute,
		constantGas,
		dynamicGas,
		minStack, maxStack,
		memorySize,
	}
}

type Hooks interface {
	OverrideJumpTable(params.Rules, *JumpTable) *JumpTable
}

var libevmHooks Hooks

func RegisterHooks(h Hooks) {
	if libevmHooks != nil {
		panic("already registered")
	}
	libevmHooks = h
}
