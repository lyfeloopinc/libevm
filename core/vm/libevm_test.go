package vm

import "github.com/ethereum/go-ethereum/core/tracing"

// The original RunPrecompiledContract was migrated to being a method on
// [evmCallArgs]. We need to replace it for use by regular geth tests.
func RunPrecompiledContract(p PrecompiledContract, input []byte, suppliedGas uint64, logger *tracing.Hooks) (ret []byte, remainingGas uint64, err error) {
	return (*evmCallArgs)(nil).RunPrecompiledContract(p, input, suppliedGas, logger)
}
