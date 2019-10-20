package bytebits

import (
	"math/bits"
	"encoding/binary"
)


// BigEndianOrder provides bitwise operations that treat the bits in each byte
// as having big-endian bit ordering.
//
// To illustrate, with big-endian bit ordering,
// a bit-field starting at bit offset 14 and having a width of five bits
// will contain the least-significant two bits of the second byte
// and the most-significant three bits of the third, as follows:
//
//	offset 14 = 8+6 bits            5-bit width
//	------------------------------> |<------->|
//	+-----------------+-----------------+-----------------+
//	| 7 6 5 4 3 2 1 0 | 7 6 5 4 3 2 1 0 | 7 6 5 4 3 2 1 0 | 
//	+-----------------+-----------------+-----------------+
//
// You normally invoke its methods via the standard BigEndian instance.
// For example, to extract the 5-bit field illustrated above, you can invoke:
//
//	BigEndian.Bits(dst, src, 14, 5)
//
// BigEndianOrder implements the BitOrder interface for big endian ordering.
// This interface may be used to parameterize the bit ordering in other code.
//
type BigEndianOrder struct{}

// BigEndian instantiates the BitOrder interface for bit-endian bit order.
var BigEndian = BigEndianOrder{}


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


// XXX want this?  what's the right API?
// If bits is not a multiple of 8, dstAlign determines whether
// the extracted bit-field is Left or Right aligned within dst.

// Bits extracts a bit-field of width bits starting at bit ofs in src,
// places the contents of the field into slice dst, and returns dst.
// Allocates a new byte slice if dst is null or not large enough.
// All other bits within dst are set to zero.
func (_ BigEndianOrder) Bits(dst, src []byte, ofs, bits int) []byte {

	l := (bits+7) >> 3
	lbits := bits & 7

	// Make sure dst is large enough
	dst = Grow(dst, l)

	//switch dstAlign {
	//case Left:
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

	//case Right:
	if false {
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
	// }

	return dst
}

// Bit returns the value of the bit at position ofs from the left end of src.
// Because the bytes in src are viewed as in big-endian bit order,
// offset 0 is the most-significant bit of src[0] and
// offset 7 is the least-significant bit of src[0].
func (_ BigEndianOrder) Bit(src []byte, ofs int) uint {
	return uint(src[ofs >> 3] >> (7 - (ofs & 7))) & 1
}

// Uint8 extracts a uint8 starting at bit position ofs from the left of src.
func (_ BigEndianOrder) Uint8(x []byte, ofs int) uint8 {
	var b [1]byte
	extract2(b[:], x, nil, ofs)
	return b[0]
}

// Uint16 extracts a uint16 starting at bit position ofs from the left of src.
func (_ BigEndianOrder) Uint16(x []byte, ofs int) uint16 {
	var b [2]byte
	extract2(b[:], x, nil, ofs)
	return binary.BigEndian.Uint16(b[:])
}

// Uint32 extracts a uint32 starting at bit position ofs from the left of src.
func (_ BigEndianOrder) Uint32(x []byte, ofs int) uint32 {
	var b [4]byte
	extract2(b[:], x, nil, ofs)
	return binary.BigEndian.Uint32(b[:])
}

// Uint64 extracts a uint64 starting at bit position ofs from the left of src.
func (_ BigEndianOrder) Uint64(x []byte, ofs int) uint64 {
	var b [8]byte
	extract2(b[:], x, nil, ofs)
	return binary.BigEndian.Uint64(b[:])
}


// SetBit sets the bit at position ofs from the left of src in z to value v,
// and returns the destination slice z.
// Copies and returns a larger slice if z is too small.
//
func (_ BigEndianOrder) SetBit(z []byte, ofs int, v uint) []byte {
	obytes := ofs >> 3
	obits := ofs & 7
	z = Grow(z, obytes+1)
	z[obytes] = (z[obytes] &^ (1 << obits)) | (byte(v & 1) << obits)
	return z
}

// SetBits sets the contents of a bit-field in slice dst of width bits,
// starting at bit position ofs, to the left-aligned bits in slice x.
// Copies and returns a larger slice if z is too small.
//
func(_ BigEndianOrder) SetBits(z, x []byte, ofs, bits int) []byte {

	// Ensure the destination slice contains the complete bit-field
	z = Grow(z, (ofs+bits+7) >> 3)

	// Nothing else to do if there are no bits to insert.
	if bits == 0 {
		return z
	}

	obytes := ofs >> 3
	obits := ofs & 7
	lbytes := bits >> 3
	lbits := bits & 7
	ebytes := (ofs+bits) >> 3
	ebits := (ofs+bits) & 7

	// Handle case where start and end of bit-field are in the same byte
	if ebytes == obytes {
		mask := (byte(0xff) >> obits) & (byte(0xff) << (8-ebits))
		z[obytes] = (z[obytes] &^ mask) | ((x[0] >> obits) & mask)
		return z
	}

	// Handle case where the bit field happens to be byte-aligned
	if obits == 0 {
		copy(z[obytes:], x[:lbytes])
		if ebits > 0 {
			mask := byte(0xff) >> ebits
			z[obytes+lbytes] = (z[obytes+lbytes] & mask) |
						(x[lbytes] &^ mask)
		}
		return z
	}

	// Handle first partial byte of the bit field
	i, j := obytes, 0
	z[i] = (z[i] &^ (byte(0xff) >> obits)) | (x[j] >> obits)
	i++

	// Handle all complete destination slice bytes in the bit-field


	// Handle any final partial byte 
	if lbits > 0 {
	}

	return z
}


// RotateLeft rotates all bytes in src left by n bits,
// places the result in dst, and returns dst.
// To rotate right by n bits, call RotateLeft(dst, src, -n).
// Allocates a new byte slice if dst is null or not long enough.
// The dst slice must not overlap src, except if -8 <= n <= 8,
// in which case dst and src may be identical for small in-place rotations.
func (_ BigEndianOrder) RotateLeft(dst, src []byte, n int) []byte {

	// Ensure the dst buffer is large enough
	l := len(src)
	if l == 0 {
		return dst		// nothing to rotate
	}
	dst = Grow(dst, l)
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

func (_ BigEndianOrder) LeadingZeros(src []byte) (n int) {
	for _, v := range(src) {
		if v != 0 {
			return n + bits.LeadingZeros8(v)
		}
		n += 8
	}
	return n
}

func (_ BigEndianOrder) TrailingZeros(src []byte) (n int) {
	for i := len(src)-1; i >= 0; i-- {
		v := src[i]
		if v != 0 {
			return n + bits.TrailingZeros8(v)
		}
		n += 8
	}
	return n
}

