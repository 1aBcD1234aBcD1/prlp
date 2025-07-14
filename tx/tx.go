package tx

import (
	"crypto/ecdsa"
	"github.com/1aBcD1234aBcD1/prlp/errors"
	"github.com/1aBcD1234aBcD1/prlp/pool"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"golang.org/x/crypto/sha3"
	"math/big"
)

var zeroHash = common.Hash{}

type CustomTx struct {
	TxType       uint8
	signedHash   []byte
	unsignedHash []byte

	// Some helpers to bytes pointer
	startTx            int // Points where the tx info starts. Used for doing the signed hash
	startTxDataPointer int // Point where the tx data starts.
	startTxSignature   int // Point where the tx signature starts.

	SignedRlpBytes   []byte
	UnsignedRlpBytes []byte // only used when the tx is unsigned. Since it can be helpful to fill the signed tx. Stores only txData not preTxType nor the total length

	// setup bytes since we need to check for length
	from []byte
	// needed for any type
	Nonce    uint64
	GasPrice *big.Int // only available for legacy txs
	Gas      uint64
	To       *common.Address
	Value    *big.Int
	Data     []byte
	V, R, S  *big.Int

	// General arguments for the other txs
	ChainID    *big.Int
	GasTipCap  *big.Int // a.k.a. maxPriorityFeePerGas
	GasFeeCap  *big.Int // a.k.a. maxFeePerGas
	AccessList types.AccessList

	// Blobtxs
	BlobFeeCap *big.Int
	Sidecar    *types.BlobTxSidecar
	BlobHashes []common.Hash
	// SetcodeTxs
	AuthList []types.SetCodeAuthorization
	// helper variable used to indicate where the real information of the tx starts in the rlpbytes slice
	// startPoint               uint64
	// startUnsignedPoint       uint64
	rlpSignedBytesLength int // used to cache the length of the rlpbytes of the entire tx
	rlpSignedBytesTxInfo int // used to cache the length of the rlpbytes of the tx values (without adding the rlp listsize or txtype + valueSize + listsize)

}

func (tx *CustomTx) FromTx(normalTx *types.Transaction) error {
	// copy standard parameters
	tx.Nonce = normalTx.Nonce()
	tx.Gas = normalTx.Gas()
	tx.To = normalTx.To()
	tx.Value = normalTx.Value()
	tx.Data = normalTx.Data()
	tx.TxType = normalTx.Type()

	v, r, s := normalTx.RawSignatureValues()
	tx.V = v
	tx.R = r
	tx.S = s

	switch normalTx.Type() {
	case types.LegacyTxType:
		tx.GasPrice = normalTx.GasPrice()
	case types.AccessListTxType:
		tx.GasPrice = normalTx.GasPrice()
		tx.AccessList = normalTx.AccessList()
		tx.ChainID = normalTx.ChainId()
	case types.DynamicFeeTxType:
		tx.GasFeeCap = normalTx.GasFeeCap()
		tx.GasTipCap = normalTx.GasTipCap()
		tx.AccessList = normalTx.AccessList()
		tx.ChainID = normalTx.ChainId()
	case types.BlobTxType:
		tx.GasFeeCap = normalTx.GasFeeCap()
		tx.GasTipCap = normalTx.GasTipCap()
		tx.AccessList = normalTx.AccessList()
		tx.ChainID = normalTx.ChainId()
		tx.BlobFeeCap = normalTx.BlobGasFeeCap()
		tx.BlobHashes = normalTx.BlobHashes()
	case types.SetCodeTxType:
		tx.GasFeeCap = normalTx.GasFeeCap()
		tx.GasTipCap = normalTx.GasTipCap()
		tx.AccessList = normalTx.AccessList()
		tx.ChainID = normalTx.ChainId()
		tx.AuthList = normalTx.SetCodeAuthorizations()
	}
	return nil
}

