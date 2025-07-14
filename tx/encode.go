package tx

import (
	"bytes"
	"github.com/ethereum/go-ethereum/core/types"
	"prlp/errors"
)

func EncodeTxsPacket(buffer *bytes.Buffer, txs []*CustomTx) error {
	var listLength int

	for _, tx := range txs {
		l, _, _ := tx.CalculateRLPSignedBytesLength()
		listLength += l
	}

	_, err := WriteListLength(buffer, listLength)
	if err != nil {
		return err
	}

	for _, tx := range txs {
		err = tx.EncodeSignedRLP(buffer, false)
		if err != nil {
			return err
		}
	}
	return nil
}

func (tx *CustomTx) EncodeSignedRLP(buffer *bytes.Buffer, save bool) error {
	if len(tx.SignedRlpBytes) > 0 {
		_, err := buffer.Write(tx.SignedRlpBytes)
		return err
	}
	switch tx.TxType {
	case types.LegacyTxType:
		return tx.EncodeSignedLegacyTx(buffer, save)
	case types.DynamicFeeTxType:
		return tx.EncodeSignedDynamicFeesTx(buffer, save)
	case types.AccessListTxType:
		return tx.EncodeSignedAccessListTx(buffer, save)
	default:
		return errors.ErrTxTypeNotSupported
	}
}

func (tx *CustomTx) EncodeSignedLegacyTx(buffer *bytes.Buffer, save bool) (err error) {
	var txValsLength int
	var listValLength int
	if len(tx.UnsignedRlpBytes) > 0 {
		txValsLength = len(tx.UnsignedRlpBytes) + tx.CalculateRLPLengthSignatureValues()
		listValLength, err = WriteListLength(buffer, txValsLength)
		if err != nil {
			return err
		}
		buffer.Write(tx.UnsignedRlpBytes)
	} else {

		_, txValsLength, err = tx.CalculateRLPSignedBytesLength()
		if err != nil {
			return err
		}

		listValLength, err = WriteListLength(buffer, txValsLength)
		if err != nil {
			return err
		}
		// we have already written the length indicating list length of the tx
		// now we have to write every value.
		err = WriteRLPUint64(buffer, tx.Nonce)
		if err != nil {
			return err
		}

		err = WriteRLPBytes(buffer, tx.GasPrice.Bytes())
		if err != nil {
			return err
		}

		err = WriteRLPUint64(buffer, tx.Gas)
		if err != nil {
			return err
		}

		if tx.To != nil {
			err = WriteRLPBytes(buffer, tx.To.Bytes())
		} else {
			err = buffer.WriteByte(ZeroUint64RLPVal)
		}
		if err != nil {
			return err
		}

		if tx.Value == nil {
			err = buffer.WriteByte(ZeroUint64RLPVal)
		} else {
			err = WriteRLPBytes(buffer, tx.Value.Bytes())
			if err != nil {
				return err
			}
		}

		err = WriteRLPBytes(buffer, tx.Data)
		if err != nil {
			return err
		}

	}
	// write the signature
	err = WriteRLPBytes(buffer, tx.V.Bytes())
	if err != nil {
		return err
	}
	err = WriteRLPBytes(buffer, tx.R.Bytes())
	if err != nil {
		return err
	}
	err = WriteRLPBytes(buffer, tx.S.Bytes())

	bufferBytes := buffer.Bytes()
	if save {
		tx.SignedRlpBytes = make([]byte, listValLength+txValsLength)
		copy(tx.SignedRlpBytes, bufferBytes[len(bufferBytes)-listValLength-txValsLength:])
	}
	tx.rlpSignedBytesLength = listValLength + txValsLength
	return err
}

