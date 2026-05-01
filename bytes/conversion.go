package bytes

import (
	"bytes"
	"foundation"
	"unsafe"
)

/*
StringToBytes converts a string into a newly allocated byte slice.

The resulting slice does not alias the original string memory and can be
modified safely by the caller.

Time Complexity: O(n), where n is the length of the input string.
Space Complexity: O(n) for the returned byte slice.
*/
func StringToBytes(value string) []byte {
	if len(value) == 0 {
		return nil
	}

	return []byte(value)
}

/*
StringSliceToBytes converts a slice of strings into a single byte slice.
Each string is appended to the result, followed by a provided terminator.
to allow for unambiguous reconstruction of the original slice.

Memory Complexity: O(n), where n is the total number of bytes in all strings.
Space Complexity: O(n) for the newly allocated byte slice.

Key Concepts:
1. Buffer Pre-allocation: We calculate total length first to avoid multiple
re-allocations during the build process.
2. Delimitation: Without a delimiter, "a", "bc" and "ab", "c" would be
identical in byte form.
*/
func StringSliceToBytes(slice []string, delimiter byte) []byte {
	if len(slice) == 0 {
		return nil
	}

	buffer := preallocateStringBuffer(slice)

	for _, s := range slice {
		buffer.WriteString(s)
		buffer.WriteByte(delimiter) // Null terminator
	}

	return buffer.Bytes()
}

/*
IntegerSliceToBytes converts a slice of fixed-width integer types into a contiguous
byte slice by directly copying the underlying memory representation of the slice.

No per-element encoding or transformation is performed — the raw in-memory layout
of the integer slice is preserved byte-for-byte.

This relies on the fact that all types in foundation.Integer have:

  - fixed, compile-time known widths
  - contiguous backing storage in slices
  - stable in-memory representations within a single architecture

────────────────────────────────────────────────────────────
Binary layout

For a slice:

	[]TInt{a, b, c}

the output byte slice is equivalent to:

	[mem(a) | mem(b) | mem(c)]

where mem(x) is the native in-memory representation of x.

────────────────────────────────────────────────────────────
Example (uint16 slice on little-endian machine):

	[]uint16{1, 513}

→ []byte{0x01,0x00, 0x01,0x02}

(the exact byte order follows host endianness)

────────────────────────────────────────────────────────────
Why no delimiter?

Each element has a fixed width determined by its concrete integer type:

	int8   → 1 byte each
	int16  → 2 bytes each
	int32  → 4 bytes each
	int64  → 8 bytes each
	uint*  → same widths

Because element boundaries are implicit, the byte stream is unambiguous.

────────────────────────────────────────────────────────────
Portability note

The produced byte slice is architecture-dependent:

  - endianness is host-defined
  - integer size of int/uint depends on platform

This function is therefore intended for:

  - in-process IR
  - memory-level serialization
  - high-performance pipelines
  - homogeneous runtime environments

For cross-platform or persistent binary formats, use an explicit encoding.

────────────────────────────────────────────────────────────
Performance characteristics:

	Time Complexity:  O(n)
	Space Complexity: O(n)

	with:
	  • single allocation
	  • linear memory copy
	  • no branching
	  • no reflection
	  • no temporary objects

This is appropriate for:

  - compiler IR pipelines
  - automata tables
  - memory-mapped data
  - high-throughput internal serialization
*/
func IntegerSliceToBytes[TInt foundation.Integer](slice []TInt) []byte {
	if len(slice) == 0 {
		return nil
	}

	elemSize := int(unsafe.Sizeof(slice[0]))
	totalSize := elemSize * len(slice)

	out := make([]byte, totalSize)

	src := unsafe.Slice((*byte)(unsafe.Pointer(&slice[0])), totalSize)
	copy(out, src)

	return out
}

// ------------------------------------------------------------ PRIVATE HELPERS

func preallocateStringBuffer(slice []string) *bytes.Buffer {
	totalSize := 0
	for _, s := range slice {
		totalSize += len(s) + 1 // +1 for the delimiter
	}

	return bytes.NewBuffer(make([]byte, 0, totalSize))
}
