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

