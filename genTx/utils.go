package genTx

import (
	"github.com/1aBcD1234aBcD1/prlp/errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"math/big"
	"math/bits"
)

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

func IntUnsignedLength(n int) int {
	if n == 0 {
		return 1
	}
	abs := n
	if n < 0 {
		abs = -n
	}
	bitLen := bits.Len64(uint64(abs)) // +1 for the sign bit
	return (bitLen + 7) / 8           // Round up to full bytes
}

func CalculateRLBigIntValueLength(val *big.Int) int {
	switch valueLength := BigIntLength(val); {
	case valueLength == 1 && val.Uint64() <= 0x7f:
		return 1
	case valueLength < 56:
		return 1 + valueLength
	default:
		return 1 + IntUnsignedLength(valueLength) + valueLength
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

func RecoverPlain(sighash common.Hash, R, S, Vb *big.Int, homestead bool) (common.Address, error) {
	if Vb.BitLen() > 8 {
		return common.Address{}, errors.ErrInvalidSig
	}
	V := byte(Vb.Uint64() - 27)
	if !crypto.ValidateSignatureValues(V, R, S, homestead) {
		return common.Address{}, errors.ErrInvalidSig
	}
	// encode the signature in uncompressed format
	r, s := R.Bytes(), S.Bytes()
	sig := make([]byte, crypto.SignatureLength)
	copy(sig[32-len(r):32], r)
	copy(sig[64-len(s):64], s)
	sig[64] = V
	// recover the public key from the signature
	pub, err := crypto.Ecrecover(sighash[:], sig)
	if err != nil {
		return common.Address{}, err
	}
	if len(pub) == 0 || pub[0] != 4 {
		return common.Address{}, errors.ErrInvalidPkb
	}
	var addr common.Address
	copy(addr[:], crypto.Keccak256(pub[1:])[12:])
	return addr, nil
}

func validateSignatureValues(v byte, r, s *big.Int, homestead bool) bool {

	if r.Cmp(common.Big1) < 0 || s.Cmp(common.Big1) < 0 {
		return false
	}
	// reject upper range of s values (ECDSA malleability)
	// see discussion in secp256k1/libsecp256k1/include/secp256k1.h
	if homestead && s.Cmp(secp256k1halfN) > 0 {
		return false
	}
	// Frontier: allow s to be in full N range
	return r.Cmp(secp256k1N) < 0 && s.Cmp(secp256k1N) < 0 && (v == 0 || v == 1)
}
