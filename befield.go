package bytebits

import (
	"io"
	"math/bits"
)


// BigEndianField represents a bit field within a byte slice
// interpreted as a big-endian sequence of bits.
type BigEndianField field

// EOF represents an end-of-File error, and is identical to io.EOF.
var EOF = io.EOF


// Init sets the field to refer to a big-endian bit field within slice buf,
// starting at bit offset ofs and extending for width bits.
// The caller must ensure that the underlying slice is large enough
// to contain the complete bit field specified;
// otherwise accesses to the bit field may yield bounds check panics.
func (z *BigEndianField) Init(buf []byte, ofs, width int) Field {
	z.b = buf[ofs >> 3:]
	z.o = ofs & 7
	z.w = width
	return z
}

// Grow points the field to a big-endian bit field within slice buf,
// copying buf to a new larger buffer if needed to include the bit field.
// Returns buf or the newly-allocated buffer if it was grown.
func (z *BigEndianField) Grow(buf []byte, ofs, width int) []byte {
	buf = Grow(buf, (ofs+width) >> 3)
	z.Init(buf, ofs, width)
	return buf
}

// ReadBits implements the BitReader interface,
// reading up to n bits from the start of the field, or 64 bits maximum.
// On success, returns the bits read
// in the least-significant bits of the returned value b,
// and shrinks the field to skip the n bits read.
// Returns an EOF error if the bit field is less than n bits wide.
func (z *BigEndianField) ReadBits(n int) (v uint64, err error) {
	if n > 64 {
		n = 64
	}
	if n > z.w {
		return 0, EOF
	}
	z.b, z.o, v = beGet(z.b, z.o, n)
	z.w -= n
	return v, nil
}

// Copy sets the contents of bit field z to that of field x,
// and returns z.
// The source field x must be at least as long as field z.
func (z *BigEndianField) Set(x Field) Field {
	xf := x.(*BigEndianField)
	xb, xo, zb, zo, w := xf.b, xf.o, z.b, z.o, z.w
	var xv uint64
	for w >= 64 {
		xb, xo, xv = beGet64(xb, xo)
		zb, zo = bePut64(zb, zo, xv)
	}
	xb, xo, xv = beGet(xb, xo, w)
	zb, zo = bePut(zb, zo, w, xv)
	return z
}

// And sets the contents of bit field z to the bitwise AND of fields x and y,
// and returns z.
// The source fields x and y must be at least as long as field z.
func (z *BigEndianField) And(x, y Field) Field {
	xf, yf := x.(*BigEndianField), y.(*BigEndianField)
	xb, xo, yb, yo, zb, zo, w := xf.b, xf.o, yf.b, yf.o, z.b, z.o, z.w
	var xv, yv uint64
	for w >= 64 {
		xb, xo, xv = beGet64(xb, xo)
		yb, yo, yv = beGet64(yb, yo)
		zb, zo = bePut64(zb, zo, xv & yv)
	}
	xb, xo, xv = beGet(xb, xo, w)
	yb, yo, yv = beGet(yb, yo, w)
	zb, zo = bePut(zb, zo, w, xv & yv)
	return z
}

// AndNot sets the contents of bit field z
// to the bitwise AND of fields x and NOT y,
// and returns z.
// The source fields x and y must be at least as long as field z.
func (z *BigEndianField) AndNot(x, y Field) Field {
	xf, yf := x.(*BigEndianField), y.(*BigEndianField)
	xb, xo, yb, yo, zb, zo, w := xf.b, xf.o, yf.b, yf.o, z.b, z.o, z.w
	var xv, yv uint64
	for w >= 64 {
		xb, xo, xv = beGet64(xb, xo)
		yb, yo, yv = beGet64(yb, yo)
		zb, zo = bePut64(zb, zo, xv &^ yv)
	}
	xb, xo, xv = beGet(xb, xo, w)
	yb, yo, yv = beGet(yb, yo, w)
	zb, zo = bePut(zb, zo, w, xv &^ yv)
	return z
}

