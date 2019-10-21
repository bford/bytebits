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



// Pos represents a bit position in an abstract bit vector.
// It effectively serves as an interator over bit vector words.
//
// Add moves the position to the right if bits is positive,
// or to the left if bits is negative.
//
// Word extracts up to one word from the bit vector
// starting at the current position.
// If n is less than the word size, extracts only the next n bits
// into the least-significant n bits of the returned Word.
// 
// SetWord deposits up to one word into the bit vector
// starting at the current position,
// without affecting any other bits in the bit vector.
// If n is less than the word size, inserts only the next n bits
// from the least-significant n bits of the word parameter.
//
//type pos interface {
//	Add(n int)			// Move right (+) or left (-) n bits
//	Word(n int) Word		// Extract up to n bits at position
//	SetWord(n int, uint64 Word)	// Insert up to n bits at position
//}


type bitPut interface {
	Put(n int, w uint64)
}

type bitGet interface {
	Get(n int) uint64
}

func bitCopy(zp bitPut, xp bitGet, bits int) {
	for bits >= 64 {
		zp.Put(64, xp.Get(64))
		bits -= 64
	}
	zp.Put(bits, xp.Get(bits))
}


// A bare, endian-neutral bit position in a byte slice
type bytePos struct{
	b []byte	// Underlying byte slice
	o int		// Bit offset within current byte, 0-7
}


// catPos implements the bitGet interface by logically concatenating
// two existing getters x and y.
// Returns bits from x until xbits is exhausted, then returns bits from y.
//
type catPos struct {
	x, y bitGet
	xbits int
}

func (p *catPos) Set(x, y bitGet, xbits int) {
	p.x, p.y, p.xbits = x, y, xbits
}



func (p *catPos) Get(n int) uint64 {
	if p.xbits == 0 {		// x is exhausted, so just use y
		return p.y.Get(n)
	}
	if n > 64 {
		n = 64
	}
	if p.xbits >= n {		// just get and return bits from x
		p.xbits -= n
		return p.x.Get(n)
	}

	// Handle the transition from x to y.
	ybits := n - p.xbits
	v := p.x.Get(p.xbits) << ybits
	v |= p.y.Get(ybits)
	p.xbits = 0
	return v
}



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

