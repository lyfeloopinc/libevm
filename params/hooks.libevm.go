package params

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/libevm"
)

// ChainConfigHooks are...
type ChainConfigHooks interface{}

// RulesHooks are...
type RulesHooks interface {
	PrecompileOverride(Rules, common.Address) (_ libevm.PrecompiledContract, override bool)
}

func (e *extraConstructors) hooksFromChainConfig(c *ChainConfig) ChainConfigHooks {
	if e == nil {
		return NoopHooks{}
	}
	return e.getter.hooksFromChainConfig(c)
}

func (e *extraConstructors) hooksFromRules(r *Rules) RulesHooks {
	if e == nil {
		return NoopHooks{}
	}
	return e.getter.hooksFromRules(r)
}

// PrecompileOverride...
func (r Rules) PrecompileOverride(addr common.Address) (libevm.PrecompiledContract, bool) {
	return registeredExtras.hooksFromRules(&r).PrecompileOverride(r, addr)
}

// NoopHooks implement both [ChainConfigHooks] and [RulesHooks] such that every
// hook is a no-op. This allows it to be returned instead of nil, which would
// require every usage site to perform a nil check. It can also be embedded in
// structs that only wish to implement a sub-set of hooks.
type NoopHooks struct{}

var _ interface {
	ChainConfigHooks
	RulesHooks
} = NoopHooks{}

// PrecompileOverride always returns `nil, false`; the `false` indicates that
// the `nil` should be ignored.
func (NoopHooks) PrecompileOverride(Rules, common.Address) (libevm.PrecompiledContract, bool) {
	return nil, false
}