// Or sets the contents of bit field z to the bitwise OR of fields x and y,
// and returns z.
// The source fields x and y must be at least as long as field z.
func (z *BigEndianField) Or(x, y Field) Field {
	xf, yf := x.(*BigEndianField), y.(*BigEndianField)
	xb, xo, yb, yo, zb, zo, w := xf.b, xf.o, yf.b, yf.o, z.b, z.o, z.w
	var xv, yv uint64
	for w >= 64 {
		xb, xo, xv = beGet64(xb, xo)
		yb, yo, yv = beGet64(yb, yo)
		zb, zo = bePut64(zb, zo, xv | yv)
	}
	xb, xo, xv = beGet(xb, xo, w)
	yb, yo, yv = beGet(yb, yo, w)
	zb, zo = bePut(zb, zo, w, xv | yv)
	return z
}

// Xor sets the contents of bit field z to the bitwise XOR of fields x and y,
// and returns z.
// The source fields x and y must be at least as long as field z.
func (z *BigEndianField) Xor(x, y Field) Field {
	xf, yf := x.(*BigEndianField), y.(*BigEndianField)
	xb, xo, yb, yo, zb, zo, w := xf.b, xf.o, yf.b, yf.o, z.b, z.o, z.w
	var xv, yv uint64
	for w >= 64 {
		xb, xo, xv = beGet64(xb, xo)
		yb, yo, yv = beGet64(yb, yo)
		zb, zo = bePut64(zb, zo, xv ^ yv)
	}
	xb, xo, xv = beGet(xb, xo, w)
	yb, yo, yv = beGet(yb, yo, w)
	zb, zo = bePut(zb, zo, w, xv ^ yv)
	return z
}

// Not sets the contents of bit field z to the bitwise NOT of field x,
// and returns z.
// The source field x must be at least as long as field z.
func (z *BigEndianField) Not(x Field) Field {
	xf := x.(*BigEndianField)
	xb, xo, zb, zo, w := xf.b, xf.o, z.b, z.o, z.w
	var xv uint64
	for w >= 64 {
		xb, xo, xv = beGet64(xb, xo)
		zb, zo = bePut64(zb, zo, ^xv)
	}
	xb, xo, xv = beGet(xb, xo, w)
	zb, zo = bePut(zb, zo, w, ^xv)
	return z
}

// RotateLeft sets field z to field x rotated left by rot bits.
// To rotate right, pass a negative value for rot.
// Field x must be at least as long as z.
// The slices underlying x and z must not overlap, except if -7 <= rot <= 7,
// in which case x and z may be identical for small in-place bit rotations.
func (z *BigEndianField) RotateLeft(x Field, rot int) Field {
	// Determine the starting bit position to copy from source field x
	zb, zo, w := z.b, z.o, z.w
	rot = rot % w
	if rot < 0 {
		rot += w
	}

	// Copy bits until the end of the source field
	xf := x.(*BigEndianField)
	xb, xo := beNorm(xf.b, xf.o + rot)
	zb, xb, zo, xo = beCopy(zb, xb, zo, xo, w - rot)

	// Then copy the rest of the bits from the beginning of the source
	zb, xb, zo, xo = beCopy(zb, xf.b, zo, xf.o, rot)
	return z
}

// Count returns the number of bits with value b (0 or 1) in field z.
func (z *BigEndianField) Count(b uint) (n int) {
	zb, zo, w := z.b, z.o, z.w
	var v uint64
	switch b {
	case 0:
		for w >= 64 {
			zb, zo, v = beGet64(zb, zo)
			n += bits.OnesCount64(^v)
		}
		zb, zo, v = beGet(zb, zo, w)
		n += bits.OnesCount64(v ^ ((1 << w) - 1))
	case 1:
		for w >= 64 {
			zb, zo, v = beGet64(zb, zo)
			n += bits.OnesCount64(v)
		}
		zb, zo, v = beGet(zb, zo, w)
		n += bits.OnesCount64(v)
	default:
		panic("Count: invalid bit value")
	}
	return n
}

// Fill sets all bits in field z to bit value b (0 or 1).
func (z *BigEndianField) Fill(b uint) {
	zb, zo, w := z.b, z.o, z.w
	switch b {
	case 0:
		for w >= 64 {
			zb, zo = bePut64(zb, zo, 0)
		}
		zb, zo = bePut(zb, zo, w, 0)
	case 1:
		for w >= 64 {
			zb, zo = bePut64(zb, zo, (1<<64)-1)
		}
		zb, zo = bePut(zb, zo, w, (1<<64)-1)
	default:
		panic("Count: invalid bit value")
	}
}

