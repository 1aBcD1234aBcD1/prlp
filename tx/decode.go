package tx

import (
	"github.com/1aBcD1234aBcD1/prlp/reader"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/holiman/uint256"
	"io"
	"math/big"
)

// DecodeTxsPacket decodes a list of transactions from the provided RlpReader and returns them as a slice of CustomTx.
// Returns an error if decoding fails.
func DecodeTxsPacket(r *reader.RlpReader) ([]*CustomTx, error) {
	var txs []*CustomTx
	// read list length
	listSize, err := r.ReadListSize()
	if err != nil {
		return nil, err
	}
	cPos := r.Pos()
	for r.Pos()-cPos < listSize {
		if r.IsNextValAList() {
			tx, err := DecodeLegacyTx(r)
			if err != nil {
				return nil, err
			}
			txs = append(txs, tx)
		} else {
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
			case types.AccessListTxType:
				{
					tx, err := DecodeAccessListTx(r, rlpBytes, startPoint)
					if err != nil {
						return nil, err
					}
					txs = append(txs, tx)
				}
			case types.DynamicFeeTxType:
				{
					tx, err := DecodeDynamicFeeTx(r, rlpBytes, startPoint)
					if err != nil {
						return nil, err
					}
					txs = append(txs, tx)
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
				continue
			}
		}
	}
	return txs, nil
}

// DecodeSetCodeAuthorization parses an RLP-encoded payload into a SetCodeAuthorization struct.
// It reads and decodes data such as chain ID, address, nonce, and signature values.
// Returns the decoded SetCodeAuthorization on success, or an error if decoding fails.
func DecodeSetCodeAuthorization(r *reader.RlpReader) (setCodeAuthorization types.SetCodeAuthorization, err error) {
	codeAuthSize, err := r.ReadListSize()
	if err != nil {
		return setCodeAuthorization, err
	}
	cPos := r.Pos()
	for r.Pos()-cPos < codeAuthSize {
		chainId, err := r.DecodeNextValue()
		if err != nil {
			return setCodeAuthorization, err
		}
		address, err := r.DecodeNextValue()
		if err != nil {
			return setCodeAuthorization, err
		}
		nonce, err := r.DecodeNextValue()
		if err != nil {
			return setCodeAuthorization, err
		}
		v, err := r.DecodeNextValue()
		if err != nil {
			return setCodeAuthorization, err
		}
		rVal, err := r.DecodeNextValue()
		if err != nil {
			return setCodeAuthorization, err
		}
		s, err := r.DecodeNextValue()
		if err != nil {
			return setCodeAuthorization, err
		}
		setCodeAuthorization.ChainID = *uint256.MustFromBig(new(big.Int).SetBytes(chainId))
		setCodeAuthorization.Address = common.BytesToAddress(address)
		setCodeAuthorization.Nonce = reader.BytesToUint64(nonce)
		if len(v) == 0 {
			setCodeAuthorization.V = 0
		} else {
			setCodeAuthorization.V = v[len(v)-1]
		}

		setCodeAuthorization.R = *uint256.MustFromBig(new(big.Int).SetBytes(rVal))
		setCodeAuthorization.S = *uint256.MustFromBig(new(big.Int).SetBytes(s))
	}
	return setCodeAuthorization, err
}

// DecodeDecodeSetCodeAuthorizationList parses an RLP-encoded payload into a SetCodeAuthorization struct.
// It reads and decodes data such as chain ID, address, nonce, and signature values.
// Returns the decoded SetCodeAuthorization on success, or an error if decoding fails.
func DecodeDecodeSetCodeAuthorizationList(r *reader.RlpReader) (list []types.SetCodeAuthorization, err error) {
	list = make([]types.SetCodeAuthorization, 0)
	listSize, err := r.ReadListSize()
	if err != nil {
		return list, err
	}
	cPos := r.Pos()
	for r.Pos()-cPos < listSize {
		setCodeAuthroization, err := DecodeSetCodeAuthorization(r)
		if err != nil {
			return list, err
		}
		list = append(list, setCodeAuthroization)
	}
	return list, err
}