func (tx *CustomTx) EncodeAccessTuple(buffer *bytes.Buffer, accessTuple types.AccessTuple) error {
	accessTupleLength := tx.calculateRLPAccessTupleLength(accessTuple)
	_, err := WriteListLength(buffer, accessTupleLength)
	if err != nil {
		return err
	}
	err = buffer.WriteByte(EncodedAddressRLPLength)
	if err != nil {
		return err
	}
	_, err = buffer.Write(accessTuple.Address.Bytes())
	if err != nil {
		return err
	}
	_, err = WriteListLength(buffer, len(accessTuple.StorageKeys)*HashRLPLength)

	for _, h := range accessTuple.StorageKeys {
		err = buffer.WriteByte(EncodedHashRLPLength)
		if err != nil {
			return err
		}
		_, err = buffer.Write(h.Bytes())
		if err != nil {
			return err
		}
	}
	return nil
}

func (tx *CustomTx) EncodeAccessList(buffer *bytes.Buffer) error {
	accessListLength := tx.calculateRLPAccessListLength()
	_, err := WriteListLength(buffer, accessListLength)
	if err != nil {
		return err
	}
	for _, a := range tx.AccessList {
		err = tx.EncodeAccessTuple(buffer, a)
		if err != nil {
			return err
		}
	}
	return nil
}

func (tx *CustomTx) EncodeSignedDynamicFeesTx(buffer *bytes.Buffer, save bool) error {

	var totalRLPLength int

	if len(tx.UnsignedRlpBytes) > 0 {
		txValsLength := len(tx.UnsignedRlpBytes) + tx.CalculateRLPLengthSignatureValues()
		rlpValsLength, err := WriteValLength(buffer, CalculateRLPListLength(txValsLength)+1)
		if err != nil {
			return err
		}
		totalRLPLength += rlpValsLength
		buffer.WriteByte(tx.TxType)
		totalRLPLength += 1
		rlpListLength, err := WriteListLength(buffer, txValsLength)
		if err != nil {
			return err
		}
		totalRLPLength += rlpListLength // this add the nBytes used to write the list of the tx data
		totalRLPLength += txValsLength  // Add the length of the tx values
		buffer.Write(tx.UnsignedRlpBytes)
		// add pointer to let the hasher from which part it should start
		tx.startTx = rlpValsLength
	} else {
		// length of the tx values + the signature vals (v,r,s)
		txValsLength := tx.calculateRLPSignedBytesLenDynamicFeesTx()

		// write first the rlp value of the txtype + txvals
		rlpValsLength, err := WriteValLength(buffer, CalculateRLPListLength(txValsLength)+1)
		if err != nil {
			return err
		}
		totalRLPLength += rlpValsLength // this adds the nBytes used to write the size of the tx
		// write the txtype
		buffer.WriteByte(tx.TxType)
		totalRLPLength += 1

		rlpListLength, err := WriteListLength(buffer, txValsLength)
		if err != nil {
			return err
		}
		totalRLPLength += rlpListLength // this add the nBytes used to write the list of the tx data
		totalRLPLength += txValsLength  // Add the length of the tx values
		err = WriteRLPBytes(buffer, tx.ChainID.Bytes())
		if err != nil {
			return err
		}
		// we have already written the length indicating list length of the tx
		// now we have to write every value.
		err = WriteRLPUint64(buffer, tx.Nonce)
		if err != nil {
			return err
		}

		err = WriteRLPBytes(buffer, tx.GasTipCap.Bytes())
		if err != nil {
			return err
		}

		err = WriteRLPBytes(buffer, tx.GasFeeCap.Bytes())
		if err != nil {
			return err
		}

		err = WriteRLPUint64(buffer, tx.Gas)
		if err != nil {
			return err
		}

		if tx.To != nil {
			err = WriteRLPBytes(buffer, tx.To.Bytes())
		} else {
			err = buffer.WriteByte(ZeroUint64RLPVal)
		}
		if err != nil {
			return err
		}

		if tx.Value == nil {
			err = buffer.WriteByte(ZeroUint64RLPVal)
		} else {
			err = WriteRLPBytes(buffer, tx.Value.Bytes())
			if err != nil {
				return err
			}
		}

		err = WriteRLPBytes(buffer, tx.Data)
		if err != nil {
			return err
		}

		if len(tx.AccessList) > 0 {
			err = tx.EncodeAccessList(buffer)
			if err != nil {
				return err
			}
		} else {
			err = buffer.WriteByte(ZeroListRLPVal)
			if err != nil {
				return err
			}
		}
		// add pointer to let the hasher from which part it should start
		tx.startTx = rlpValsLength
	}

	err := WriteRLPBytes(buffer, tx.V.Bytes())
	if err != nil {
		return err
	}
	err = WriteRLPBytes(buffer, tx.R.Bytes())
	if err != nil {
		return err
	}
	err = WriteRLPBytes(buffer, tx.S.Bytes())
	if save {
		bufferBytes := buffer.Bytes()
		tx.SignedRlpBytes = make([]byte, totalRLPLength)
		copy(tx.SignedRlpBytes, bufferBytes[len(bufferBytes)-totalRLPLength:])
	}
	tx.rlpSignedBytesLength = totalRLPLength
	return err
}

