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
package bytebits


// Extract an l-byte bit-field at bit position ofs
// from two logically-concatenated byte-slices.
// This is a useful primitive to implement multiple other operations.
func extract2(z, x, y []byte, ofs, l int) []byte {
	if ofs < 0 || l < 0 {
		panic("extract2: invalid arguments") 
	}

	// Ensure that destination slice z is large enough
	if len(z) < l {
		z = make([]byte, l)
	}

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
			z[i] = (x[i-xlen] << obits) | (y[i+1-xlen] >> rbits)
		}
	}

	return z
}

var zeroByte = []byte{0}	// just a single zero byte

// Extract extracts from src a bit-field of width bits starting at bit ofs,
// left-aligns the contents of the bit-field into slice dst, and returns dst.
// Allocates a new byte slice if dst is null or not large enough.
func ExtractLeft(dst, src []byte, ofs, bits int) []byte {

	// Extract a byte-padded bit-field from the source,
	// with a single zero byte logically appended to ensure enough bytes
	l := (bits+7) >> 3
	dst = extract2(dst, src, zeroByte, ofs, l)

	// Mask any extra bits and/or bytes at the end
	lbits := bits & 7
	if lbits != 0 {
		dst[l-1] &= byte(int(0xff00) >> lbits)
	}
	for ; l < len(dst); l++ {
		dst[l] = 0
	}
	return dst
}

// RotateLeft rotates all bytes in src left by n bits,
// places the result in dst, and returns dst.
// To rotate right by n bits, call RotateLeft(dst, src, -n).
// Allocates a new byte slice if dst is null or not long enough.
// The dst slice must not overlap src, except if -8 <= n <= 8,
// in which case dst and src may be identical for small in-place rotations.
func RotateLeft(dst, src []byte, n int) []byte {

	// Ensure the dst buffer is large enough
	l := len(src)
	lbits := l*8
	if len(dst) < l {
		dst = make([]byte, l)
	}

	// Handle small rotations we can handle in-place
	if n > 8 {			// large left rotate
		return extract2(dst, src, src, n % lbits, lbits)

	} else if n < -8 {		// large right rotate
		return extract2(dst, src, src, l + n % lbits, lbits)

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

// XXX ExtractRight?
// XXX ShiftLeft, ShiftRight

