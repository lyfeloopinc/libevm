package core

// canExecuteTransaction is a convenience wrapper for calling the
// [params.RulesHooks.CanExecuteTransaction] hook.
func (st *StateTransition) canExecuteTransaction() error {
	vmCtx := st.evm.Context
	rules := st.evm.ChainConfig().Rules(vmCtx.BlockNumber, vmCtx.Random != nil, vmCtx.Time)
	return rules.Hooks().CanExecuteTransaction(st.msg.From, st.msg.To, st.state)
}