func (tx *CustomTx) EncodeSignedAccessListTx(buffer *bytes.Buffer, save bool) error {

	var totalRLPLength int

	if len(tx.UnsignedRlpBytes) > 0 {
		txValsLength := len(tx.UnsignedRlpBytes) + tx.CalculateRLPLengthSignatureValues()
		rlpValsLength, err := WriteValLength(buffer, CalculateRLPListLength(txValsLength)+1)
		if err != nil {
			return err
		}
		totalRLPLength += rlpValsLength
		buffer.WriteByte(tx.TxType)
		totalRLPLength += 1
		rlpListLength, err := WriteListLength(buffer, txValsLength)
		if err != nil {
			return err
		}
		totalRLPLength += rlpListLength // this add the nBytes used to write the list of the tx data
		totalRLPLength += txValsLength  // Add the length of the tx values
		buffer.Write(tx.UnsignedRlpBytes)
		// add pointer to let the hasher from which part it should start
		tx.startTx = rlpValsLength
	} else {
		// length of the tx values + the signature vals (v,r,s)
		txValsLength := tx.calculateRLPSignedBytesLenAccessListTx()

		// write first the rlp value of the txtype + txvals
		rlpValsLength, err := WriteValLength(buffer, CalculateRLPListLength(txValsLength)+1)
		if err != nil {
			return err
		}
		totalRLPLength += rlpValsLength // this adds the nBytes used to write the size of the tx
		// write the txtype
		buffer.WriteByte(tx.TxType)
		totalRLPLength += 1

		rlpListLength, err := WriteListLength(buffer, txValsLength)
		if err != nil {
			return err
		}
		totalRLPLength += rlpListLength // this add the nBytes used to write the list of the tx data
		totalRLPLength += txValsLength  // Add the length of the tx values
		err = WriteRLPBytes(buffer, tx.ChainID.Bytes())
		if err != nil {
			return err
		}
		// we have already written the length indicating list length of the tx
		// now we have to write every value.
		err = WriteRLPUint64(buffer, tx.Nonce)
		if err != nil {
			return err
		}

		err = WriteRLPBytes(buffer, tx.GasPrice.Bytes())
		if err != nil {
			return err
		}

		err = WriteRLPUint64(buffer, tx.Gas)
		if err != nil {
			return err
		}

		if tx.To != nil {
			err = WriteRLPBytes(buffer, tx.To.Bytes())
		} else {
			err = buffer.WriteByte(ZeroUint64RLPVal)
		}
		if err != nil {
			return err
		}

		if tx.Value == nil {
			err = buffer.WriteByte(ZeroUint64RLPVal)
		} else {
			err = WriteRLPBytes(buffer, tx.Value.Bytes())
			if err != nil {
				return err
			}
		}

		err = WriteRLPBytes(buffer, tx.Data)
		if err != nil {
			return err
		}

		if len(tx.AccessList) > 0 {
			err = tx.EncodeAccessList(buffer)
			if err != nil {
				return err
			}
		} else {
			err = buffer.WriteByte(ZeroListRLPVal)
			if err != nil {
				return err
			}
		}
		// add pointer to let the hasher from which part it should start
		tx.startTx = rlpValsLength
	}

	err := WriteRLPBytes(buffer, tx.V.Bytes())
	if err != nil {
		return err
	}
	err = WriteRLPBytes(buffer, tx.R.Bytes())
	if err != nil {
		return err
	}
	err = WriteRLPBytes(buffer, tx.S.Bytes())
	if save {
		bufferBytes := buffer.Bytes()
		tx.SignedRlpBytes = make([]byte, totalRLPLength)
		copy(tx.SignedRlpBytes, bufferBytes[len(bufferBytes)-totalRLPLength:])
	}
	return err
}

