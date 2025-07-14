package tx

import (
	"bytes"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

// Here it is defined global variables and the initialization of other variables
var CHAIN_ID = big.NewInt(1).Int64()
var EncodedAddressRLPLength = byte(0x94)
var EncodedHashRLPLength = byte(0xa0)
var AddressRLPLength = 20 + 1
var HashRLPLength = 32 + 1
var V_MULTIPLIER = big.NewInt(CHAIN_ID * 2)
var V_MULTIPLER_INT64 = V_MULTIPLIER.Int64()
var V_MULTIPLER_UINT64 = V_MULTIPLIER.Uint64()

var (
	secp256k1N, _       = new(big.Int).SetString("fffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd0364141", 16)
	secp256k1NBytes     = secp256k1N.Bytes()
	secp256k1halfN      = new(big.Int).Div(secp256k1N, big.NewInt(2))
	secp256k1halfNBytes = secp256k1halfN.Bytes()
	BYTE_1              = common.Big1.Bytes()
)

var (
	SIGNER_VALUES        []byte
	SIGNER_VALUES_LENGTH int
)

func Init(chainId *big.Int) {
	CHAIN_ID = chainId.Int64()
	V_MULTIPLIER = big.NewInt(CHAIN_ID * 2)
	V_MULTIPLER_INT64 = V_MULTIPLIER.Int64()
	V_MULTIPLER_UINT64 = V_MULTIPLIER.Uint64()
	calculateSignerValues()
}

func calculateSignerValues() {
	l := new(big.Int).SetInt64(CHAIN_ID)
	result := make([]byte, 0, 10)
	buf := bytes.NewBuffer(result)
	_, err := WriteUint64(buf, l.Uint64())
	if err != nil {
		panic(err)
	}
	buf.WriteByte(0x80)
	buf.WriteByte(0x80)

	SIGNER_VALUES = buf.Bytes()
	SIGNER_VALUES_LENGTH = len(SIGNER_VALUES)
}
