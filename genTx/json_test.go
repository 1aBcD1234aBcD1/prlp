package genTx

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"github.com/1aBcD1234aBcD1/prlp/pool"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/v2"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/stretchr/testify/assert"
	"math/big"
	"os"
	"testing"
	"time"
)

func TestCustomTx_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		Name       string
		JsonString string
		WantError  error
	}{
		{
			Name:       "Legacy tx",
			JsonString: `{"blockHash":"0x525d6d7d6d9a88d230dce2db083f6e048f13b510b906cee115af84559d4c5095","blockNumber":"0x3309346","from":"0x5861d617403a1683451a96530ebb2c15f71b9953","gas":"0xacb7c","gasPrice":"0xf7f4900","hash":"0x0d24f68f2612fd49909efc2638b37086925ab14792b37dcb1dab98499eb7d036","input":"0x4de165c30000000000000000000000000000000000000000000000003fd625c9474d470000000000000000000000000000000000000000000000000000000000027a1889000000000000000000000000ff7d6a96ae471bbcd7713af9cb1feeb16cf56b4100000000000000000000000000000000000000000000031af0f66bee4d40f40000000000000000000000000055d398326f99059ff775485246999027b319795500000000000000000000000000000000000000000000003855ff255f88cace3100000000000000000000000000000000000000000000000000000000686f721a00000000000000000000000000000000000000000000000000000000686f721a00000000000000000000000000000000000000000000000000000000000001200000000000000000000000000000000000000000000000000000000000000584e5e8894b0000000000000000000000003d90f66b534dd8482b181e24655a9e8265316be9000000000000000000000000ff7d6a96ae471bbcd7713af9cb1feeb16cf56b4100000000000000000000000000000000000000000000031af0f66bee4d40f40000000000000000000000000055d398326f99059ff775485246999027b319795500000000000000000000000000000000000000000000003855ff255f88cace3100000000000000000000000000000000000000000000000000000000000000c00000000000000000000000000000000000000000000000000000000000000484b80c2f090000000000000000000000000000000000000000000000000000000000000000000000000000000000000000ff7d6a96ae471bbcd7713af9cb1feeb16cf56b4100000000000000000000000055d398326f99059ff775485246999027b319795500000000000000000000000000000000000000000000031af0f66bee4d40f40000000000000000000000000000000000000000000000003855ff255f88cace3200000000000000000000000000000000000000000000000000000000686f73f0000000000000000000000000000000000000000000000000000000000000012000000000000000000000000000000000000000000000000000000000000001600000000000000000000000000000000000000000000000000000000000000460000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000031af0f66bee4d40f400000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000a000000000000000000000000000000000000000000000000000000000000000e000000000000000000000000000000000000000000000000000000000000001200000000000000000000000000000000000000000000000000000000000000160000000000000000000000000ff7d6a96ae471bbcd7713af9cb1feeb16cf56b410000000000000000000000000000000000000000000000000000000000000001000000000000000000000000ca852767b43a395ac1dd54737193eba5e20c78bd0000000000000000000000000000000000000000000000000000000000000001000000000000000000000000ca852767b43a395ac1dd54737193eba5e20c78bd0000000000000000000000000000000000000000000000000000000000000001800000000000000000002710380aadf63d84d3a434073f1d5d95f02fb23d52280000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000c0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000000000000000000000000000000000000060000000000000000000000000ff7d6a96ae471bbcd7713af9cb1feeb16cf56b4100000000000000000000000055d398326f99059ff775485246999027b3197955000000000000000000000000000000000000000000000000000000000000006400000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000","nonce":"0x63cd5","to":"0x6aba0315493b7e6989041c91181337b662fb1b90","transactionIndex":"0x7","value":"0x0","type":"0x0","chainId":"0x38","v":"0x94","r":"0x75300dfedaf352f25ef1ee18c8a47698c49f83cbf58ffd200894ecc33c07c798","s":"0x38b4d3d619b6333c64278259b2ffb768266af588a0fa02c65b97b49f89627b2d"}`,
		},
		{
			Name:       "Dynamic fee tx",
			JsonString: `{"type":"0x2","chainId":"0x2105","nonce":"0x13a599","gas":"0x2dc6c0","maxFeePerGas":"0x7e0c25","maxPriorityFeePerGas":"0x1b8be","to":"0x2887cbd3ef000217db0ecfe13e0a24736fe9ec1e","value":"0x0","accessList":[],"input":"0x543c2e8f53a655e4321d6c0f330b32852b9c55ac2f494c0d3c633165534c36c33ee71e4c7a4d2b1055fd541b52ef3397745712f15fef1eaf5e7a3507638c2ccc79b364cd5db15617792a2d8d3bf2556e87627f4e6e3a699e54367e4d5ea44b297dfd567b9a8564da6eb77bb18db0606a83ef7b8763296baf5b4181e180ba820d5fde747a850979f27e34a1937d6aa2527f9e354fa6c65debca1a91b45a82a651aa767a12a2e3d009a39d693c80a9aac68e97a22b527f4da4579eb3adabae","r":"0x7af79ea2cef4f7fb4e84baf59e42e3dbf1d1075fe8def362bc45dda5dca71e04","s":"0x7f58500f2c9e972804033f53428cf82b1594f90c3fc47b781b20e7401d8faa42","yParity":"0x1","v":"0x1","hash":"0x7b48777e75ff844d0d2fc2a621e55f98bf08d537f936097e5eee0bfd14705be8","blockHash":"0xf68c614d585c50afb08d3939838f0e555a10869242573144315261823ccf7903","blockNumber":"0x1f293ee","transactionIndex":"0xe2","from":"0xa660f3cc9a74a699d14f84b7d0032b6a234b1efc","gasPrice":"0x2daf92"}`,
		},
		{
			Name:       "Dynamic Fee with access list",
			JsonString: `{"blockHash":"0x244ad2b8622e43a4a6573325f2f1b5a49365c99f3cf30a8bec7d2861d696bf21","blockNumber":"0x3e64a55","from":"0x592ce2eeba52a434755c51c22e8b89bc9a1ad1d2","gas":"0xca03a","gasPrice":"0xf1aea54","maxFeePerGas":"0x255736bb","maxPriorityFeePerGas":"0x48f54ef","hash":"0x1437e1a6093981352d4f6f9b6c12b500de4673120831e072b3e9a98ab7766b43","input":"0xccb5f8090000000000000000000000000000000000000000000000000000000000000000000000000000000000000000592ce2eeba52a434755c51c22e8b89bc9a1ad1d2","nonce":"0x3abd9","to":"0x253b2784c75e510dd0ff1da844684a1ac0aa5fcf","transactionIndex":"0x13","value":"0x0","type":"0x2","accessList":[{"address":"0x0200000000000000000000000000000000000005","storageKeys":["0x000000000001af6a974f467006d94388f438014162dd12ec2d1475c48faf09ff","0xe7222d59e4780000026200000000000100000014253b2784c75e510dd0ff1da8","0x44684a1ac0aa5fcf000002400000000000000000000000000000000000000000","0x0000000000000000000000200000000000000000000000000000000000000000","0x00000000000000000003ab5f00000000000000000000000059211b32ac0bfc2c","0xdd67ee1cb89979d2d33e74c80427d4b22a2a78bcddd456742caf91b56badbff9","0x85ee19aef14573e7343fd652000000000000000000000000a8baad3115a133b1","0x01ef935cb2e198fd04f1c6590000000000000000000000000000000000000000","0x000000000000000000030d400000000000000000000000000000000000000000","0x0000000000000000000001000000000000000000000000000000000000000000","0x0000000000000000000001200000000000000000000000000000000000000000","0x0000000000000000000001400000000000000000000000000000000000000000","0x0000000000000000000000000000000000000000000000000000000000000000","0x0000000000000000000000000000000000000000000000000000000000000000","0x0000000000000000000000c00000000000000000000000000000000000000000","0x0000000000000000000000030000000000000000000000000000000000000000","0x0000000000000000000000400000000000000000000000000000000000000000","0x0000000000000000000000600000000000000000000000000200000000000000","0x000000000000000000000001000000000000000000000000a735584961887962","0xa39dcdf628ab21d5e9e8d5170000000000000000000000000000000000000000","0x00000005b81b4b097bb600000000000000000002097fb0113dd8f301e2c6d917","0x486afefcccc7d78095a51c1c0d8614072ce1e117da24bf01e190ca80f8eb5190","0xa5381e90bfa90da43ee1ffdf54822ed0f6d176038bc4b5c1207f20552a95cf1c","0x94021543f8d4190d8036e07c5bf9a7203b5fda7b4c09ff000000000000000000"]}],"chainId":"0xa86a","v":"0x0","r":"0x4b24a3fc0def4e1c0aa93ed6c4d4e83c71e4046bd7becadde9ee9af2a084c4b","s":"0x29b9ee35a3f08cc9b72fd45ce57ebb4a0f17df1c6ce19b8eb792e64eb584e210","yParity":"0x0"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			var customTx CustomTx
			var tx types.Transaction
			err := json.Unmarshal([]byte(tt.JsonString), &customTx)
			assert.NoError(t, err)

			err = json.Unmarshal([]byte(tt.JsonString), &tx)
			if err != nil {
				t.Fatal(err, "error decoding json tx using standard library")
			}

			switch tx.Type() {
			case types.LegacyTxType:
				compareLegacyTx(t, &customTx, &tx)
			case types.DynamicFeeTxType:
				compareDynamicFeesTx(t, &customTx, &tx)
			case types.AccessListTxType:
				compareAccessListTx(t, &customTx, &tx)
			default:
				t.Fatal("unknown tx type")
			}

			buff := pool.GetRLPBuffer()
			defer pool.PutRLPBuffer(buff)

			err = customTx.EncodeSignedRLP(buff, false)
			if err != nil {
				t.Fatal(err, "error encoding signed custom tx")
			}

			b, err := rlp.EncodeToBytes(&tx)
			if err != nil {
				t.Fatal(err, "error encoding signed tx")
			}

			assert.Equal(t, fmt.Sprintf("%x", buff.Bytes()), fmt.Sprintf("%x", b), "encoded signed tx does not match")

		})
	}
}

func TestWS_Sub(t *testing.T) {
	rpcUrl := os.Getenv("BSC_RPC_URL")
	// Needs a websocket subscription since it test stuff in real time.
	rpcClient, err := rpc.Dial(rpcUrl)
	defer rpcClient.Close()

	if err != nil {
		t.Fatalf("failed to connect to RPC server: %v", err)
	}

	// ethclient to ask for chainId and initialize componentes
	ethClient, err := ethclient.Dial(rpcUrl)
	if err != nil {
		t.Fatalf("failed to connect to RPC server: %v", err)
	}

	chainId, err := ethClient.NetworkID(context.Background())
	if err != nil {
		t.Fatalf("failed to get chainId: %v", err)
	}

	Init(chainId)
	ch := make(chan *CustomTx)
	sus, err := rpcClient.EthSubscribe(context.Background(), ch, "newPendingTransactions", true)
	if err != nil {
		t.Fatalf("failed to subscribe: %v", err)
	}
	assert.NotNil(t, sus)
	defer sus.Unsubscribe()

	readUpTo := 1000
	var i int
mainLoop:
	for {
		select {
		case tx := <-ch:
			{
				assert.NotNil(t, tx)
				switch tx.TxType {
				case types.LegacyTxType, types.DynamicFeeTxType, types.AccessListTxType:
					// add this in a function so we can call defer
					func() {
						// store signed hash
						rlpBuffer := pool.GetRLPBuffer()
						defer pool.PutRLPBuffer(rlpBuffer)
						err = tx.EncodeSignedRLP(rlpBuffer, false)
						if err != nil {
							t.Fatalf("failed to encode tx: %v", err)
						}

						var gotTx *types.Transaction

						err = rlp.DecodeBytes(rlpBuffer.Bytes(), &gotTx)
						if err != nil {
							t.Fatalf("failed to decode %s tx %v", tx.Hash(), err)
						}
						// just test some stuff like calculating the hash and from of the tx
						assert.Equal(t, tx.Hash(), gotTx.Hash())

						gotFrom, err := types.LatestSignerForChainID(chainId).Sender(gotTx)
						assert.NoError(t, err)
						wantFrom, err := tx.From()
						assert.NoError(t, err)
						assert.Equal(t, gotFrom, wantFrom)
					}()
					t.Logf("tx \t  %d\t %v", i, tx.Hash())
				default:
					assert.Len(t, tx.signedHash, 32)
					t.Logf("unknown tx type: %v %s", tx.TxType, tx.Hash())
				}
				if i >= readUpTo {
					break mainLoop
				}
				i++
			}

		case err := <-sus.Err():
			t.Fatal(err)
		}

	}

}

func Test_Send_Tx_Through_RPC(t *testing.T) {
	rpcURL := os.Getenv("BSC_RPC_URL")
	ethClient, err := ethclient.Dial(rpcURL)
	if err != nil {
		t.Fatalf("failed to connect to RPC server: %v", err)
	}
	rpcClient, err := rpc.Dial(rpcURL)
	if err != nil {
		t.Fatalf("failed to connect to RPC server: %v", err)
	}
	chainId, err := ethClient.NetworkID(context.Background())
	if err != nil {
		t.Fatalf("failed to get chainId: %v", err)
	}
	Init(chainId)

	privKeyString := os.Getenv("PRIVATE_KEY")

	// Remove the "0x" prefix if present
	cleanPrivKey := privKeyString
	if len(privKeyString) >= 2 && privKeyString[:2] == "0x" {
		cleanPrivKey = privKeyString[2:]
	}

	privateKey, err := crypto.HexToECDSA(cleanPrivKey)
	if err != nil {
		t.Fatalf("failed to decode private key: %v", err)
	}

	// Get the public key and sender address
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		fmt.Println("Failed to cast public key to ECDSA")
		return
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	fmt.Printf("Sender Address: %s\n", fromAddress.Hex())

	tests := []struct {
		Name string
		Tx   *CustomTx
	}{
		{
			Name: "Legacy transaction",
			Tx: &CustomTx{
				ChainID: chainId,
				TxType:  types.LegacyTxType,
				To:      &common.Address{0xde, 0xad, 0x01, 0x02},
				Data:    []byte("This is a test using a custom rlp encoding"),
				Gas:     23000,
			},
		},
		{
			Name: "Legacy transaction with value to same tx",
			Tx: &CustomTx{
				ChainID: chainId,
				TxType:  types.LegacyTxType,
				To: func() *common.Address {
					a := common.HexToAddress("0xBc53b71e03e05cE01e4039feeCCDD262e020C825")
					return &a
				}(),
				Data:  []byte("This is a test using a custom rlp encoding"),
				Gas:   23000,
				Value: big.NewInt(1),
			},
		},
		{
			Name: "Access list transaction with value and access list to same tx",
			Tx: &CustomTx{
				ChainID: chainId,
				TxType:  types.AccessListTxType,
				To: func() *common.Address {
					a := common.HexToAddress("0xBc53b71e03e05cE01e4039feeCCDD262e020C825")
					return &a
				}(),
				Data:  []byte("This is a test using a custom rlp encoding"),
				Gas:   34072,
				Value: big.NewInt(1),
				AccessList: types.AccessList{
					{
						Address: common.HexToAddress("0x0000000000000000000000000000000000000000"),
						StorageKeys: []common.Hash{
							common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000000"),
							common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
						},
					},
					{
						Address: common.HexToAddress("0x1000000000000000000000000000000000000000"),
						StorageKeys: []common.Hash{
							common.HexToHash("0x1000000000000000000000000000000000000000000000000000000000000000"),
							common.HexToHash("0x1000000000000000000000000000000000000000000000000000000000000001"),
						},
					},
				},
			},
		},
		{
			Name: "Access list transaction with value and empty access list to same tx",
			Tx: &CustomTx{
				ChainID: chainId,
				TxType:  types.AccessListTxType,
				To: func() *common.Address {
					a := common.HexToAddress("0xBc53b71e03e05cE01e4039feeCCDD262e020C825")
					return &a
				}(),
				Data:  []byte("This is a test using a custom rlp encoding"),
				Gas:   50000,
				Value: big.NewInt(1),
			},
		},
		{
			Name: "Dynamic fee transaction with no value and random to",
			Tx: &CustomTx{
				ChainID: chainId,
				TxType:  types.DynamicFeeTxType,
				To: func() *common.Address {
					a := common.HexToAddress("0xdead0102")
					return &a
				}(),
				Data: []byte("This is a test using a custom rlp encoding"),
				Gas:  23000,
			},
		},
		{
			Name: "Dynamic fee transaction with value and same to",
			Tx: &CustomTx{
				ChainID: chainId,
				TxType:  types.DynamicFeeTxType,
				To:      &fromAddress,
				Data:    []byte("This is a test using a custom rlp encoding"),
				Gas:     23000,
				Value:   big.NewInt(100000),
			},
		},
		{
			Name: "Dynamic fee transaction with value and same to and accessList",
			Tx: &CustomTx{
				ChainID: chainId,
				TxType:  types.DynamicFeeTxType,
				To:      &fromAddress,
				Data:    []byte("This is a test using a custom rlp encoding"),
				Gas:     34072,
				Value:   big.NewInt(100000),
				AccessList: types.AccessList{
					{
						Address: common.HexToAddress("0x0000000000000000000000000000000000000000"),
						StorageKeys: []common.Hash{
							common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000000"),
							common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
						},
					},
					{
						Address: common.HexToAddress("0x1000000000000000000000000000000000000000"),
						StorageKeys: []common.Hash{
							common.HexToHash("0x1000000000000000000000000000000000000000000000000000000000000000"),
							common.HexToHash("0x1000000000000000000000000000000000000000000000000000000000000001"),
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			tx := tt.Tx
			// get current nonce
			nonce, err := ethClient.NonceAt(context.Background(), fromAddress, nil)
			if err != nil {
				t.Fatal(err)
			}
			tx.Nonce = nonce
			gasPrice, err := ethClient.SuggestGasPrice(context.Background())
			if err != nil {
				t.Fatal(err)
			}

			switch tx.TxType {
			// get gasPrice
			case types.LegacyTxType, types.AccessListTxType:
				tx.GasPrice = gasPrice
			default:
				tx.GasFeeCap = gasPrice
				gasTipCap, err := ethClient.SuggestGasTipCap(context.Background())
				if err != nil {
					t.Fatal(err)
				}
				tx.GasTipCap = gasTipCap
			}

			err = tx.SignTx(privateKey)
			if err != nil {
				t.Fatal(err)
			}

			fromTx, err := tx.From()
			assert.NoError(t, err)
			assert.Equal(t, fromAddress, fromTx)

			hash := tx.Hash()

			rlpBuffer := pool.GetRLPBuffer()
			defer pool.PutRLPBuffer(rlpBuffer)

			err = tx.EncodeSignedRLP(rlpBuffer, false)
			if err != nil {
				t.Fatal(err)
			}

			var previewTx *types.Transaction
			err = rlp.DecodeBytes(rlpBuffer.Bytes(), &previewTx)

			if err != nil {
				t.Fatal(err)
			}

			err = rpcClient.CallContext(context.Background(), &hash, "eth_sendRawTransaction", hexutil.Encode(rlpBuffer.Bytes()[tx.startTx:]))

			assert.NoError(t, err, "%x", rlpBuffer.Bytes()[tx.startTx:])

			_, err = bind.WaitMined(context.Background(), ethClient, hash)

			t.Logf("%s tx hash %d: %s", fromAddress, nonce, hash.Hex())
			time.Sleep(time.Second * 5)

		})
	}

}
