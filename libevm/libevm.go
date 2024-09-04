package libevm

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/uint256"
)

// PrecompiledContract is an exact copy of vm.PrecompiledContract, mirrored here
// for instances where importing that package would result in a circular
// dependency.
type PrecompiledContract interface {
	RequiredGas(input []byte) uint64
	Run(input []byte) ([]byte, error)
}

// StateReader is a subset of vm.StateDB, exposing only methods that read from
// but do not modify state. See method comments in vm.StateDB, which aren't
// copied here as they risk becoming outdated.
type StateReader interface {
	GetBalance(common.Address) *uint256.Int
	GetNonce(common.Address) uint64

	GetCodeHash(common.Address) common.Hash
	GetCode(common.Address) []byte
	GetCodeSize(common.Address) int

	GetRefund() uint64

	GetCommittedState(common.Address, common.Hash) common.Hash
	GetState(common.Address, common.Hash) common.Hash

	GetTransientState(addr common.Address, key common.Hash) common.Hash

	HasSelfDestructed(common.Address) bool

	Exist(common.Address) bool
	Empty(common.Address) bool

	AddressInAccessList(addr common.Address) bool
	SlotInAccessList(addr common.Address, slot common.Hash) (addressOk bool, slotOk bool)
}

type ContractCreation struct {
	Origin, Caller, Contract common.Address
}
