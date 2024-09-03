package core

func (st *StateTransition) canExecuteTransaction() error {
	vmCtx := st.evm.GetVMContext()
	rules := st.evm.ChainConfig().Rules(vmCtx.BlockNumber, vmCtx.Random != nil, vmCtx.Time)
	return rules.Hooks().CanExecuteTransaction(st.msg.From, st.msg.To, st.state)
}
