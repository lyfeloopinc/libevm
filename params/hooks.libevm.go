package params

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/libevm"
)

// ChainConfigHooks are required for all types registered as [Extras] for
// [ChainConfig] payloads.
type ChainConfigHooks interface{}

// RulesHooks are required for all types registered as [Extras] for [Rules]
// payloads.
type RulesHooks interface {
	// PrecompileOverride signals whether or not the EVM interpreter MUST
	// override its treatment of the address when deciding if it is a
	// precompiled contract. If PrecompileOverride returns `true` then the
	// interpreter will treat the address as a precompile i.f.f the
	// [PrecompiledContract] is non-nil. If it returns `false` then the default
	// precompile behaviour is honoured.
	PrecompileOverride(Rules, common.Address) (_ libevm.PrecompiledContract, override bool)
}

// hooksFromChainConfig returns the registered hooks, or [NOOPHooks] if none
// were registered.
func hooksFromChainConfig(c *ChainConfig) ChainConfigHooks {
	if e := registeredExtras; e != nil {
		return e.getter.hooksFromChainConfig(c)
	}
	return NOOPHooks{}
}

// hooksFromRules returns the registered hooks, or [NOOPHooks] if none were
// registered.
func hooksFromRules(r *Rules) RulesHooks {
	if e := registeredExtras; e != nil {
		return e.getter.hooksFromRules(r)
	}
	return NOOPHooks{}
}

// PrecompileOverride is a wrapper for the equivalent method of the registered
// [RulesHooks] carried by the Rules. If no [Extras] were registered, a
// [NOOPHooks] is used instead, such that the default precompile behaviour is
// exhibited.
func (r Rules) PrecompileOverride(addr common.Address) (libevm.PrecompiledContract, bool) {
	return hooksFromRules(&r).PrecompileOverride(r, addr)
}

// NOOPHooks implements both [ChainConfigHooks] and [RulesHooks] such that every
// hook is a no-op. This allows it to be returned instead of a nil interface,
// which would otherwise require every usage site to perform a nil check. It can
// also be embedded in structs that only wish to implement a sub-set of hooks.
// Use of a NOOPHooks is equivalent to default Ethereum behaviour.
type NOOPHooks struct{}

var _ interface {
	ChainConfigHooks
	RulesHooks
} = NOOPHooks{}

// PrecompileOverride instructs the EVM interpreter to use the default
// precompile behaviour.
func (NOOPHooks) PrecompileOverride(Rules, common.Address) (libevm.PrecompiledContract, bool) {
	return nil, false
}
