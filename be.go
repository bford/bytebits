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
// XXX todo: LittleEndianBits, ...
//
package bytebits

import (
	"math/bits"
	"encoding/binary"
)


// BigEndianBits implements the BitOrder interfacea for big-endian bit-ordering.
type BigEndianBits struct{}

// BigEndian instantiates the BitOrder interface for bit-endian bit order.
var BigEndian = BigEndianBits{}


// Returns dst if dst is at least l bytes, otherwise allocates a new dst.
func mk(dst []byte, l int) []byte {
	if len(dst) < l {
		return make([]byte, l)
	}
	return dst
}


// Extract a bit-field the size of slice z
// from the two logically-concatenated byte-slices x and y,
// starting at bit position ofs within the concatenated source slices.
// This is a useful primitive to implement multiple other operations.
func extract2(z, x, y []byte, ofs int) {
	l := len(z)
	obits := ofs & 7
	obytes := ofs >> 3

	// Skip any prefix bytes we don't need from x and possibly y
	if obytes < len(x) {
		x = x[obytes:]			// some of x remains
	} else {
		x = y[obytes - len(x):]		// all of x skipped
		y = nil
	}

	xlen := len(x)
	ylen := len(y)

	if obits == 0 {		// Special case of extraction on byte boundary
		copy(z, x)
		if l > xlen {
			if l > xlen + ylen {
				panic("input does not contain bit field")
			}
			copy(z[xlen:l], y)
		}
	} else {
		i := 0
		rbits := 8-obits
		if l < xlen {
			xlen = l+1	// we'll only need x
		}
		for ; i+1 < xlen; i++ {	// both bytes within x
			z[i] = (x[i] << obits) | (x[i+1] >> rbits)
		}
		if i < l {		// boundary crossing from x to y
			z[i] = (x[i] << obits) | (y[0] >> rbits)
			i++
		}
		for ; i < l; i++ {	// both bytes within z
			z[i] = (y[i-xlen] << obits) | (y[i+1-xlen] >> rbits)
		}
	}
}

// Bits extracts a bit-field of width bits starting at bit ofs in src,
// places the contents of the field into slice dst, and returns dst.
// Allocates a new byte slice if dst is null or not large enough.
// If bits is not a multiple of 8, dstAlign determines whether
// the extracted bit-field is Left or Right aligned within dst.
// All other bits within dst are set to zero.
func (_ BigEndianBits) Bits(dst, src []byte, ofs, bits int, dstAlign Align) []byte {

	l := (bits+7) >> 3
	lbits := bits & 7

	// Make sure dst is large enough
	dst = mk(dst, l)

	switch dstAlign {
	case Left:
		// Extract a byte-padded left-aligned bit-field from src.
		// Logically append a zero byte to ensure enough source bits.
		extract2(dst[:l], src, zeroByte, ofs)

		// Mask any extra bits and/or bytes at the right end
		if lbits != 0 {
			dst[l-1] &= byte(int(0xff00) >> lbits)
		}
		for ; l < len(dst); l++ {
			dst[l] = 0
		}

	case Right:
		// Extract a byte-padded right-aligned bit-field from src.
		pad := len(dst) - l
		for i := range(dst[:pad]) {
			dst[i] = 0
		}
		fld := dst[pad:]	// only the l-byte field itself

		// Logically prepend a zero byte to ensure enough source bits.
		if lbits == 0 {
			extract2(fld, src, nil, ofs)
		} else {
			extract2(fld, zeroByte, src, lbits + ofs)

			// Mask extra bits on the left end
			fld[0] &= byte(0xff) >> (8 - lbits)
		}
	}
	return dst
}

// Returns the value of the bit at position ofs from the left end of src.
// Because the bytes in src are viewed as in big-endian bit order,
// offset 0 is the most-significant bit of src[0] and
// offset 7 is the least-significant bit of src[0].
func (_ BigEndianBits) Bit(src []byte, ofs int) uint {
	return uint(src[ofs >> 3] >> (7 - (ofs & 7))) & 1
}

// Extracts a uint8 starting at bit position ofs from the left of src.
func (_ BigEndianBits) Uint8(src []byte, ofs int) uint8 {
	var b [1]byte
	extract2(b[:], src, nil, ofs)
	return b[0]
}

// Extracts a uint16 starting at bit position ofs from the left of src.
func (_ BigEndianBits) Uint16(src []byte, ofs int) uint16 {
	var b [2]byte
	extract2(b[:], src, nil, ofs)
	return binary.BigEndian.Uint16(b[:])
}

// Extracts a uint32 starting at bit position ofs from the left of src.
func (_ BigEndianBits) Uint32(src []byte, ofs int) uint32 {
	var b [4]byte
	extract2(b[:], src, nil, ofs)
	return binary.BigEndian.Uint32(b[:])
}

// Extracts a uint64 starting at bit position ofs from the left of src.
func (_ BigEndianBits) Uint64(src []byte, ofs int) uint64 {
	var b [8]byte
	extract2(b[:], src, nil, ofs)
	return binary.BigEndian.Uint64(b[:])
}

// RotateLeft rotates all bytes in src left by n bits,
// places the result in dst, and returns dst.
// To rotate right by n bits, call RotateLeft(dst, src, -n).
// Allocates a new byte slice if dst is null or not long enough.
// The dst slice must not overlap src, except if -8 <= n <= 8,
// in which case dst and src may be identical for small in-place rotations.
func (_ BigEndianBits) RotateLeft(dst, src []byte, n int) []byte {

	// Ensure the dst buffer is large enough
	l := len(src)
	if l == 0 {
		return dst		// nothing to rotate
	}
	dst = mk(dst, l)
	if n < -8 {			// large right rotate
		lbits := l * 8		// recalculate it as a left rotate
		n = lbits + (n % lbits)
	}
	if n > 8 {			// large left rotate
		extract2(dst[:l], src[(n >> 3) % l:], src, n & 7)

	} else if n == 0 {		// special case of no rotation
		copy(dst, src)

	} else if n > 0 && n <= 8 {	// small left rotate
		rbits := 8-n
		c := src[0] >> rbits
		for i := l-1; i >= 0; i-- {
			v := src[i]
			dst[i] = (v << n) | c
			c = v >> rbits
		}
	} else if n < 0 && n >= -8 {	// small right rotate
		n = -n
		lbits := 8-n
		c := src[l-1] << lbits
		for i := 0; i < l; i++ {
			v := src[i]
			dst[i] = (v >> n) | c
			c = v << lbits
		}
	}
	return dst
}

// XXX SetBits etc
// XXX ExtractRight?
// XXX ShiftLeft, ShiftRight

func (_ BigEndianBits) OnesCount(src []byte) (n int) {
	return onesCount(src)
}

func (_ BigEndianBits) LeadingZeros(src []byte) (n int) {
	for _, v := range(src) {
		if v != 0 {
			return n + bits.LeadingZeros8(v)
		}
		n += 8
	}
	return n
}

func (_ BigEndianBits) TrailingZeros(src []byte) (n int) {
	for i := len(src)-1; i >= 0; i-- {
		v := src[i]
		if v != 0 {
			return n + bits.TrailingZeros8(v)
		}
		n += 8
	}
	return n
}

