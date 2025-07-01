package prlp

import (
	"math/big"
	"math/bits"
)

var AddressRLPLength = 20 + 1
var HashRLPLength = 32 + 1

// BigIntLength returns the length of a big.Int in bytes
func BigIntLength(i *big.Int) int {
	return (i.BitLen() + 7) / 8
}

// Uint64Length returns the length of a uint64 in bytes
func Uint64Length(n uint64) int {
	if n == 0 {
		return 1 // 0 still requires 1 byte to represent
	}
	bitLen := bits.Len64(n) // Number of bits needed
	return (bitLen + 7) / 8 // Round up to nearest byte
}

func IntLength(n int) int {
	if n == 0 {
		return 1
	}
	abs := n
	if n < 0 {
		abs = -n
	}
	bitLen := bits.Len64(uint64(abs)) + 1 // +1 for the sign bit
	return (bitLen + 7) / 8               // Round up to full bytes
}

func CalculateRLBigIntValueLength(val *big.Int) int {
	switch valueLength := BigIntLength(val); {
	case valueLength == 1 && val.Uint64() <= 0x7f:
		return 1
	case valueLength < 56:
		return 1 + valueLength
	default:
		return 1 + IntLength(valueLength) + valueLength
	}
}

func CalculateRLPBytesLength(data []byte) int {
	switch valueLength := len(data); {
	case valueLength < 56:
		return 1 + valueLength
	default:
		return 1 + Uint64Length(uint64(valueLength)) + valueLength
	}
}

func CalculateRLP64ValueLength(val uint64) int {
	switch valueLength := Uint64Length(val); {
	case valueLength == 1 && val <= 0x7f:
		return 1
	case valueLength < 56:
		return 1 + valueLength
	default:
		return 1 + Uint64Length(uint64(valueLength)) + valueLength
	}
}

func CalculateNBytesLength(length uint64) int {
	switch {
	case length < 56:
		return int(1 + length)
	default:
		return 1 + Uint64Length(length) + int(length)
	}
}

func CalculateRLPListLength(listSize int) int {
	switch {
	case listSize < 56:
		return 1 + listSize
	default:
		return 1 + Uint64Length(uint64(listSize)) + listSize
	}
}