// DecodeAccessTuple decodes an RLP-encoded access tuple from the provided RlpReader.
// It returns the decoded access tuple and any error encountered during parsing.
func DecodeAccessTuple(r *reader.RlpReader) (accessTuple types.AccessTuple, err error) {
	accessTupleSize, err := r.ReadListSize()
	if err != nil {
		return accessTuple, err
	}
	cPos := r.Pos()
	for r.Pos()-cPos < accessTupleSize {
		address, err := r.DecodeNextValue()
		if err != nil {
			return accessTuple, err
		}
		accessTuple.Address = common.BytesToAddress(address)
		storageKeysSize, err := r.ReadListSize()
		if err != nil {
			return accessTuple, err
		}
		cStorageKeysPos := r.Pos()
		for r.Pos()-cStorageKeysPos < storageKeysSize {
			storageKey, err := r.DecodeNextValue()
			if err != nil {
				return accessTuple, err
			}
			accessTuple.StorageKeys = append(accessTuple.StorageKeys, common.BytesToHash(storageKey))
		}
	}
	return accessTuple, err
}

func DecodeAccessList(r *reader.RlpReader) (accessList types.AccessList, err error) {
	accessList = make(types.AccessList, 0)
	accessListSize, err := r.ReadListSize()
	if err != nil {
		return accessList, err
	}
	cPos := r.Pos()
	for r.Pos()-cPos < accessListSize {
		accessTuple, err := DecodeAccessTuple(r)
		if err != nil {
			return accessList, err
		}
		accessList = append(accessList, accessTuple)
	}
	return accessList, err
}

// DecodeLegacyTx decodes a legacy transaction from the provided RLP-encoded byte array and returns a CustomTx instance.
func DecodeLegacyTx(tx *reader.RlpReader) (*CustomTx, error) {
	// store where does the hashing data for signed hash starts
	// by legacyTx are at 0
	// store where does the

	cPos := tx.Pos() // store the rlpBytes
	// check for slice length
	bytesLength, err := tx.ReadListSize()
	if err != nil {
		return nil, err
	}
	// store where does the txData starts
	startTxDataPointer := tx.Pos() - cPos
	// TODO move this outside the function?
	rlpBytes := tx.GetBytes(cPos, cPos+startTxDataPointer+bytesLength)
	rlpBytesLength := len(rlpBytes)
	rlpBytesTxInfo := rlpBytesLength - int(startTxDataPointer)

	nonce, err := tx.DecodeNextValue()
	if err != nil {
		return nil, err
	}
	gasPrice, err := tx.DecodeNextValue()
	if err != nil {
		return nil, err
	}
	gas, err := tx.DecodeNextValue()
	if err != nil {
		return nil, err
	}
	toBytes, err := tx.DecodeNextValue()
	if err != nil {
		return nil, err
	}
	var to *common.Address
	if len(toBytes) > 0 {
		to = new(common.Address)
		to.SetBytes(toBytes)
	}
	value, err := tx.DecodeNextValue()
	if err != nil {
		return nil, err
	}
	data, err := tx.DecodeNextValue()
	if err != nil {
		return nil, err
	}
	startTxSignature := tx.Pos() - cPos
	v, err := tx.DecodeNextValue()
	if err != nil {
		return nil, err
	}
	r, err := tx.DecodeNextValue()
	if err != nil {
		return nil, err
	}
	s, err := tx.DecodeNextValue()
	if err != nil {
		return nil, err
	}
	return &CustomTx{
		TxType:               types.LegacyTxType,
		SignedRlpBytes:       rlpBytes,
		Nonce:                reader.BytesToUint64(nonce),
		GasPrice:             new(big.Int).SetBytes(gasPrice),
		Gas:                  reader.BytesToUint64(gas),
		To:                   to,
		Value:                new(big.Int).SetBytes(value),
		Data:                 data,
		V:                    new(big.Int).SetBytes(v),
		R:                    new(big.Int).SetBytes(r),
		S:                    new(big.Int).SetBytes(s),
		startTx:              0,
		startTxDataPointer:   int(startTxDataPointer),
		startTxSignature:     int(startTxSignature),
		rlpSignedBytesLength: len(rlpBytes),
		rlpSignedBytesTxInfo: rlpBytesTxInfo,
	}, nil
}