func (tx *CustomTx) CalculateRLPLengthSignatureValues() int {
	return CalculateRLBigIntValueLength(tx.V) + CalculateRLBigIntValueLength(tx.R) + CalculateRLBigIntValueLength(tx.S)
}

// Measure the length of a tx including the
func (tx *CustomTx) CalculateRLPSignedBytesLength() (int, int, error) {
	if tx.rlpSignedBytesLength != 0 && tx.rlpSignedBytesTxInfo != 0 {
		return len(tx.SignedRlpBytes), tx.rlpSignedBytesTxInfo, nil
	}
	var (
		l, valsLength int
	)
	switch tx.TxType {
	case types.AccessListTxType:
		valsLength = tx.calculateRLPSignedBytesLenAccessListTx()
		l = CalculateNBytesLength(uint64(CalculateRLPListLength(valsLength) + 1))
	case types.DynamicFeeTxType:
		valsLength = tx.calculateRLPSignedBytesLenDynamicFeesTx()
		l = CalculateNBytesLength(uint64(CalculateRLPListLength(valsLength) + 1))
	case types.LegacyTxType:
		valsLength = tx.calculateRLPSignedBytesLenLegacyTx()
		l = CalculateRLPListLength(valsLength)
	default:
		return 0, 0, errors.ErrTxTypeNotSupported
	}
	tx.rlpSignedBytesTxInfo = l
	tx.rlpSignedBytesLength = valsLength
	return l, valsLength, nil

}

func (tx *CustomTx) calculateRLPAccessListLength() int {
	var length int
	for _, a := range tx.AccessList {
		length += CalculateRLPListLength(tx.calculateRLPAccessTupleLength(a))
	}
	return length
}

func (tx *CustomTx) calculateRLPAccessTupleLength(accessTuple types.AccessTuple) int {
	// this can be precalculated much easier
	// storage keys
	length := CalculateRLPListLength(len(accessTuple.StorageKeys) * HashRLPLength)
	length += AddressRLPLength
	return length
}

func (tx *CustomTx) calculateRLPSignedBytesLenAccessListTx() int {
	var length int
	if len(tx.UnsignedRlpBytes) > 0 {
		length += len(tx.UnsignedRlpBytes)
	} else {
		length += tx.calculateRLPUnSignedBytesLenAccessListTx()
	}
	length += CalculateRLBigIntValueLength(tx.V)
	length += CalculateRLBigIntValueLength(tx.R)
	length += CalculateRLBigIntValueLength(tx.S)
	return length
}
func (tx *CustomTx) calculateRLPUnSignedBytesLenAccessListTx() int {
	var length int
	length += CalculateRLBigIntValueLength(tx.ChainID)
	length += CalculateRLP64ValueLength(tx.Nonce)
	length += CalculateRLBigIntValueLength(tx.GasPrice)
	length += CalculateRLP64ValueLength(tx.Gas)
	if tx.To != nil {
		length += AddressRLPLength
	} else {
		length += 1
	}
	if tx.Value == nil {
		length += 1 // 0x80
	} else {
		length += CalculateRLBigIntValueLength(tx.Value)
	}

	length += CalculateRLPBytesLength(tx.Data)
	length += CalculateRLPListLength(tx.calculateRLPAccessListLength())
	return length
}

func (tx *CustomTx) calculateRLPSignedBytesLenDynamicFeesTx() int {
	var length int
	if len(tx.UnsignedRlpBytes) > 0 {
		length += len(tx.UnsignedRlpBytes)
	} else {
		length += tx.calculateRLPUnSignedBytesLenDynamicFeesTx()
	}
	length += CalculateRLBigIntValueLength(tx.V)
	length += CalculateRLBigIntValueLength(tx.R)
	length += CalculateRLBigIntValueLength(tx.S)
	return length
}