func (tx *CustomTx) EncodeUnsignedLegacyTx(buffer *bytes.Buffer) error {
	if len(tx.SignedRlpBytes) > 0 {
		// we just get the length of the tx without the signed part
		txValsLength := tx.startTxSignature - tx.startTxDataPointer + SIGNER_VALUES_LENGTH
		_, err := WriteListLength(buffer, txValsLength)
		if err != nil {
			return err
		}
		buffer.Write(tx.SignedRlpBytes[tx.startTxDataPointer:tx.startTxSignature])
		buffer.Write(SIGNER_VALUES)
	} else {
		// when doing this it that the tx is a new tx (not signed). If we are doing this most probably we are sending this tx
		// so we store the rlp bytes.
		txValsLength := tx.calculateRLPUnsignedBytesLenLegacyTx()
		_, err := WriteListLength(buffer, txValsLength+SIGNER_VALUES_LENGTH)
		if err != nil {
			return err
		}

		// we have already written the length indicating list length of the tx
		// now we have to write every value.
		err = WriteRLPUint64(buffer, tx.Nonce)
		if err != nil {
			return err
		}

		err = WriteRLPBytes(buffer, tx.GasPrice.Bytes())
		if err != nil {
			return err
		}

		err = WriteRLPUint64(buffer, tx.Gas)
		if err != nil {
			return err
		}

		if tx.To != nil {
			err = WriteRLPBytes(buffer, tx.To.Bytes())
		} else {
			err = buffer.WriteByte(ZeroUint64RLPVal)
		}
		if err != nil {
			return err
		}

		if tx.Value == nil {
			err = buffer.WriteByte(ZeroUint64RLPVal)
		} else {
			err = WriteRLPBytes(buffer, tx.Value.Bytes())
			if err != nil {
				return err
			}
		}

		err = WriteRLPBytes(buffer, tx.Data)
		if err != nil {
			return err
		}

		tx.UnsignedRlpBytes = make([]byte, txValsLength)
		// copy this data to the unsigned rlp bytes
		copy(tx.UnsignedRlpBytes, buffer.Bytes()[len(buffer.Bytes())-txValsLength:])
		// add signer stuff
		_, err = buffer.Write(SIGNER_VALUES)
	}

	return nil
}

