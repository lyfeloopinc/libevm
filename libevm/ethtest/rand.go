package ethtest

import (
	"github.com/ethereum/go-ethereum/common"
	"golang.org/x/exp/rand"
)

type Rand struct {
	*rand.Rand
}

func NewRand(seed uint64) *Rand {
	return &Rand{rand.New(rand.NewSource(seed))}
}

func (r *Rand) Address() (a common.Address) {
	r.Read(a[:])
	return a
}

func (r *Rand) AddressPtr() *common.Address {
	a := r.Address()
	return &a
}

func (r *Rand) Hash() (h common.Hash) {
	r.Read(h[:])
	return h
}

func (r *Rand) Bytes(n uint) []byte {
	b := make([]byte, n)
	r.Read(b)
	return b
}
