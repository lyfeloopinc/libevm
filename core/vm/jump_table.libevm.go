package vm

import "github.com/ethereum/go-ethereum/params"

// NewOperation constructs a new operation for inclusion in a [JumpTable].
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

// Hooks are arbitrary configuration functions to modify default VM behaviour.
type Hooks interface {
	// OverrideJumpTable will only be called if
	// [params.RulesHooks.OverrideJumpTable] returns true. This allows for
	// recursive calling into [LookupInstructionSet].
	OverrideJumpTable(params.Rules, *JumpTable) *JumpTable
}

var libevmHooks Hooks

// RegisterHooks registers the Hooks. It is expected to be called in an `init()`
// function and MUST NOT be called more than once.
func RegisterHooks(h Hooks) {
	if libevmHooks != nil {
		panic("already registered")
	}
	libevmHooks = h
}
