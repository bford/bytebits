package bytebits

import (
	"io"
)


type BigEndianField field

// EOF represents an end-of-File error, and is identical to io.EOF.
var EOF = io.EOF


// SetBytes sets the field to refer to a big-endian bit field within slice buf,
// starting at bit offset ofs and extending for width bits.
// The caller must ensure that the underlying slice is large enough
// to contain the complete bit field specified;
// otherwise accesses to the bit field may yield bounds check panics.
func (f *BigEndianField) SetBytes(buf []byte, ofs, width int) *BigEndianField {
	f.b = buf[ofs >> 3:]
	f.o = ofs & 7
	f.w = width
	return f
}

// Grow points the field to a big-endian bit field within slice buf,
// copying buf to a new larger buffer if needed to include the bit field.
// Returns buf or the newly-allocated buffer if it was grown.
func (f *BigEndianField) GrowBytes(buf []byte, ofs, width int) []byte {
	buf = Grow(buf, (ofs+width) >> 3)
	f.SetBytes(buf, ofs, width)
	return buf
}

// ReadBits implements the BitReader interface,
// reading up to n bits from the start of the field, or 64 bits maximum.
// On success, returns the bits read
// in the least-significant bits of the returned value b,
// and shrinks the field to skip the n bits read.
// Returns an EOF error if the bit field is less than n bits wide.
func (f *BigEndianField) ReadBits(n int) (v uint64, err error) {
	if n > 64 {
		n = 64
	}
	if n > f.w {
		return 0, EOF
	}
	f.b, f.o, v = beGet(f.b, f.o, n)
	f.w -= n
	return v, nil
}

// And sets the contents of bit field z to the bitwise AND of fields x and y,
// and returns z.
// The source fields x and y must be at least as long as field z.
func (z *BigEndianField) And(x, y *BigEndianField) *BigEndianField {
	xb, xo, yb, yo, zb, zo, w := x.b, x.o, y.b, y.o, z.b, z.o, z.w
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

// RotateLeft sets field z to field x rotated left by rot bits.
// To rotate right, pass a negative value for rot.
// Field x must be at least as long as z.
// The slices underlying x and z must not overlap, except if -7 <= rot <= 7,
// in which case x and z may be identical for small in-place bit rotations.
func (z *BigEndianField) RotateLeft(x *BigEndianField, rot int) *BigEndianField {
	// Determine the starting bit position to copy from source field x
	zb, zo, w := z.b, z.o, z.w
	rot = rot % w
	if rot < 0 {
		rot += w
	}

	// Copy bits until the end of the source field
	xb, xo := beNorm(x.b, x.o + rot)
	zb, xb, zo, xo = beCopy(zb, xb, zo, xo, w - rot)

	// Then copy the rest of the bits from the beginning of the source
	zb, xb, zo, xo = beCopy(zb, x.b, zo, x.o, rot)
	return z
}

