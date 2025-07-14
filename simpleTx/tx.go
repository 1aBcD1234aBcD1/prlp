package simpleTx

import (
	"github.com/1aBcD1234aBcD1/prlp/pool"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

type SimpleTx struct {
	TxType byte

	ChainId    *big.Int
	RLPBytes   []byte
	hash       []byte
	startPoint uint64 // Used to know when hashing from which part of the RLPBytes it should start
}

func (tx *SimpleTx) Hash() common.Hash {
	if len(tx.hash) == 0 {
		tx.hash = pool.HashData(tx.RLPBytes[tx.startPoint:])
	}
	return common.BytesToHash(tx.hash)
}