func (tx *CustomTx) calculateRLPUnSignedBytesLenDynamicFeesTx() int {
	var length int
	length += CalculateRLBigIntValueLength(tx.ChainID)
	length += CalculateRLP64ValueLength(tx.Nonce)
	length += CalculateRLBigIntValueLength(tx.GasTipCap)
	length += CalculateRLBigIntValueLength(tx.GasFeeCap)
	length += CalculateRLP64ValueLength(tx.Gas)
	if tx.To != nil {
		length += AddressRLPLength
	} else {
		length += 1
	}
	if tx.Value != nil {
		length += CalculateRLBigIntValueLength(tx.Value)
	} else {
		length += 1
	}
	length += CalculateRLPBytesLength(tx.Data)
	length += CalculateRLPListLength(tx.calculateRLPAccessListLength())
	return length
}

func (tx *CustomTx) calculateRLPSignedBytesLenLegacyTx() int {
	var length int
	if len(tx.UnsignedRlpBytes) > 0 {
		length += len(tx.UnsignedRlpBytes)
	} else {
		length += tx.calculateRLPUnsignedBytesLenLegacyTx()
	}
	length += CalculateRLBigIntValueLength(tx.V)
	length += CalculateRLBigIntValueLength(tx.R)
	length += CalculateRLBigIntValueLength(tx.S)

	return length
}

func (tx *CustomTx) calculateRLPUnsignedBytesLenLegacyTx() int {
	var length int
	length += CalculateRLP64ValueLength(tx.Nonce)
	length += CalculateRLBigIntValueLength(tx.GasPrice)
	length += CalculateRLP64ValueLength(tx.Gas)
	if tx.To != nil {
		length += AddressRLPLength
	} else {
		length += 1
	}
	if tx.Value != nil {
		length += CalculateRLBigIntValueLength(tx.Value)
	} else {
		length += 1
	}
	length += CalculateRLPBytesLength(tx.Data)
	return length
}

func (tx *CustomTx) CalculateSignedHash() common.Hash {
	if len(tx.SignedRlpBytes) == 0 {
		// Save rlp it may be used later
		err := tx.EncodeSignedRLP(pool.GetRLPBuffer(), true)
		if err != nil {
			return zeroHash
		}
	}
	// TODO optimize this with pools
	hasher := sha3.NewLegacyKeccak256().(crypto.KeccakState)
	if tx.TxType == types.LegacyTxType {
		hasher.Write(tx.SignedRlpBytes)
	} else {
		hasher.Write(tx.SignedRlpBytes[tx.startTx:])
	}
	tx.signedHash = make([]byte, 32)
	hasher.Read(tx.signedHash)
	return common.BytesToHash(tx.signedHash)

}

func (tx *CustomTx) UnsignedHash() common.Hash {
	if len(tx.unsignedHash) == 0 {
		return tx.CalculateUnsignedHash()
	}
	return common.BytesToHash(tx.unsignedHash)
}

func (tx *CustomTx) Hash() common.Hash {
	if len(tx.signedHash) == 0 {
		return tx.CalculateSignedHash()
	}
	return common.BytesToHash(tx.signedHash)
}

func (tx *CustomTx) From() (common.Address, error) {
	if len(tx.from) != 0 {
		return common.BytesToAddress(tx.from), nil
	} else {
		var err error
		switch tx.TxType {
		case types.LegacyTxType:
			tx.from, err = tx.getFromLegacyTx()
			return common.BytesToAddress(tx.from), err
		case types.DynamicFeeTxType, types.AccessListTxType:
			tx.from, err = tx.getFromOtherTxTypes()
			return common.BytesToAddress(tx.from), err
		default:
			return common.Address{}, errors.ErrTxTypeNotSupported
		}
	}

}

