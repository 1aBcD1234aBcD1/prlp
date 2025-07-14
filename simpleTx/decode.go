package simpleTx

import (
	"github.com/1aBcD1234aBcD1/prlp/errors"
	"github.com/1aBcD1234aBcD1/prlp/reader"
	"github.com/ethereum/go-ethereum/core/types"
	"io"
	"math/big"
)

func DecodeTxsPacket(r *reader.RlpReader) ([]*SimpleTx, error) {
	var txs []*SimpleTx
	// read list length
	listSize, err := r.ReadListSize()
	if err != nil {
		return txs, err
	}
	cPos := r.Pos()
	for r.Pos()-cPos < listSize {
		tx, err := DecodeTx(r)

		if err != nil {
			if errors.Is(err, errors.ErrTxTypeNotSupported) {
				continue
			}
			return txs, err
		}
		txs = append(txs, tx)
	}
	return txs, nil
}

// DecodeTx decodes a transaction from the provided RLP-encoded byte array and returns a SimpleTx instance.
func DecodeTx(r *reader.RlpReader) (*SimpleTx, error) {
	if r.IsNextValAList() {
		return DecodeLegacyTx(r)
	} else {
		return DecodeModernTx(r)
	}
}

// DecodeLegacyTx decodes a legacy transaction from the provided RLP-encoded byte array and returns a CustomTx instance.
func DecodeLegacyTx(r *reader.RlpReader) (*SimpleTx, error) {
	cPos := r.Pos() // store the rlpBytes
	// check for slice length
	bytesLength, err := r.ReadListSize()
	if err != nil {
		return nil, err
	}
	newPos := r.Pos() - cPos
	rlpBytes := r.GetBytes(cPos, cPos+newPos+bytesLength)
	return &SimpleTx{
		TxType:     types.LegacyTxType,
		RLPBytes:   rlpBytes,
		startPoint: 0,
	}, r.Skip(bytesLength)
}

func DecodeModernTx(r *reader.RlpReader) (*SimpleTx, error) {
	// get current point so we can store the rlpbytes
	pos := r.Pos()
	// we already assume that this is another tx type so we just read how many bytes it has
	valLength, err := r.ReadValueSize()
	if err != nil {
		panic(err)
	}
	// check that there are enough bytes to read the tx
	if !r.EnoughBytes(valLength) {
		return nil, io.EOF
	}
	// starting point just indicates from which byte from the rlp needs to read for the tx hash
	startPoint := r.Pos() - pos

	rlpBytes := r.GetBytes(pos, pos+valLength+startPoint)
	txType, err := r.ReadByte()
	switch txType {
	case types.AccessListTxType, types.DynamicFeeTxType, types.SetCodeTxType, types.BlobTxType:
		{
			txListSize, err := r.ReadListSize()
			if err != nil {
				return nil, err
			}

			currentPos := r.Pos()
			chainId, err := r.DecodeNextValue()
			bytesRead := r.Pos() - currentPos

			return &SimpleTx{
				TxType:     txType,
				RLPBytes:   rlpBytes,
				ChainId:    new(big.Int).SetBytes(chainId),
				startPoint: startPoint,
			}, r.Skip(txListSize - bytesRead)
		}
	default:
		// up to this point we have read that it is not a supported tx,
		// so the next thing to do is read the list length and skip the nbytes
		txListSize, err := r.ReadListSize()
		if err != nil {
			return nil, err
		}
		err = r.Skip(txListSize)
		if err != nil {
			return nil, err
		}
		return nil, errors.ErrTxTypeNotSupported
	}

}
