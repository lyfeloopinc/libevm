// Package as converts between native Ethereum types and their libevm
// equivalents.
//
// All functions are named for their returned type, not their input, to improve
// readability at the call site:
//
//	list := as.GethAccessList(...) // list is a geth-native AccessList
//	list := as.LibEVMAccessList(...) // list is a libevm AccessList mirror
package as

import (
	"unsafe"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/libevm"
)

// GethAccessList converts the libevm AccessList to its geth equivalent.
func GethAccessList(l libevm.AccessList) types.AccessList {
	return *convert[libevm.AccessList, types.AccessList](&l)
}

// LibEVMAccessList converts the geth AccessList to its libevm equivalent.
func LibEVMAccessList(l types.AccessList) libevm.AccessList {
	return *convert[types.AccessList, libevm.AccessList](&l)
}

// GethAccessTuple converts the libevm AccessTuple to its geth equivalent.
func GethAccessTuple(t libevm.AccessTuple) types.AccessTuple {
	return *convert[libevm.AccessTuple, types.AccessTuple](&t)
}

// LibEVMAccessTuple converts the geth AccessTuple to its libevm equivalent.
func LibEVMAccessTuple(t types.AccessTuple) libevm.AccessTuple {
	return *convert[types.AccessTuple, libevm.AccessTuple](&t)
}

// convert uses [unsafe.Pointer] type conversion as allowed by Pattern (1) of
// its documentation. Any converter using convert MUST have extensive tests in
// place to prove identity of the two types.
func convert[T any, U any](v *T) *U {
	return (*U)(unsafe.Pointer(v))
}
