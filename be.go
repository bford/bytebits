package bytebits

import (
	"math/bits"
	"encoding/binary"
)


// BigEndianOrder provides bitwise operations that treat the bits in each byte
// as having big-endian bit ordering.
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


// Normalize a byte-slice and arbitrary bit offset so the offset is 0-7.
func beNorm(b []byte, o int) ([]byte, int) {
	return b[o >> 3:], o & 7
}

// Normalize and grow a slice if needed to hold an entire bit field.
// Returns the full byte slice after growing it if needed,
// and the normalized slice and offset of the field within the full size.
func beGrow(b []byte, o, w int) ([]byte, []byte, int) {
	b = Grow(b, (o + w) >> 3)
	xb, xo := beNorm(b, o)
	return b, xb, xo
}

// Get the next 64 bits from slice b at bit offset o (0-7).
// Returns the byte slice and bit offset just past the returned bits.
func beGet64(b []byte, o int) ([]byte, int, uint64) {

	// get next 8 bytes into a uint64
	v :=	uint64(b[0]) << 56 |
		uint64(b[1]) << 48 |
		uint64(b[2]) << 40 |
		uint64(b[3]) << 32 |
		uint64(b[4]) << 24 |
		uint64(b[5]) << 16 |
		uint64(b[6]) << 8 |
		uint64(b[7])
	b = b[8:]

	if o == 0 {		// byte-aligned: touches only 8 bytes
		return b, o, v
	}

	// not byte-aligned: touches 9 bytes total
	return b, o, (v << o) | uint64(b[0]) >> (8-o)
}

// Get n or a maximum of 64 bits from slice b at bit offset o (0-7).
// Returns the byte slice and bit offset just past the returned bits,
// and the read data in the least-significant bits of a uint64.
func beGet(b []byte, o, n int) ([]byte, int, uint64) {

	if n >= 64 {			// get a full uint64
		return beGet64(b, o)
	} 

	// Get a partial uint64 a few bits at a time
	v := uint64(0)
	for n > 0 {
		if o == 0 {		// currently byte-aligned
			if n >= 8 {	// get the next full byte
				v = (v << 8) | uint64(b[0])
				b = b[1:]
				n -= 8
			} else {	// get a last partial byte
				v = (v << n) | uint64(b[0]) >> (8-n)
				o = n
				n = 0
			}
		} else {		// not currently byte-aligned
			r := 8-o	// remaining bits in the current byte
			if n >= r {	// get the rest of the current byte
				v = (v << r) | (uint64(b[0]) &^ (0xff << r))
				b = b[1:]
				o = 0
				n -= r
			} else {	// get only part of the current byte
				v = (v << n) | ((uint64(b[0]) >> (r-n)) &^
							(0xff << n))
				o += n
				n = 0
			}
		}
	}
	return b, o, v
}

// Put 64 bits into slice b at offset o, which must be in the range 0-7.
func bePut64(b []byte, o int, v uint64) ([]byte, int) {

	if o == 0 {		// currently byte-aligned
		b[0] = byte(v >> 56)
		b[1] = byte(v >> 48)
		b[2] = byte(v >> 40)
		b[3] = byte(v >> 32)
		b[4] = byte(v >> 24)
		b[5] = byte(v >> 16)
		b[6] = byte(v >> 8)
		b[7] = byte(v)
		return b[8:], o
	}

	// not byte-aligned: touches 9 bytes total
	r := 8-o		// remaining bits in the current byte
	b[0] = (b[0] &^ (0xff >> o)) | byte(v >> (56+o))
	v <<= r			// shift rest of v into byte-alignment
	b[1] = byte(v >> 56)
	b[2] = byte(v >> 48)
	b[3] = byte(v >> 40)
	b[4] = byte(v >> 32)
	b[5] = byte(v >> 24)
	b[6] = byte(v >> 16)
	b[7] = byte(v >> 8)
	b[8] = byte(v) | (b[8] & (byte(0xff) >> o))
	return b[8:], o
}

