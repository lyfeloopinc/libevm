package vm

import "github.com/ethereum/go-ethereum/params"

// RegisterHooks registers the Hooks. It is expected to be called in an `init()`
// function and MUST NOT be called more than once.
func RegisterHooks(h Hooks) {
	if libevmHooks != nil {
		panic("already registered")
	}
	libevmHooks = h
}

var libevmHooks Hooks

// Hooks are arbitrary configuration functions to modify default VM behaviour.
type Hooks interface {
	OverrideJumpTable(params.Rules, *JumpTable) *JumpTable
}

// overrideJumpTable returns `libevmHooks.OverrideJumpTable(r,jt)` i.f.f. Hooks
// have been registered.
func overrideJumpTable(r params.Rules, jt *JumpTable) *JumpTable {
	if libevmHooks == nil {
		return jt
	}
	return libevmHooks.OverrideJumpTable(r, jt)
}