// DecodeAccessListTx decodes an access list transaction from RLP encoded bytes using the provided RlpReader.
// rlpBytes fields provides all the bytes of the rlp of the tx in the wire.
// starPoint indicates where the tx info starts in the rlpBytes slice. Needed to calculate the hash
func DecodeAccessListTx(tx *reader.RlpReader, rlpBytes []byte, startPoint uint64) (*CustomTx, error) {
	// start point will always be txvalsize info - txType, where the position of txType is startPoint.
	cPos := tx.Pos() - startPoint - 1 // after the startpoint we read one byte, thats why the - 1
	_, err := tx.ReadListSize()
	if err != nil {
		return nil, err
	}
	startTxDataPointer := tx.Pos() - cPos
	rlpBytesLength := len(rlpBytes)
	chainId, err := tx.DecodeNextValue()
	if err != nil {
		return nil, err
	}
	nonce, err := tx.DecodeNextValue()
	if err != nil {
		return nil, err
	}
	gasPrice, err := tx.DecodeNextValue()
	if err != nil {
		return nil, err
	}
	gas, err := tx.DecodeNextValue()
	if err != nil {
		return nil, err
	}
	toBytes, err := tx.DecodeNextValue()
	if err != nil {
		return nil, err
	}
	var to *common.Address
	if len(toBytes) > 0 {
		to = new(common.Address)
		to.SetBytes(toBytes)
	}
	value, err := tx.DecodeNextValue()
	if err != nil {
		return nil, err
	}
	data, err := tx.DecodeNextValue()
	if err != nil {
		return nil, err
	}
	var accessList types.AccessList
	accessList, err = DecodeAccessList(tx)
	if err != nil {
		return nil, err
	}
	startTxSignature := tx.Pos() - cPos
	v, err := tx.DecodeNextValue()
	if err != nil {
		return nil, err
	}
	r, err := tx.DecodeNextValue()
	if err != nil {
		return nil, err
	}
	s, err := tx.DecodeNextValue()
	if err != nil {
		return nil, err
	}
	return &CustomTx{
		TxType:               types.AccessListTxType,
		SignedRlpBytes:       rlpBytes,
		ChainID:              new(big.Int).SetBytes(chainId),
		Nonce:                reader.BytesToUint64(nonce),
		GasPrice:             new(big.Int).SetBytes(gasPrice),
		Gas:                  reader.BytesToUint64(gas),
		To:                   to,
		Value:                new(big.Int).SetBytes(value),
		Data:                 data,
		V:                    new(big.Int).SetBytes(v),
		R:                    new(big.Int).SetBytes(r),
		S:                    new(big.Int).SetBytes(s),
		AccessList:           accessList,
		rlpSignedBytesLength: rlpBytesLength,
		rlpSignedBytesTxInfo: rlpBytesLength - int(startTxDataPointer),
		startTx:              int(startPoint),
		startTxDataPointer:   int(startTxDataPointer),
		startTxSignature:     int(startTxSignature),
	}, nil
}

// DecodeDynamicFeeTx decodes a dynamic fee transaction from RLP encoded bytes using the provided RlpReader.
// rlpBytes fields provides all the bytes of the rlp of the tx in the wire.
// starPoint indicates where the tx info starts in the rlpBytes slice. Needed to calculate the hash
func DecodeDynamicFeeTx(tx *reader.RlpReader, rlpBytes []byte, startPoint uint64) (*CustomTx, error) {
	// start point will always be txvalsize info - txType, where the position of txType is startPoint.
	cPos := tx.Pos() - startPoint - 1 // after the startpoint we read one byte, thats why the - 1
	_, err := tx.ReadListSize()
	if err != nil {
		return nil, err
	}
	startTxDataPointer := tx.Pos() - cPos
	rlpBytesLength := len(rlpBytes)
	chainId, err := tx.DecodeNextValue()
	if err != nil {
		return nil, err
	}
	nonce, err := tx.DecodeNextValue()
	if err != nil {
		return nil, err
	}
	gasTipCap, err := tx.DecodeNextValue()
	if err != nil {
		return nil, err
	}
	gasFeeCap, err := tx.DecodeNextValue()
	if err != nil {
		return nil, err
	}
	gas, err := tx.DecodeNextValue()
	if err != nil {
		return nil, err
	}
	toBytes, err := tx.DecodeNextValue()
	if err != nil {
		return nil, err
	}
	var to *common.Address
	if len(toBytes) > 0 {
		to = new(common.Address)
		to.SetBytes(toBytes)
	}
	value, err := tx.DecodeNextValue()
	if err != nil {
		return nil, err
	}
	data, err := tx.DecodeNextValue()
	if err != nil {
		return nil, err
	}
	var accessList types.AccessList
	accessList, err = DecodeAccessList(tx)
	if err != nil {
		return nil, err
	}
	startTxSignature := tx.Pos() - cPos
	v, err := tx.DecodeNextValue()
	if err != nil {
		return nil, err
	}
	r, err := tx.DecodeNextValue()
	if err != nil {
		return nil, err
	}
	s, err := tx.DecodeNextValue()
	if err != nil {
		return nil, err
	}
	return &CustomTx{
		TxType:               types.DynamicFeeTxType,
		SignedRlpBytes:       rlpBytes,
		ChainID:              new(big.Int).SetBytes(chainId),
		Nonce:                reader.BytesToUint64(nonce),
		GasTipCap:            new(big.Int).SetBytes(gasTipCap),
		GasFeeCap:            new(big.Int).SetBytes(gasFeeCap),
		Gas:                  reader.BytesToUint64(gas),
		To:                   to,
		Value:                new(big.Int).SetBytes(value),
		Data:                 data,
		V:                    new(big.Int).SetBytes(v),
		R:                    new(big.Int).SetBytes(r),
		S:                    new(big.Int).SetBytes(s),
		AccessList:           accessList,
		rlpSignedBytesLength: rlpBytesLength,
		rlpSignedBytesTxInfo: rlpBytesLength - int(startTxDataPointer),
		startTx:              int(startPoint),
		startTxDataPointer:   int(startTxDataPointer),
		startTxSignature:     int(startTxSignature),
	}, nil
}