// Put n or a maximum of 64 bits into slice b at bit offset o (0-7).
// The bits to put are passed in the least-significant bits of v.
// Returns the byte slice and bit offset just past the put bits.
func bePut(b []byte, o, n int, v uint64) ([]byte, int) {

	if n >= 64 {			// put a full uint64
		return bePut64(b, o, v)
	}

	// Put a partial uint64 a few bits at a time
	for n > 0 {
		if o == 0 {		// currently byte-aligned
			if n >= 8 {	// put the next full byte
				n -= 8
				b[0] = byte(v >> n)
				b = b[1:]
			} else {	// put a last partial byte
				b[0] = byte(v << (8-n)) | (b[0] & (0xff >> n))
				o = n
				n = 0
			}
		} else {		// not currently byte-aligned
			r := 8-o	// remaining bits in the current byte
			if n >= r {	// put the rest of the current byte
				n -= r
				m := byte(0xff) >> o
				b[0] = (b[0] &^ m) | (byte(v >> n) & m)
				b = b[1:]
				o = 0
			} else {	// put only part of the current byte
				m := byte((0xff >> o) & (0xff << (r-n)))
				b[0] = (b[0] &^ m) | (byte(v << (r-n)) & m)
				o += n
				n = 0
			}
		}
	}
	return b, o
}


func beCopy(zb, xb []byte, zo, xo, w int) ([]byte, []byte, int, int) {
	var v uint64
	for w >= 64 {
		xb, xo, v = beGet64(xb, xo)
		zb, zo = bePut64(zb, zo, v)
		w -= 64
	}
	xb, xo, v = beGet(xb, xo, w)
	zb, zo = bePut(zb, zo, w, v)
	return zb, xb, zo, xo
}


// Copy copies a bit-field of width bits starting at offset xofs in x
// into a field of the same width starting at offset zofs in z,
// then returns z.
// Copies z and returns a new slice if z is null or not large enough.
// All other bits within z are left unmodified.
//
func (be BigEndianOrder) Copy(z []byte, x []byte, zofs, xofs, w int) []byte {
	xb, xo := beNorm(x, xofs)
	z, zb, zo := beGrow(z, zofs, w)
	beCopy(zb, xb, zo, xo, w)
	return z
}

// Get up to w bits from slice xb at bit offset xo.
func (_ BigEndianOrder) get(xb []byte, xo, w int) (v uint64) {
	xb, xo = beNorm(xb, xo)
	xb, xo, v = beGet(xb, xo, w)
	return v
}

// Bit returns the value of the bit at position xofs bits from the left of x.
func (be BigEndianOrder) Bit(x []byte, xofs int) uint {
	return uint(be.get(x, xofs, 1))
}

// Uint8 extracts a uint8 starting at bit position xofs from the left of x.
func (be BigEndianOrder) Uint8(x []byte, xofs int) uint8 {
	return uint8(be.get(x, xofs, 8))
}

// Uint16 extracts a uint16 starting at bit position xofs from the left of x.
func (be BigEndianOrder) Uint16(x []byte, xofs int) uint16 {
	return uint16(be.get(x, xofs, 16))
}

// Uint32 extracts a uint32 starting at bit position xofs from the left of x.
func (be BigEndianOrder) Uint32(x []byte, xofs int) uint32 {
	return uint32(be.get(x, xofs, 32))
}

// Uint64 extracts a uint64 starting at bit position xofs from the left of x.
func (be BigEndianOrder) Uint64(x []byte, xofs int) uint64 {
	return uint64(be.get(x, xofs, 64))
}


// SetBits setsw all bits in a bit-field of width bits
// starting at offset zofs in z to the same bit value b.
// Copies z and returns a new slice if z is null or not large enough.
//
//func (_ BigEndianOrder) SetBits(z []byte, zofs, bits int, bit uint) []byte {
//	var zp bePos
//	z = zp.Grow(z, zofs, bits)
//	v := ^(uint64(bit & 1) - 1)	// v = all zero bits or all one bits
//	for bits >= 64 {
//		zp.Put(64, v)
//	}
//	zp.Put(bits, v)
//	return z
//}

// Put up to w bits into slice zb at bit offset zofs.
func (_ BigEndianOrder) put(z []byte, zofs, w int, v uint64) []byte {
	z, zb, zo := beGrow(z, zofs, w)
	zb, zo = bePut(zb, zo, w, v)
	return z
}

// PutBit sets the bit at zofs in slice z to bit value v.
// Copies z and returns a new slice if z is null or not large enough.
//
func (be BigEndianOrder) PutBit(z []byte, zofs int, v uint) []byte {
	return be.put(z, zofs, 1, uint64(v))
}

// PutUint8 sets the uint8 starting at zofs in slice z to value v.
// Copies z and returns a new slice if z is null or not large enough.
//
func (be BigEndianOrder) PutUint8(z []byte, zofs int, v uint8) []byte {
	return be.put(z, zofs, 8, uint64(v))
}

