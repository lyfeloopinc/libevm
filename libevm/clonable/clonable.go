// Copyright 2024 the libevm authors.
//
// The libevm additions to go-ethereum are free software: you can redistribute
// them and/or modify them under the terms of the GNU Lesser General Public License
// as published by the Free Software Foundation, either version 3 of the License,
// or (at your option) any later version.
//
// The libevm additions are distributed in the hope that they will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU Lesser
// General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see
// <http://www.gnu.org/licenses/>.

// Package clonable defines types that can clone themselves.
package clonable

// A ClonerOf creates deep copies of itself. It is intended for use as a
// self-referential type parameter.
//
//	func x[T ClonerOf[T]](selfCloner T) {
type ClonerOf[T any] interface {
	Clone() T
}

// Bool is a boolean that satisfies the [ClonerOf] interface.
type Bool bool

// Clone returns a copy of b.
func (b Bool) Clone() Bool { return b }

// Unwrap returns b as its raw, underlying type.
func (b Bool) Unwrap() bool { return bool(b) }
