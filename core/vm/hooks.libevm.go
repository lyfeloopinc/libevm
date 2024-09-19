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
	// OverrideJumpTable will only be called if
	// [params.RulesHooks.OverrideJumpTable] returns true. This allows for
	// recursive calling into [LookupInstructionSet].
	OverrideJumpTable(params.Rules, *JumpTable) *JumpTable
}

// overrideJumpTable returns `libevmHooks.OverrideJumpTable(r,jt)“ i.f.f. the
// Rules' hooks indicate that it must, otherwise it echoes `jt` unchanged.
func overrideJumpTable(r params.Rules, jt *JumpTable) *JumpTable {
	if libevmHooks == nil {
		return jt
	}
	return libevmHooks.OverrideJumpTable(r, jt)
}