// PutUint16 sets the uint16 starting at zofs in slice z to value v.
// Copies z and returns a new slice if z is null or not large enough.
//
func (be BigEndianOrder) PutUint16(z []byte, zofs int, v uint16) []byte {
	return be.put(z, zofs, 16, uint64(v))
}

// PutUint32 sets the uint32 starting at zofs in slice z to value v.
// Copies z and returns a new slice if z is null or not large enough.
//
func (be BigEndianOrder) PutUint32(z []byte, zofs int, v uint32) []byte {
	return be.put(z, zofs, 32, uint64(v))
}

// PutUint64 sets the uint64 starting at zofs in slice z to value v.
// Copies z and returns a new slice if z is null or not large enough.
//
func (be BigEndianOrder) PutUint64(z []byte, zofs int, v uint64) []byte {
	return be.put(z, zofs, 64, uint64(v))
}

// PutBytes writes the contents of byte b slice into slice z at bit offset zofs.
// Copies z and returns a new slice if z is nil or not large enough.
//
func (be BigEndianOrder) PutBytes(z []byte, zofs int, b []byte) []byte {
	z, zb, zo := beGrow(z, zofs, len(b) * 8)
	for len(b) >= 8 {	// put 8 bytes at a time
		v := binary.BigEndian.Uint64(b)
		zb, zo = bePut64(zb, zo, v)
		b = b[8:]
	}
	for len(b) > 0 {	// put last few bytes one at a time
		zb, zo = bePut(zb, zo, 8, uint64(b[0]))
		b = b[1:]
	}
	return z
}


// RotateLeft sets slice z to the contents of x rotated left by rot bits.
// To rotate right, pass a negative value for rot.
// Copies z and returns a new slice if z is nil or not large enough.
// The slices x and z must not overlap, except if -8 <= rot <= 8,
// in which case x and z may be identical for small in-place bit rotations.
func (be BigEndianOrder) RotateLeft(z, x []byte, rot int) []byte {

	// Ensure destination z is large enough.
	z = Grow(z, len(x))

	if rot == 0 || len(x) == 0 {	// Special case: no rotation
		copy(z, x)
		return z
	} else if rot > 0 && rot <= 8 {	// Special case: small left rotation
		l := len(x)
		r := 8-rot
		c := x[0] >> r
		for i := l-1; i >= 0; i-- {
			v := x[i]
			z[i] = (v << rot) | c
			c = v >> r
		}
		return z
	} else if rot < 0 && rot >= -8 { // Special case: small right rotation
		rot = -rot
		l := len(x)
		r := 8-rot
		c := x[l-1] << r
		for i := range x {
			v := x[i]
			z[i] = (v >> rot) | c
			c = v << r
		}
		return z
	}

	// Determine the starting bit position to copy from source field x
	w := len(x) * 8
	if w == 0 {
		return z	// nothing to rotate in empty field
	}
	rot = rot % w
	if rot < 0 {
		rot += w
	}

	// Copy bits until the end of the source field
	xb, xo := beNorm(x, rot)
	zb, xb, zo, xo := beCopy(z, xb, 0, xo, w - rot)

	// Then copy the rest of the bits from the beginning of the source
	zb, xb, zo, xo = beCopy(zb, x, zo, 0, rot)
	return z
}


// Leading counts the number of consecutive leading bits with value b
// in slice z starting from the most-significant bit of the first byte.
func (be BigEndianOrder) Leading(z []byte, b uint) (n int) {
	switch b {
	case 0:
		for _, v := range(z) {
			if v != 0 {
				return n + bits.LeadingZeros8(v)
			}
			n += 8
		}
	case 1:
		for _, v := range(z) {
			if v != 0 {
				return n + bits.LeadingZeros8(^v)
			}
			n += 8
		}
	default:
		panic("Count: invalid bit value")
	}
	return n
}

// Trailing counts the number of consecutive trailing bits with value b
// in slice z starting from the least-significant bit of the last byte.
func (be BigEndianOrder) Trailing(z []byte, b uint) (n int) {
	switch b {
	case 0:
		for i := len(z)-1; i >= 0; i-- {
			v := z[i]
			if v != 0 {
				return n + bits.TrailingZeros8(v)
			}
			n += 8
		}
	case 1:
		for i := len(z)-1; i >= 0; i-- {
			v := z[i]
			if v != 0xff {
				return n + bits.TrailingZeros8(^v)
			}
			n += 8
		}
	default:
		panic("Count: invalid bit value")
	}
	return n
}

func (be BigEndianOrder) Field(buf []byte, ofs, width int) Field {
	return (&BigEndianField{}).Init(buf, ofs, width)
}

