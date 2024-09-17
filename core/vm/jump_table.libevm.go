package vm

import "github.com/ethereum/go-ethereum/params"

// overrideJumpTable returns `libevmHooks.OverrideJumpTable(r,jt)“ i.f.f. the
// Rules' hooks indicate that it must, otherwise it echoes `jt` unchanged.
func overrideJumpTable(r params.Rules, jt *JumpTable) *JumpTable {
	if !r.Hooks().OverrideJumpTable() {
		return jt
	}
	// We don't check that libevmHooks is non-nil because the user has clearly
	// signified that they want an override. A nil-pointer dereference will
	// occur in tests whereas graceful degradation would frustrate the user with
	// a hard-to-find bug.
	return libevmHooks.OverrideJumpTable(r, jt)
}

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
