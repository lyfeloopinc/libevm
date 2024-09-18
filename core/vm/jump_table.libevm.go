package vm

import "github.com/ethereum/go-ethereum/libevm"

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
	ReadOnly bool
	// StateDB will be non-nil i.f.f !ReadOnly.
	StateDB StateDB
	// ReadOnlyState will always be non-nil.
	ReadOnlyState libevm.StateReader
}

// internal converts an exported [OperationFunc] into an un-exported
// [executionFunc] as required to build an [operation].
func (fn OperationFunc) internal() executionFunc {
	return func(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
		env := &OperationEnvironment{
			ReadOnly:      interpreter.readOnly,
			ReadOnlyState: interpreter.evm.StateDB,
		}
		if !env.ReadOnly {
			env.StateDB = interpreter.evm.StateDB
		}
		return fn(env, pc, interpreter, scope)
	}
}
