package simpleTx

import (
	"github.com/ethereum/go-ethereum/common"
	"prlp/pool"
)

type SimpleTx struct {
	TxType     byte
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