// DecodeDynamicFeeTx decodes an dynamic fee transaction from RLP encoded bytes using the provided RlpReader.
// rlpBytes fields provides all the bytes of the rlp of the tx in the wire.
// starPoint indicates where the tx info starts in the rlpBytes slice. Needed to calculate the hash
func DecodeSetCodeTx(tx *reader.RlpReader, rlpBytes []byte, startPoint uint64) (*CustomTx, error) {
	// start point will always be txvalsize info - txType, where the position of txType is startPoint.
	cPos := tx.Pos() - startPoint - 1 // after the startpoint we read one byte, thats why the - 1
	_, err := tx.ReadListSize()
	if err != nil {
		return nil, err
	}
	startTxDataPointer := tx.Pos() - cPos
	rlpBytesLength := len(rlpBytes)
	chainId, err := tx.DecodeNextValue()
	if err != nil {
		return nil, err
	}
	nonce, err := tx.DecodeNextValue()
	if err != nil {
		return nil, err
	}
	gasTipCap, err := tx.DecodeNextValue()
	if err != nil {
		return nil, err
	}
	gasFeeCap, err := tx.DecodeNextValue()
	if err != nil {
		return nil, err
	}
	gas, err := tx.DecodeNextValue()
	if err != nil {
		return nil, err
	}
	toBytes, err := tx.DecodeNextValue()
	if err != nil {
		return nil, err
	}
	var to *common.Address
	if len(toBytes) > 0 {
		to = new(common.Address)
		to.SetBytes(toBytes)
	}
	value, err := tx.DecodeNextValue()
	if err != nil {
		return nil, err
	}
	data, err := tx.DecodeNextValue()
	if err != nil {
		return nil, err
	}
	var accessList types.AccessList
	accessList, err = DecodeAccessList(tx)
	if err != nil {
		return nil, err
	}
	var authList []types.SetCodeAuthorization
	authList, err = DecodeDecodeSetCodeAuthorizationList(tx)
	if err != nil {
		return nil, err
	}
	startTxSignature := tx.Pos() - cPos
	v, err := tx.DecodeNextValue()
	if err != nil {
		return nil, err
	}
	r, err := tx.DecodeNextValue()
	if err != nil {
		return nil, err
	}
	s, err := tx.DecodeNextValue()
	if err != nil {
		return nil, err
	}
	return &CustomTx{
		TxType:               types.SetCodeTxType,
		SignedRlpBytes:       rlpBytes,
		ChainID:              new(big.Int).SetBytes(chainId),
		Nonce:                reader.BytesToUint64(nonce),
		GasTipCap:            new(big.Int).SetBytes(gasTipCap),
		GasFeeCap:            new(big.Int).SetBytes(gasFeeCap),
		Gas:                  reader.BytesToUint64(gas),
		To:                   to,
		Value:                new(big.Int).SetBytes(value),
		Data:                 data,
		V:                    new(big.Int).SetBytes(v),
		R:                    new(big.Int).SetBytes(r),
		S:                    new(big.Int).SetBytes(s),
		AccessList:           accessList,
		AuthList:             authList,
		rlpSignedBytesLength: rlpBytesLength,
		rlpSignedBytesTxInfo: rlpBytesLength - int(startTxDataPointer),
		startTx:              int(startPoint),
		startTxDataPointer:   int(startTxDataPointer),
		startTxSignature:     int(startTxSignature),
	}, nil
}
