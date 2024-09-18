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
// [JumpTable].
type OperationBuilder struct {
	Execute            OperationFunc
	ConstantGas        uint64
	DynamicGas         func(_ *EVM, _ *Contract, _ *Stack, _ *Memory, requestedMemorySize uint64) (uint64, error)
	MinStack, MaxStack int
	MemorySize         func(s *Stack) (size uint64, overflow bool)
}

// Build constructs the operation.
func (b OperationBuilder) Build() *operation {
	o := &operation{
		execute:     b.Execute.internal(),
		constantGas: b.ConstantGas,
		dynamicGas:  b.DynamicGas,
		minStack:    b.MinStack,
		maxStack:    b.MaxStack,
		memorySize:  b.MemorySize,
	}
	return o
}

// An OperationFunc is the execution function of a custom instruction.
type OperationFunc func(_ *OperationEnvironment, pc *uint64, _ *EVMInterpreter, _ *ScopeContext) ([]byte, error)

// An OperationEnvironment provides information about the context in which a
// custom instruction is being executed.
type OperationEnvironment struct {
	StateDB StateDB
}

// internal converts an exported [OperationFunc] into an un-exported
// [executionFunc] as required to build an [operation].
func (fn OperationFunc) internal() executionFunc {
	return func(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
		return fn(
			&OperationEnvironment{
				StateDB: interpreter.evm.StateDB,
			}, pc, interpreter, scope,
		)
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
