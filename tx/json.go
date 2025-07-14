package tx

import (
	"encoding/json"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
)

type auxCustomTx struct {
	Type  hexutil.Uint64  `json:"type"`
	From  common.Address  `json:"from"`
	Nonce hexutil.Uint64  `json:"nonce"`
	Gas   hexutil.Uint64  `json:"gas"`
	To    *common.Address `json:"to,omitempty"`
	Value *hexutil.Big    `json:"value"`
	Data  hexutil.Bytes   `json:"input"`
	V     *hexutil.Big    `json:"v,omitempty"`
	R     *hexutil.Big    `json:"r,omitempty"`
	S     *hexutil.Big    `json:"s,omitempty"`

	ChainId   *hexutil.Big `json:"chainId,omitempty"`
	GasPrice  *hexutil.Big `json:"gasPrice,omitempty"`
	GasTipCap *hexutil.Big `json:"maxPriorityFeePerGas,omitempty"`
	GasFeeCap *hexutil.Big `json:"maxFeePerGas,omitempty"`

	AccessList types.AccessList             `json:"accessList"`
	AuthList   []types.SetCodeAuthorization `json:"authList"`

	Hash *common.Hash `json:"hash,omitempty"`
}

func (tx *CustomTx) UnmarshalJSON(data []byte) error {
	var aux auxCustomTx
	err := json.Unmarshal(data, &aux)
	if err != nil {
		return err
	}

	tx.TxType = uint8(aux.Type)
	tx.from = aux.From.Bytes()
	tx.Nonce = uint64(aux.Nonce)
	tx.Gas = uint64(aux.Gas)
	tx.To = aux.To
	tx.Value = aux.Value.ToInt()
	tx.Data = aux.Data
	tx.V = aux.V.ToInt()
	tx.R = aux.R.ToInt()
	tx.S = aux.S.ToInt()
	tx.ChainID = aux.ChainId.ToInt()
	if aux.Hash != nil {
		tx.signedHash = aux.Hash.Bytes()
	}

	if tx.TxType == types.LegacyTxType || tx.TxType == types.AccessListTxType {
		tx.GasPrice = aux.GasPrice.ToInt()
	} else {
		tx.GasTipCap = aux.GasTipCap.ToInt()
		tx.GasFeeCap = aux.GasFeeCap.ToInt()
	}

	tx.AccessList = aux.AccessList
	tx.AuthList = aux.AuthList

	// TODO add more data if needed. But atm we dont want to interact with more txs
	return nil
}

func (tx *CustomTx) MarshalJson() ([]byte, error) {
	var aux auxCustomTx
	aux.Type = hexutil.Uint64(tx.TxType)
	aux.To = tx.To
	// only include from if it has been previously calculated
	if len(tx.from) > 0 {
		aux.From = common.BytesToAddress(tx.from)
	}
	aux.Nonce = hexutil.Uint64(tx.Nonce)
	aux.Gas = hexutil.Uint64(tx.Gas)
	aux.Value = (*hexutil.Big)(tx.Value)
	aux.Data = tx.Data
	aux.V = (*hexutil.Big)(tx.V)
	aux.R = (*hexutil.Big)(tx.R)
	aux.S = (*hexutil.Big)(tx.S)
	if tx.ChainID != nil {
		aux.ChainId = (*hexutil.Big)(tx.ChainID)
	}
	if tx.TxType == types.LegacyTxType || tx.TxType == types.AccessListTxType {
		aux.GasPrice = (*hexutil.Big)(tx.GasPrice)
	} else {
		aux.GasTipCap = (*hexutil.Big)(tx.GasTipCap)
		aux.GasFeeCap = (*hexutil.Big)(tx.GasFeeCap)
	}
	aux.AccessList = tx.AccessList
	aux.AuthList = tx.AuthList

	return json.Marshal(aux)

}
