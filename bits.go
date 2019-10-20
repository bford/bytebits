// This package provides bit-manipulation and bit-field-manipulation functions
// that operate on byte slices.
// It contains functions similar to those the standard 'math/bits' package
// provides for unsigned integers, but operating on byte slices.
// I would like to see this functionality added to the Go standard library,
// perhaps as a "bytes/bits" sub-package of the current "bytes" package.
//
// This package is preliminary, unstable, and incompletely tested.  Beware.
//
// These functions could probably be sped up significantly
// via architecture-specific optimizations
// similarly to the math/bits primitives,
// but this implementation currently does not do so.
//
// XXX todo: LittleEndianBits, more tests, ...
//
package bytebits

import (
	"math/bits"
)


// BitOrder defines an interface to bit-field operations
// that depend on bit order.
// This package provides the two implementations
// BigEndian and LittleEndian.
type BitOrder interface {

	Bit(src []byte, ofs int) uint
	Uint8(src []byte, ofs int) uint8
	Uint16(src []byte, ofs int) uint16
	Uint32(src []byte, ofs int) uint32
	Uint64(src []byte, ofs int) uint64
	Bits(dst, src []byte, ofs, bits int) []byte

	SetBit(dst []byte, ofs int, b uint) []byte
	SetUint8(dst []byte, ofs int, v uint8) []byte
	SetUint16(dst []byte, ofs int, v uint16) []byte
	SetUint32(dst []byte, ofs int, v uint32) []byte
	SetUint64(dst []byte, ofs int, v uint64) []byte
	SetBits(dst, src []byte, ofs, bits int) []byte

	ShiftLeft(dst, src []byte, n int) []byte
	ShiftRight(dst, src []byte, n int) []byte
	RotateLeft(dst, src []byte, n int) []byte

	LeadingZeros(src []byte) int
	TrailingZeros(src []byte) int
}


var zeroByte = []byte{0}	// just a single zero byte


// Align indicates Left or Right bit-field alignment
// for the bit-field Insert and Extract operations.
//type Align bool

//const Left Align = false	// Left alignment
//const Right Align = true	// Right alignment


func len2(x, y []byte) int {
	l := len(x)
	if len(y) != l {
		panic("input slices must be the same length")
	}
	return l
}

// Grow grows slice z to have a length of at least l.
// If z is already of length l or longer, just returns z.
// If z is shorter but the slice has capacity at least l, returns z[:l].
// Otherwise, copies the contents of z to a new slice and returns it.
func Grow(z []byte, l int) []byte {
	if l <= len(z) {	// Slice is already long enough
		return z
	}
	if l <= cap(z) {	// Slice already has enough capacity
		return z[:l]
	}

	// Make sure slice at least doubles each allocation
	// to avoid many allocations when growing gradually
	nc := cap(z) * 2
	if nc < l {
		nc = l
	}
	nz := make([]byte, nc)
	copy(nz, z)
	return nz[:l]
}

// And sets z to the bitwise AND of slices x and y, and returns z.
// The source slices x and y must be of the same length.
// Allocates and returns a new destination slice if z is not long enough.
func And(z, x, y []byte) []byte {
	l := len2(x, y)
	z = Grow(z, l)
	for i := range x {
		z[i] = x[i] & y[i]
	}
	return z
}

// AndNot sets z to the bitwise AND of slices x and NOT y, and returns z.
// The source slices x and y must be of the same length.
// Allocates and returns a new destination slice if z is not long enough.
func AndNot(z, x, y []byte) []byte {
	l := len2(x, y)
	z = Grow(z, l)
	for i := range x {
		z[i] = x[i] &^ y[i]
	}
	return z
}

// And sets z to the bitwise OR of slices x and y, and returns z.
// The source slices x and y must be of the same length.
// Allocates and returns a new destination slice if z is not long enough.
func Xor(z, x, y []byte) []byte {
	l := len2(x, y)
	z = Grow(z, l)
	for i := range x {
		z[i] = x[i] ^ y[i]
	}
	return z
}

// Not sets z to the bitwise NOT of slice x, and returns z.
// Allocates and returns a new destination slice if z is not long enough.
func Not(z, x []byte) []byte {
	l := len(x)
	z = Grow(z, l)
	for i := range x {
		z[i] = ^x[i]
	}
	return z
}

// OnesCount returns the number of bits set in slice src.
func OnesCount(src []byte) (n int) {
	for _, v := range src {
		n += bits.OnesCount8(v)
	}
	return n
}

