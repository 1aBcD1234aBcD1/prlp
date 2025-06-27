package prlp

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"golang.org/x/crypto/sha3"
	"math/big"
)

var zeroHash = common.Hash{}

type CustomTx struct {
	TxType uint8

	signedHash   common.Hash
	unsignedHash common.Hash

	RlpBytes         []byte
	UnsignedRlpBytes []byte // Just needed when we are generating a transaction

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
	startPoint uint64
}

func (tx *CustomTx) CalculateSignedHash() (h common.Hash) {
	if len(tx.RlpBytes) == 0 {
		panic("implement me")
		return
	} else {
		// TODO optimize this with pools
		hasher := sha3.NewLegacyKeccak256().(crypto.KeccakState)
		if tx.TxType == types.LegacyTxType {
			hasher.Write(tx.RlpBytes)
		} else {
			hasher.Write(tx.RlpBytes[tx.startPoint:])
		}
		hasher.Read(h[:])
		return h
	}
}

func (tx *CustomTx) Hash() common.Hash {
	if tx.signedHash == zeroHash {
		return tx.CalculateSignedHash()
	} else {
		return tx.signedHash
	}
}
