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

// An OperationBuilder is a factory for a new operations to include in a
// [JumpTable]. All of its fields are required.
type OperationBuilder[G interface {
	uint64 | func(_ *EVM, _ *Contract, _ *Stack, _ *Memory, requestedMemorySize uint64) (uint64, error)
}] struct {
	Execute            func(pc *uint64, interpreter *EVMInterpreter, callContext *ScopeContext) ([]byte, error)
	Gas                G
	MinStack, MaxStack int
	MemorySize         func(s *Stack) (size uint64, overflow bool)
}

type (
	// OperationBuilderConstantGas is the constant-gas version of an
	// OperationBuilder.
	OperationBuilderConstantGas = OperationBuilder[uint64]
	// OperationBuilderDynamicGas is the dynamic-gas version of an
	// OperationBuilder.
	OperationBuilderDynamicGas = OperationBuilder[func(_ *EVM, _ *Contract, _ *Stack, _ *Memory, requestedMemorySize uint64) (uint64, error)]
)

// Build constructs the operation.
func (b OperationBuilder[G]) Build() *operation {
	o := &operation{
		execute:    b.Execute,
		minStack:   b.MinStack,
		maxStack:   b.MaxStack,
		memorySize: b.MemorySize,
	}

	switch g := any(b.Gas).(type) {
	case uint64:
		o.constantGas = g
	case gasFunc:
		o.dynamicGas = g
	}
	return o
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
