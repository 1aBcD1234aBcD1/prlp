package genTx

import (
	"github.com/1aBcD1234aBcD1/prlp/pool"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

func TestEncodeTxsPacket(t *testing.T) {
	tests := []struct {
		Name      string
		CustomTxs []*CustomTx
		ChainId   *big.Int
		WantError error
	}{
		{
			Name:    "Test multiple txs with chainid 56",
			ChainId: big.NewInt(56),
			CustomTxs: []*CustomTx{
				{
					ChainID:  big.NewInt(56),
					Nonce:    100,
					To:       &common.Address{0x00, 0x01, 0x02, 0x03, 0x04, 0x05},
					Value:    big.NewInt(1000000000000000000),
					Gas:      21000,
					GasPrice: big.NewInt(10000000000),
					TxType:   types.LegacyTxType,
				},
				{
					ChainID:  big.NewInt(56),
					Nonce:    100,
					Value:    big.NewInt(1000000000000000000),
					To:       &common.Address{0x00, 0x01, 0x02, 0x03, 0x04, 0x05},
					Gas:      21000,
					Data:     []byte{0x01, 0x02, 0x03, 0x04, 0x05},
					GasPrice: big.NewInt(10000000000),
					TxType:   types.AccessListTxType,
					AccessList: types.AccessList{
						{
							Address: common.Address{0x00, 0x01, 0x02, 0x03, 0x04, 0x05},
							StorageKeys: []common.Hash{
								common.Hash{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f},
							},
						},
					},
				},
				{
					ChainID:    big.NewInt(56),
					Nonce:      100,
					To:         &common.Address{0x00, 0x01, 0x02, 0x03, 0x04, 0x05},
					Value:      big.NewInt(1000000000000000000),
					Gas:        21000,
					GasPrice:   big.NewInt(10000000000),
					TxType:     types.AccessListTxType,
					AccessList: types.AccessList{},
				},
				{
					ChainID:  big.NewInt(56),
					Nonce:    100,
					Value:    big.NewInt(1000000000000000000),
					Gas:      21000,
					Data:     []byte{0x01, 0x02, 0x03, 0x04, 0x05},
					GasPrice: big.NewInt(10000000000),
					TxType:   types.LegacyTxType,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			Init(tt.ChainId)
			// sign the transaction
			privKey, err := getPrivkeyForTests()
			for _, tx := range tt.CustomTxs {
				err = tx.SignTx(privKey)
				if err != nil {
					t.Fatalf("Failed to sign transaction: %v", err)
				}
			}

			if err != nil {
				t.Fatalf("Failed to get private key: %v", err)
			}

			buf := pool.GetRLPBuffer()
			defer pool.PutRLPBuffer(buf)
			err = EncodeTxsPacket(buf, tt.CustomTxs)
			if err != nil {
				t.Fatalf("Failed to encode txs: %v", err)
			}

			var wantTxs types.Transactions
			err = rlp.DecodeBytes(buf.Bytes(), &wantTxs)
			if err != nil {
				t.Fatalf("Failed to decode txs: %v", err)
			}

			assert.Equal(t, len(wantTxs), len(tt.CustomTxs))
			for i, wantTx := range wantTxs {
				assert.Equal(t, wantTx.Type(), tt.CustomTxs[i].TxType, "tx type not equal at pos %d", i)
				switch wantTx.Type() {
				case types.LegacyTxType:
					compareLegacyTx(t, tt.CustomTxs[i], wantTx)
				case types.AccessListTxType:
					compareAccessListTx(t, tt.CustomTxs[i], wantTx)
				case types.DynamicFeeTxType:
					compareDynamicFeesTx(t, tt.CustomTxs[i], wantTx)
				default:
					t.Fatalf("Unexpected tx type: %d", wantTx.Type())
				}
			}
		})
	}

}
func TestCustomTx_EncodeSingedLegacyTx(t *testing.T) {
	tests := []struct {
		Name     string
		CustomTx *CustomTx

		WantError error
	}{
		{
			Name: "Test simple legacy tx with to and no data",
			CustomTx: &CustomTx{
				ChainID:  big.NewInt(56),
				Nonce:    100,
				To:       &common.Address{0x00, 0x01, 0x02, 0x03, 0x04, 0x05},
				Value:    big.NewInt(1000000000000000000),
				Gas:      21000,
				GasPrice: big.NewInt(10000000000),
				TxType:   types.LegacyTxType,
			},
		},
		{
			Name: "Test legacy tx without to",
			CustomTx: &CustomTx{
				ChainID:  big.NewInt(56),
				Nonce:    100,
				Value:    big.NewInt(1000000000000000000),
				Gas:      21000,
				GasPrice: big.NewInt(10000000000),
				TxType:   types.LegacyTxType,
			},
		},
		{
			Name: "Test empty legacy with to and data",
			CustomTx: &CustomTx{
				ChainID:  big.NewInt(56),
				Nonce:    100,
				Value:    big.NewInt(1000000000000000000),
				To:       &common.Address{0x00, 0x01, 0x02, 0x03, 0x04, 0x05},
				Gas:      21000,
				Data:     []byte{0x01, 0x02, 0x03, 0x04, 0x05},
				GasPrice: big.NewInt(10000000000),
				TxType:   types.LegacyTxType,
			},
		},
		{
			Name: "Test legacy without to but with data",
			CustomTx: &CustomTx{
				ChainID:  big.NewInt(56),
				Nonce:    100,
				Value:    big.NewInt(1000000000000000000),
				Gas:      21000,
				Data:     []byte{0x01, 0x02, 0x03, 0x04, 0x05},
				GasPrice: big.NewInt(10000000000),
				TxType:   types.LegacyTxType,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			// sign the transaction
			privKey, err := getPrivkeyForTests()
			if err != nil {
				t.Fatalf("Failed to get private key: %v", err)
			}
			Init(tt.CustomTx.ChainID)
			err = tt.CustomTx.SignTx(privKey)
			if err != nil {
				t.Fatalf("Failed to sign transaction: %v", err)
			}

			buf := pool.GetRLPBuffer()
			defer pool.PutRLPBuffer(buf)
			err = tt.CustomTx.EncodeSignedRLP(buf, false)

			var want *types.Transaction
			err = rlp.DecodeBytes(buf.Bytes(), &want)
			if err != nil {
				t.Fatalf("Failed to decode signed RLP: %v", err)
			}

			if tt.WantError != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.WantError)
			} else {
				// some previous checks
				assert.Equal(t, want.Type(), tt.CustomTx.TxType)
				compareLegacyTx(t, tt.CustomTx, want)
			}

			wantRLP, err := rlp.EncodeToBytes(want)
			if err != nil {
				t.Fatalf("Failed to encode signed RLP: %v", err)
			}
			assert.Equal(t, wantRLP, buf.Bytes(), "signed RLP not equal")

		})
	}
}
func TestCustomTx_EncodeSignedAccessListTx(t *testing.T) {
	tests := []struct {
		Name      string
		CustomTx  *CustomTx
		WantError error
	}{
		{
			Name: "Test empty access list tx with to and no value",
			CustomTx: &CustomTx{
				ChainID:    big.NewInt(56),
				Nonce:      100,
				To:         &common.Address{0x00, 0x01, 0x02, 0x03, 0x04, 0x05},
				Gas:        21000,
				GasPrice:   big.NewInt(10000000000),
				TxType:     types.AccessListTxType,
				AccessList: types.AccessList{},
			},
		},
		{
			Name: "Test empty access list tx without to and none value",
			CustomTx: &CustomTx{
				ChainID:    big.NewInt(56),
				Nonce:      100000,
				Gas:        21000,
				GasPrice:   big.NewInt(10000000000),
				TxType:     types.AccessListTxType,
				AccessList: types.AccessList{},
			},
		},
		{
			Name: "Test empty access list tx with to",
			CustomTx: &CustomTx{
				ChainID:    big.NewInt(56),
				Nonce:      100,
				To:         &common.Address{0x00, 0x01, 0x02, 0x03, 0x04, 0x05},
				Value:      big.NewInt(1000000000000000000),
				Gas:        21000,
				GasPrice:   big.NewInt(10000000000),
				TxType:     types.AccessListTxType,
				AccessList: types.AccessList{},
			},
		},
		{
			Name: "Test empty access list tx without to",
			CustomTx: &CustomTx{
				ChainID:    big.NewInt(56),
				Nonce:      100,
				Value:      big.NewInt(1000000000000000000),
				Gas:        21000,
				GasPrice:   big.NewInt(10000000000),
				TxType:     types.AccessListTxType,
				AccessList: types.AccessList{},
			},
		},
		{
			Name: "Test empty access list with to and data",
			CustomTx: &CustomTx{
				ChainID:    big.NewInt(56),
				Nonce:      100,
				Value:      big.NewInt(1000000000000000000),
				To:         &common.Address{0x00, 0x01, 0x02, 0x03, 0x04, 0x05},
				Gas:        21000,
				Data:       []byte{0x01, 0x02, 0x03, 0x04, 0x05},
				GasPrice:   big.NewInt(10000000000),
				TxType:     types.AccessListTxType,
				AccessList: types.AccessList{},
			},
		},
		{
			Name: "Test with access list with to and data",
			CustomTx: &CustomTx{
				ChainID:  big.NewInt(56),
				Nonce:    100,
				Value:    big.NewInt(1000000000000000000),
				To:       &common.Address{0x00, 0x01, 0x02, 0x03, 0x04, 0x05},
				Gas:      21000,
				Data:     []byte{0x01, 0x02, 0x03, 0x04, 0x05},
				GasPrice: big.NewInt(10000000000),
				TxType:   types.AccessListTxType,
				AccessList: types.AccessList{
					{
						Address: common.Address{0x00, 0x01, 0x02, 0x03, 0x04, 0x05},
						StorageKeys: []common.Hash{
							common.Hash{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			// sign the transaction
			privKey, err := getPrivkeyForTests()
			if err != nil {
				t.Fatalf("Failed to get private key: %v", err)
			}
			Init(tt.CustomTx.ChainID)
			err = tt.CustomTx.SignTx(privKey)
			if err != nil {
				t.Fatalf("Failed to sign transaction: %v", err)
			}

			buf := pool.GetRLPBuffer()
			defer pool.PutRLPBuffer(buf)
			err = tt.CustomTx.EncodeSignedRLP(buf, false)

			var want *types.Transaction
			err = rlp.DecodeBytes(buf.Bytes(), &want)
			if err != nil {
				t.Fatalf("Failed to decode signed RLP: %v", err)
			}

			if tt.WantError != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.WantError)
			} else {
				// some previous checks
				assert.Equal(t, want.Type(), tt.CustomTx.TxType)
				compareAccessListTx(t, tt.CustomTx, want)
			}

			wantRLP, err := rlp.EncodeToBytes(want)
			if err != nil {
				t.Fatalf("Failed to encode signed RLP: %v", err)
			}
			assert.Equal(t, wantRLP, buf.Bytes(), "signed RLP not equal")
		})
	}
}
func TestCustomTx_EncodeSignedDynamicFeetTx(t *testing.T) {
	tests := []struct {
		Name      string
		CustomTx  *CustomTx
		WantError error
	}{
		{
			Name: "Test empty dynamic fees tx with to and no value",
			CustomTx: &CustomTx{
				ChainID:    big.NewInt(56),
				Nonce:      100,
				To:         &common.Address{0x00, 0x01, 0x02, 0x03, 0x04, 0x05},
				Gas:        21000,
				GasTipCap:  big.NewInt(10000000000),
				GasFeeCap:  big.NewInt(10000000000),
				TxType:     types.DynamicFeeTxType,
				AccessList: types.AccessList{},
			},
		},
		{
			Name: "Test empty dynamic fees tx without to and none value",
			CustomTx: &CustomTx{
				ChainID:    big.NewInt(56),
				Nonce:      100000,
				Gas:        21000,
				GasTipCap:  big.NewInt(10000000000),
				GasFeeCap:  big.NewInt(10000000000),
				TxType:     types.DynamicFeeTxType,
				AccessList: types.AccessList{},
			},
		},
		{
			Name: "Test empty dynamic fees tx with to",
			CustomTx: &CustomTx{
				ChainID:    big.NewInt(56),
				Nonce:      100,
				To:         &common.Address{0x00, 0x01, 0x02, 0x03, 0x04, 0x05},
				Value:      big.NewInt(1000000000000000000),
				Gas:        21000,
				GasTipCap:  big.NewInt(10000000000),
				GasFeeCap:  big.NewInt(10000000000),
				TxType:     types.DynamicFeeTxType,
				AccessList: types.AccessList{},
			},
		},
		{
			Name: "Test dynamic fees tx without to",
			CustomTx: &CustomTx{
				ChainID:    big.NewInt(56),
				Nonce:      100,
				Value:      big.NewInt(1000000000000000000),
				Gas:        21000,
				GasTipCap:  big.NewInt(10000000000),
				GasFeeCap:  big.NewInt(10000000000),
				TxType:     types.DynamicFeeTxType,
				AccessList: types.AccessList{},
			},
		},
		{
			Name: "Test empty dynamic fees with to and data",
			CustomTx: &CustomTx{
				ChainID:   big.NewInt(56),
				Nonce:     100,
				Value:     big.NewInt(1000000000000000000),
				To:        &common.Address{0x00, 0x01, 0x02, 0x03, 0x04, 0x05},
				Gas:       21000,
				Data:      []byte{0x01, 0x02, 0x03, 0x04, 0x05},
				GasTipCap: big.NewInt(10000000000),
				GasFeeCap: big.NewInt(10000000000),
				TxType:    types.DynamicFeeTxType,
			},
		},
		{
			Name: "Test dynamic fees with to and data and access list ",
			CustomTx: &CustomTx{
				ChainID:   big.NewInt(56),
				Nonce:     100,
				Value:     big.NewInt(1000000000000000000),
				To:        &common.Address{0x00, 0x01, 0x02, 0x03, 0x04, 0x05},
				Gas:       21000,
				Data:      []byte{0x01, 0x02, 0x03, 0x04, 0x05},
				GasTipCap: big.NewInt(10000000000),
				GasFeeCap: big.NewInt(10000000000),
				TxType:    types.DynamicFeeTxType,
				AccessList: types.AccessList{
					{
						Address: common.Address{0x00, 0x01, 0x02, 0x03, 0x04, 0x05},
						StorageKeys: []common.Hash{
							common.Hash{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			// sign the transaction
			privKey, err := getPrivkeyForTests()
			if err != nil {
				t.Fatalf("Failed to get private key: %v", err)
			}
			Init(tt.CustomTx.ChainID)
			err = tt.CustomTx.SignTx(privKey)
			if err != nil {
				t.Fatalf("Failed to sign transaction: %v", err)
			}

			buf := pool.GetRLPBuffer()
			defer pool.PutRLPBuffer(buf)
			err = tt.CustomTx.EncodeSignedRLP(buf, false)

			var want *types.Transaction
			err = rlp.DecodeBytes(buf.Bytes(), &want)
			if err != nil {
				t.Fatalf("Failed to decode signed RLP: %v", err)
			}

			if tt.WantError != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.WantError)
			} else {
				// some previous checks
				assert.Equal(t, want.Type(), tt.CustomTx.TxType)
				compareDynamicFeesTx(t, tt.CustomTx, want)
			}

			wantRLP, err := rlp.EncodeToBytes(want)
			if err != nil {
				t.Fatalf("Failed to encode signed RLP: %v", err)
			}
			assert.Equal(t, wantRLP, buf.Bytes(), "signed RLP not equal")
		})
	}
}
