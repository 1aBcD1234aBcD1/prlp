package prlp

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/holiman/uint256"
	"math/big"
)

func DecodeSetCodeAuthorization(r *RlpReader) (setCodeAuthorization types.SetCodeAuthorization, err error) {
	codeAuthSize, err := r.ReadListSize()
	if err != nil {
		return setCodeAuthorization, err
	}
	cPos := r.currentPos
	for r.currentPos-cPos < codeAuthSize {
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
		setCodeAuthorization.Nonce = BytesToUint64(nonce)
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

func DecodeDecodeSetCodeAuthorizationList(r *RlpReader) (list []types.SetCodeAuthorization, err error) {
	list = make([]types.SetCodeAuthorization, 0)
	listSize, err := r.ReadListSize()
	if err != nil {
		return list, err
	}
	cPos := r.currentPos
	for r.currentPos-cPos < listSize {
		setCodeAuthroization, err := DecodeSetCodeAuthorization(r)
		if err != nil {
			return list, err
		}
		list = append(list, setCodeAuthroization)
	}
	return list, err
}
func DecodeAccessTuple(r *RlpReader) (accessTuple types.AccessTuple, err error) {
	accessTupleSize, err := r.ReadListSize()
	if err != nil {
		return accessTuple, err
	}
	cPos := r.currentPos
	for r.currentPos-cPos < accessTupleSize {
		address, err := r.DecodeNextValue()
		if err != nil {
			return accessTuple, err
		}
		accessTuple.Address = common.BytesToAddress(address)
		storageKeysSize, err := r.ReadListSize()
		if err != nil {
			return accessTuple, err
		}
		cStorageKeysPos := r.currentPos
		for r.currentPos-cStorageKeysPos < storageKeysSize {
			storageKey, err := r.DecodeNextValue()
			if err != nil {
				return accessTuple, err
			}
			accessTuple.StorageKeys = append(accessTuple.StorageKeys, common.BytesToHash(storageKey))
		}
	}
	return accessTuple, err
}

func DecodeAccessList(r *RlpReader) (accessList types.AccessList, err error) {
	accessList = make(types.AccessList, 0)
	accessListSize, err := r.ReadListSize()
	if err != nil {
		return accessList, err
	}
	cPos := r.currentPos
	for r.currentPos-cPos < accessListSize {
		accessTuple, err := DecodeAccessTuple(r)
		if err != nil {
			return accessList, err
		}
		accessList = append(accessList, accessTuple)
	}
	return accessList, err
}

// DecodeLegacyTx decodes a legacy transaction from the provided RLP-encoded byte array and returns a CustomTx instance.
func DecodeLegacyTx(tx *RlpReader) (*CustomTx, error) {
	rlpBytes := tx.bytes // store the rlpBytes
	// check for slice length
	_, err := tx.ReadListSize()
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
		RlpBytes: rlpBytes,
		Nonce:    BytesToUint64(nonce),
		GasPrice: new(big.Int).SetBytes(gasPrice),
		Gas:      BytesToUint64(gas),
		To:       to,
		Value:    new(big.Int).SetBytes(value),
		Data:     data,
		V:        new(big.Int).SetBytes(v),
		R:        new(big.Int).SetBytes(r),
		S:        new(big.Int).SetBytes(s),
	}, nil
}

// DecodeAccessListTx decodes an access list transaction from RLP encoded bytes using the provided RlpReader.
// rlpBytes fields provides all the bytes of the rlp of the tx in the wire.
// starPoint indicates where the tx info starts in the rlpBytes slice. Needed to calculate the hash
func DecodeAccessListTx(tx *RlpReader, rlpBytes []byte, startPoint uint64) (*CustomTx, error) {

	//rlpBytes := rlpBytes // store the rlpBytes
	// check for slice length
	_, err := tx.ReadListSize()
	if err != nil {
		return nil, err
	}

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
		TxType:     types.AccessListTxType,
		RlpBytes:   rlpBytes,
		ChainID:    new(big.Int).SetBytes(chainId),
		Nonce:      BytesToUint64(nonce),
		GasPrice:   new(big.Int).SetBytes(gasPrice),
		Gas:        BytesToUint64(gas),
		To:         to,
		Value:      new(big.Int).SetBytes(value),
		Data:       data,
		V:          new(big.Int).SetBytes(v),
		R:          new(big.Int).SetBytes(r),
		S:          new(big.Int).SetBytes(s),
		AccessList: accessList,
		startPoint: startPoint,
	}, nil
}

// DecodeDynamicFeeTx decodes an dynamic fee transaction from RLP encoded bytes using the provided RlpReader.
// rlpBytes fields provides all the bytes of the rlp of the tx in the wire.
// starPoint indicates where the tx info starts in the rlpBytes slice. Needed to calculate the hash
func DecodeDynamicFeeTx(tx *RlpReader, rlpBytes []byte, startPoint uint64) (*CustomTx, error) {

	_, err := tx.ReadListSize()
	if err != nil {
		return nil, err
	}

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
		TxType:     types.DynamicFeeTxType,
		RlpBytes:   rlpBytes,
		ChainID:    new(big.Int).SetBytes(chainId),
		Nonce:      BytesToUint64(nonce),
		GasTipCap:  new(big.Int).SetBytes(gasTipCap),
		GasFeeCap:  new(big.Int).SetBytes(gasFeeCap),
		Gas:        BytesToUint64(gas),
		To:         to,
		Value:      new(big.Int).SetBytes(value),
		Data:       data,
		V:          new(big.Int).SetBytes(v),
		R:          new(big.Int).SetBytes(r),
		S:          new(big.Int).SetBytes(s),
		AccessList: accessList,
		startPoint: startPoint,
	}, nil
}

// DecodeDynamicFeeTx decodes an dynamic fee transaction from RLP encoded bytes using the provided RlpReader.
// rlpBytes fields provides all the bytes of the rlp of the tx in the wire.
// starPoint indicates where the tx info starts in the rlpBytes slice. Needed to calculate the hash
func DecodeSetCodeTx(tx *RlpReader, rlpBytes []byte, startPoint uint64) (*CustomTx, error) {

	_, err := tx.ReadListSize()
	if err != nil {
		return nil, err
	}

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
		TxType:     types.SetCodeTxType,
		RlpBytes:   rlpBytes,
		ChainID:    new(big.Int).SetBytes(chainId),
		Nonce:      BytesToUint64(nonce),
		GasTipCap:  new(big.Int).SetBytes(gasTipCap),
		GasFeeCap:  new(big.Int).SetBytes(gasFeeCap),
		Gas:        BytesToUint64(gas),
		To:         to,
		Value:      new(big.Int).SetBytes(value),
		Data:       data,
		V:          new(big.Int).SetBytes(v),
		R:          new(big.Int).SetBytes(r),
		S:          new(big.Int).SetBytes(s),
		AccessList: accessList,
		AuthList:   authList,
		startPoint: startPoint,
	}, nil
}
