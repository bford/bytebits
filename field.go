package bytebits


// Field is an interface to an object representing a bit-field
// providing bit-stream I/O and bit-manipulation operations.
//type Field interface {
//	SetBytes(SetBytes(buf []byte, ofs, width int)
//	GrowBytes(buf []byte, ofs, width int) []byte
//	ReadBits(n int) (v uint64, err error)
//	...
//}


// field represents bare, endian-neutral bit field within a byte slice.
type field struct{
	b []byte	// Underlying byte slice
	o int		// Bit offset within current byte, 0-7
	w int		// Total width of the field in bits
}

// Field is an interface to a bit-field
// providing common bit manipulation operations.
type Field interface {
	Set(x Field) Field		// Set to x
	And(x, y Field) Field		// Set to x & y
	AndNot(x, y Field) Field	// Set to x &^ y
	Or(x, y Field) Field		// Set to x | y
	Xor(x, y Field) Field		// Set to x ^ y
	Not(x Field) Field		// Set to ^x
	Count(b uint) int		// Count bits with value b
	Fill(b uint)			// Fill with bit value b
	RotateLeft(x Field, rot int) Field
	// XXX ShiftLeft, ...
}