func (tx *CustomTx) EncodeUnsignedDynamicFeesTx(buffer *bytes.Buffer) error {

	if len(tx.SignedRlpBytes) > 0 {
		txValsLength := tx.startTxSignature - tx.startTxDataPointer
		err := buffer.WriteByte(tx.TxType)
		if err != nil {
			return err
		}
		_, err = WriteListLength(buffer, txValsLength)
		if err != nil {
			return err
		}
		_, err = buffer.Write(tx.SignedRlpBytes[tx.startTxDataPointer:tx.startTxSignature])
		return err
	} else {

		// notice that we dont att the list value length prefix, since this is only used for getting the unsigned hash
		// also we dont need the prefix when encoding the signed tx
		txValsLength := tx.calculateRLPUnSignedBytesLenDynamicFeesTx()

		buffer.WriteByte(tx.TxType)

		_, err := WriteListLength(buffer, txValsLength)
		if err != nil {
			return err
		}
		err = WriteRLPBytes(buffer, tx.ChainID.Bytes())
		if err != nil {
			return err
		}
		// we have already written the length indicating list length of the tx
		// now we have to write every value.
		err = WriteRLPUint64(buffer, tx.Nonce)
		if err != nil {
			return err
		}

		err = WriteRLPBytes(buffer, tx.GasTipCap.Bytes())
		if err != nil {
			return err
		}

		err = WriteRLPBytes(buffer, tx.GasFeeCap.Bytes())
		if err != nil {
			return err
		}

		err = WriteRLPUint64(buffer, tx.Gas)
		if err != nil {
			return err
		}

		if tx.To != nil {
			err = WriteRLPBytes(buffer, tx.To.Bytes())
		} else {
			err = buffer.WriteByte(ZeroUint64RLPVal)
		}
		if err != nil {
			return err
		}

		if tx.Value == nil {
			err = buffer.WriteByte(ZeroUint64RLPVal)
		} else {
			err = WriteRLPBytes(buffer, tx.Value.Bytes())
			if err != nil {
				return err
			}
		}

		err = WriteRLPBytes(buffer, tx.Data)
		if err != nil {
			return err
		}

		if len(tx.AccessList) > 0 {

			err = tx.EncodeAccessList(buffer)
			if err != nil {
				return err
			}
		} else {
			err = buffer.WriteByte(ZeroListRLPVal)
			if err != nil {
				return err
			}
		}
		bufferBytes := buffer.Bytes()
		tx.UnsignedRlpBytes = make([]byte, txValsLength)
		// copy this data to the unsigned rlp bytes
		copy(tx.UnsignedRlpBytes, bufferBytes[len(bufferBytes)-txValsLength:])
		// add signer stuff
		return err
	}
}

func (tx *CustomTx) EncodeUnsignedAccessListTx(buffer *bytes.Buffer) error {

	if len(tx.SignedRlpBytes) > 0 {
		txValsLength := tx.startTxSignature - tx.startTxDataPointer
		err := buffer.WriteByte(tx.TxType)
		if err != nil {
			return err
		}
		_, err = WriteListLength(buffer, txValsLength)
		if err != nil {
			return err
		}
		_, err = buffer.Write(tx.SignedRlpBytes[tx.startTxDataPointer:tx.startTxSignature])
		return err
	} else {

		// notice that we dont att the list value length prefix, since this is only used for getting the unsigned hash
		// also we dont need the prefix when encoding the signed tx
		txValsLength := tx.calculateRLPUnSignedBytesLenAccessListTx()

		buffer.WriteByte(tx.TxType)

		_, err := WriteListLength(buffer, txValsLength)
		if err != nil {
			return err
		}
		err = WriteRLPBytes(buffer, tx.ChainID.Bytes())
		if err != nil {
			return err
		}
		// we have already written the length indicating list length of the tx
		// now we have to write every value.
		err = WriteRLPUint64(buffer, tx.Nonce)
		if err != nil {
			return err
		}

		err = WriteRLPBytes(buffer, tx.GasPrice.Bytes())
		if err != nil {
			return err
		}

		err = WriteRLPUint64(buffer, tx.Gas)
		if err != nil {
			return err
		}

		if tx.To != nil {
			err = WriteRLPBytes(buffer, tx.To.Bytes())
		} else {
			err = buffer.WriteByte(ZeroUint64RLPVal)
		}
		if err != nil {
			return err
		}

		if tx.Value == nil {
			err = buffer.WriteByte(ZeroUint64RLPVal)
		} else {
			err = WriteRLPBytes(buffer, tx.Value.Bytes())
			if err != nil {
				return err
			}
		}

		err = WriteRLPBytes(buffer, tx.Data)
		if err != nil {
			return err
		}

		if len(tx.AccessList) > 0 {
			err = tx.EncodeAccessList(buffer)
			if err != nil {
				return err
			}
		} else {
			err = buffer.WriteByte(ZeroListRLPVal)
			if err != nil {
				return err
			}
		}
		bufferBytes := buffer.Bytes()
		tx.UnsignedRlpBytes = make([]byte, txValsLength)
		// copy this data to the unsigned rlp bytes
		copy(tx.UnsignedRlpBytes, bufferBytes[len(bufferBytes)-txValsLength:])
		// add signer stuff
		return err
	}
}
