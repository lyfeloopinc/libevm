package hookstest

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/libevm"
	"github.com/ethereum/go-ethereum/params"
)

// A Stub is a test double for [params.ChainConfigHooks] and
// [params.RulesHooks].
type Stub struct {
	PrecompileOverrides     map[common.Address]libevm.PrecompiledContract
	CanExecuteTransactionFn func(common.Address, *common.Address, libevm.StateReader) error
	CanCreateContractFn     func(*libevm.ContractCreation, libevm.StateReader) error
}

func (s *Stub) RegisterForRules() {
	params.RegisterExtras(params.Extras[params.NOOPHooks, Stub]{
		NewRules: func(_ *params.ChainConfig, _ *params.Rules, _ *params.NOOPHooks, blockNum *big.Int, isMerge bool, timestamp uint64) *Stub {
			return s
		},
	})
}

func (s Stub) PrecompileOverride(a common.Address) (libevm.PrecompiledContract, bool) {
	if len(s.PrecompileOverrides) == 0 {
		return nil, false
	}
	p, ok := s.PrecompileOverrides[a]
	return p, ok
}

func (s Stub) CanExecuteTransaction(from common.Address, to *common.Address, sr libevm.StateReader) error {
	if f := s.CanExecuteTransactionFn; f != nil {
		return f(from, to, sr)
	}
	return nil
}

func (s Stub) CanCreateContract(cc *libevm.ContractCreation, sr libevm.StateReader) error {
	if f := s.CanCreateContractFn; f != nil {
		return f(cc, sr)
	}
	return nil
}

var _ interface {
	params.ChainConfigHooks
	params.RulesHooks
} = Stub{}