func (tx *CustomTx) getFromLegacyTx() ([]byte, error) {
	V := byte(tx.V.Uint64() - V_MULTIPLER_UINT64 - 8 - 27)
	//V := c.v[0]
	if !validateSignatureValues(V, tx.R, tx.S, false) {
		return []byte{}, errors.ErrInvalidSig
	}
	// encode the signature in uncompressed format
	sig := make([]byte, crypto.SignatureLength)

	r := tx.R.Bytes()
	s := tx.S.Bytes()
	copy(sig[32-len(r):32], r)
	copy(sig[64-len(s):64], s)
	sig[64] = V
	// recover the public key from the signature
	pub, err := crypto.Ecrecover(tx.UnsignedHash().Bytes(), sig)
	if err != nil {
		return []byte{}, err
	}
	if len(pub) == 0 || pub[0] != 4 {
		return []byte{}, errors.ErrInvalidPkb
	}
	return crypto.Keccak256(pub[1:])[12:], nil
}

func (tx *CustomTx) getFromOtherTxTypes() ([]byte, error) {
	V := byte(len(tx.V.Bytes())) // always 0 or 1
	if !validateSignatureValues(V, tx.R, tx.S, false) {
		return []byte{}, errors.ErrInvalidSig
	}
	r := tx.R.Bytes()
	s := tx.S.Bytes()
	// encode the signature in uncompressed format
	sig := make([]byte, crypto.SignatureLength)
	copy(sig[32-len(r):32], r)
	copy(sig[64-len(s):64], s)
	sig[64] = V
	// recover the public key from the signature
	pub, err := crypto.Ecrecover(tx.UnsignedHash().Bytes(), sig)
	if err != nil {
		return []byte{}, err
	}
	if len(pub) == 0 || pub[0] != 4 {
		return []byte{}, errors.ErrInvalidPkb
	}
	//copy(addr[:], crypto.Keccak256(pub[1:])[12:])
	//hash, _ := c.GetSignedHashBytes()
	//fmt.Printf("Hash: %x, \t From 0x%x: \n", hash, crypto.Keccak256(pub[1:])[12:])
	return crypto.Keccak256(pub[1:])[12:], nil
}

func (tx *CustomTx) CalculateUnsignedHash() common.Hash {
	hasher := pool.GetHasher()
	buffer := pool.GetRLPBuffer()
	defer pool.PutHasher(hasher)
	defer pool.PutRLPBuffer(buffer)

	switch tx.TxType {
	case types.LegacyTxType:
		tx.EncodeUnsignedLegacyTx(buffer)
	case types.DynamicFeeTxType:
		tx.EncodeUnsignedDynamicFeesTx(buffer)
	case types.AccessListTxType:
		tx.EncodeUnsignedAccessListTx(buffer)
	default:
		return zeroHash
	}
	buffer.WriteTo(hasher)
	tx.unsignedHash = make([]byte, 32)
	hasher.Read(tx.unsignedHash[:])
	return common.BytesToHash(tx.unsignedHash)
}

func (tx *CustomTx) SignTx(key *ecdsa.PrivateKey) error {
	hasher := pool.GetHasher()
	defer pool.PutHasher(hasher)

	h := tx.UnsignedHash()
	sig, err := crypto.Sign(h.Bytes(), key)
	if err != nil {
		return err
	}

	tx.R = new(big.Int).SetBytes(sig[:32])
	tx.S = new(big.Int).SetBytes(sig[32:64])
	tx.V = new(big.Int)
	switch tx.TxType {
	case types.LegacyTxType:
		tx.V.Add(big.NewInt(int64(sig[64]+35)), V_MULTIPLIER)
	default:
		if sig[64] > 0 {
			tx.V.SetUint64(uint64(sig[64]))
		}
	}
	return nil
}

// ResetSignedVals clear and empties the values that store any kind of signed values:
// - V, R, S values
// - SignedRlpBytes
// - signedHash
func (tx *CustomTx) ResetSignedVals() {
	tx.V = nil
	tx.R = nil
	tx.S = nil
	tx.signedHash = []byte{}
	tx.SignedRlpBytes = []byte{}
}
