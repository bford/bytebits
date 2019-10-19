// This package provides bit-manipulation and bit-field-manipulation functions
// that operate on byte slices.
// It contains functions similar to those the standard 'math/bits' package
// provides for unsigned integers, but operating on byte slices.
// I would like to see some package like this added to the Go standard library,
// perhaps as a "bytes/bits" sub-package of the current "bytes" package.
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
	Bits(dst, src []byte, ofs, bits int, dstAlign Align) []byte

	SetBit(dst []byte, ofs int, b uint) []byte
	SetUint8(dst []byte, ofs int, v uint8) []byte
	SetUint16(dst []byte, ofs int, v uint16) []byte
	SetUint32(dst []byte, ofs int, v uint32) []byte
	SetUint64(dst []byte, ofs int, v uint64) []byte
	SetBits(dst, src []byte, ofs, bits int, srcAlign Align) []byte

	ShiftLeft(dst, src []byte, n int) []byte
	ShiftRight(dst, src []byte, n int) []byte
	RotateLeft(dst, src []byte, n int) []byte

	OnesCount(src []byte) int
	LeadingZeros(src []byte) int
	TrailingZeros(src []byte) int
}


// Align indicates Left or Right bit-field alignment
// for the bit-field Insert and Extract operations.
type Align bool

const Left Align = false	// Left alignment
const Right Align = true	// Right alignment


var zeroByte = []byte{0}	// just a single zero byte


// This function oddly doesn't depend on bit-order
func onesCount(src []byte) (n int) {
	for _, v := range src {
		n += bits.OnesCount8(v)
	}
	return n
}

