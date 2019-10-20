
This package implements bitwise and bit-field operations on byte slices
in the [Go](https://golang.org) language.
These are functions that I wish
were in the Go standard library's
[bytes](https://golang.org/pkg/bytes/) package,
or perhaps a subpackage that might be called `bytes/bits`.

For full API information please see the
[Godoc documentation](https://godoc.org/github.com/bford/bytebits).

In brief, the package includes:

*	Insertion and extraction of bit-fields of arbitrary bit widths
	at arbitrary bit offsets in a byte slice.
*	Insertion and extraction of single bits and fixed-length integers
	at arbitrary bit offsets in a byte slice.
*	Support for big-endian and (soon) little-endian bit-ordering.
*	Bitwise And, AndNot, Or, Xor, and Not on byte slices.
*	LeadingZeros, TrailingZeros, and OnesCount on byte slices.

I hope you find it useful!  Pull requests welcome.

