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

package clonable

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func test[T ClonerOf[T]](t *testing.T, selfCloner T) {
	t.Helper()
	assert.Equal(t, selfCloner, selfCloner.Clone())
}

func TestClone(t *testing.T) {
	// By nature of being valid types for test(), they are self-cloning.
	test[Bool](t, true)
	test[Bool](t, false)
}
