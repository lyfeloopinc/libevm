package core

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/libevm"
	"github.com/ethereum/go-ethereum/libevm/ethtest"
	"github.com/ethereum/go-ethereum/libevm/hookstest"
	"github.com/stretchr/testify/require"
)

func TestCanExecuteTransaction(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	var (
		account common.Address
		slot    common.Hash
	)
	rng.Read(account[:])
	rng.Read(slot[:])

	makeErr := func(from common.Address, to *common.Address, val common.Hash) error {
		return fmt.Errorf("From: %v To: %v State: %v", from, to, val)
	}
	hooks := &hookstest.Stub{
		CanExecuteTransactionFn: func(from common.Address, to *common.Address, s libevm.StateReader) error {
			return makeErr(from, to, s.GetState(account, slot))
		},
	}
	hooks.RegisterForRules(t)

	var (
		from, to common.Address
		value    common.Hash
	)
	rng.Read(from[:])
	rng.Read(to[:])
	rng.Read(value[:])

	state, evm := ethtest.NewZeroEVM(t)
	state.SetState(account, slot, value)
	msg := &Message{
		From: from,
		To:   &to,
	}
	_, err := ApplyMessage(evm, msg, new(GasPool).AddGas(30e6))
	require.EqualError(t, err, makeErr(from, &to, value).Error())
}
