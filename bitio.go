package bytebits

// BitWriter is an interface to a stream
// that supports writing a few bits at a time.
//
// The WriteBits method writes n bits to the stream
// from the least-significant bits of b, or writes 64 bits if n > 64.
// WriteBits returns nil if all n bits were successfully written
// or an error if not all n bits could be written.
// Whether WriteBits interprets the least- or most-significant bits of b
// as coming "first" depends on the endianness of the bit stream.
//
type BitWriter interface {
	WriteBits(n int, b uint64) (err error)
}

// BitReader is an interface to a stream
// that supports reading a few bits at a time.
// The ReadBits method reads up to n bits from the stream
// into the least-significant bits of the returned value b,
// or reads 64 bits if n > 64.
// ReadBits returns the number of bits read and any error that occurred;
// the err return must be non-nil if r != n.
// Unlike the io.Reader interface, early EOF is considered an error:
// the BitReader must return EOF if the stream ends before reading n bits.
// Whether the least- or most-significant bits of the returned value b
// are treated as coming "first" depends on the endianness of the bit stream.
//
type BitReader interface {
	ReadBits(n int) (b uint64, err error)
}

