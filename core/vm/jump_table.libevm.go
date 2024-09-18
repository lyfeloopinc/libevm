package vm

import (
	"github.com/ethereum/go-ethereum/libevm"
)

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
type OperationFunc func(_ Environment, pc *uint64, _ *EVMInterpreter, _ *ScopeContext) ([]byte, error)

// internal converts an exported [OperationFunc] into an un-exported
// [executionFunc] as required to build an [operation].
func (fn OperationFunc) internal() executionFunc {
	return func(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
		env := &environment{
			evm:      interpreter.evm,
			readOnly: interpreter.readOnly,
			addrs: libevm.AddressContext{
				Origin: interpreter.evm.Origin,
				Caller: scope.Contract.CallerAddress,
				Self:   scope.Contract.Address(),
			},
		}
		return fn(env, pc, interpreter, scope)
	}
}
