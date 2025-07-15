package genTx

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"fmt"
	reader2 "github.com/1aBcD1234aBcD1/prlp/reader"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/stretchr/testify/assert"
	"log"
	"math/big"
	"testing"
)

func getPrivkeyForTests() (*ecdsa.PrivateKey, error) {
	privKeyString := "0xa0a0a0a0a0a0a0a0a0a0a0a0a0a0a0a0a0a0a0a0a0a0a0a0a0a0a0a0a0a0a0a0" // Replace with your actual private key hex string

	// Remove the "0x" prefix if present
	cleanPrivKey := privKeyString
	if len(privKeyString) >= 2 && privKeyString[:2] == "0x" {
		cleanPrivKey = privKeyString[2:]
	}

	return crypto.HexToECDSA(cleanPrivKey)
}

func TestCustomTx_From_AlreadySigned_Tx_legacy(t *testing.T) {
	client, err := ethclient.Dial("http://localhost:8080/bsc/rpc")
	if err != nil {
		t.Fatalf("Failed to connect to Ethereum node: %v", err)
	}
	defer client.Close()

	hashes := []common.Hash{
		common.HexToHash("0x39b8cedb315237b0fb80f942bb70d9b197e0620279910489cdad82fa388de3ed"), //nonce 1
	}

	for _, h := range hashes {
		fmt.Println("testing hash ", h)
		tx, _, err := client.TransactionByHash(context.Background(), h)
		if err != nil {
			panic(err)
		}

		Init(tx.ChainId())

		var txBuf bytes.Buffer
		err = tx.EncodeRLP(&txBuf)
		if err != nil {
			t.Fatalf("Failed to RLP encode transaction: %v", err)
		}
		fmt.Println(fmt.Sprintf("%x", txBuf.Bytes()))

		from, err := types.Sender(types.LatestSignerForChainID(tx.ChainId()), tx)
		assert.NoError(t, err)
		customTx, err := DecodeLegacyTx(reader2.NewReader(txBuf.Bytes()))
		assert.NoError(t, err)
		customTxFrom, err := customTx.From()
		assert.NoError(t, err)
		assert.Equal(t, from, customTxFrom)
	}
}

func TestCustomTx_From_AlreadySigned_Tx_legacy_None_RLPSIgned(t *testing.T) {
	client, err := ethclient.Dial("http://localhost:8080/bsc/rpc")
	if err != nil {
		t.Fatalf("Failed to connect to Ethereum node: %v", err)
	}
	defer client.Close()

	hashes := []common.Hash{
		common.HexToHash("0x39b8cedb315237b0fb80f942bb70d9b197e0620279910489cdad82fa388de3ed"), //nonce 1
	}

	for _, h := range hashes {
		fmt.Println("testing hash ", h)
		tx, _, err := client.TransactionByHash(context.Background(), h)
		if err != nil {
			panic(err)
		}

		Init(tx.ChainId())

		var txBuf bytes.Buffer
		err = tx.EncodeRLP(&txBuf)
		if err != nil {
			t.Fatalf("Failed to RLP encode transaction: %v", err)
		}
		fmt.Println(fmt.Sprintf("%x", txBuf.Bytes()))

		from, err := types.Sender(types.LatestSignerForChainID(tx.ChainId()), tx)
		assert.NoError(t, err)
		customTx, err := DecodeLegacyTx(reader2.NewReader(txBuf.Bytes()))
		assert.NoError(t, err)
		customTx.SignedRlpBytes = []byte{}
		customTxFrom, err := customTx.From()
		assert.NoError(t, err)
		assert.Equal(t, from, customTxFrom)
	}
}

func TestCustomTx_From_AlreadySigned_Tx_Dynamic_Fees(t *testing.T) {
	client, err := ethclient.Dial("http://localhost:8080/bsc/rpc")
	if err != nil {
		t.Fatalf("Failed to connect to Ethereum node: %v", err)
	}
	defer client.Close()

	hashes := []common.Hash{
		common.HexToHash("0x7669c12696183cfcccc8d97c12aac89cbcd08fe5516c3522fc0058c8a33c153f"), //nonce 1
	}

	for _, h := range hashes {
		fmt.Println("testing hash ", h)
		tx, _, err := client.TransactionByHash(context.Background(), h)
		if err != nil {
			panic(err)
		}

		Init(tx.ChainId())

		var txBuf bytes.Buffer
		err = tx.EncodeRLP(&txBuf)
		if err != nil {
			t.Fatalf("Failed to RLP encode transaction: %v", err)
		}
		fmt.Println(fmt.Sprintf("%x", txBuf.Bytes()))
		rlpBytes := txBuf.Bytes()
		reader := reader2.NewReader(rlpBytes)

		// returns a slice of bytes with the tx type and the list of values of the tx
		startPoint := reader.Pos()
		// we already assume that this is a AccessList tx so we just read how many bytes it has
		_, err = reader.ReadValueSize()
		if err != nil {
			panic(err)
		}
		startPoint = reader.Pos() - startPoint
		if v, err := reader.ReadByte(); v != 0x2 || err != nil {
			if err != nil {
				panic(err)
			} else {
				panic("not a access list tx")
			}
		}

		from, err := types.Sender(types.LatestSignerForChainID(tx.ChainId()), tx)
		assert.NoError(t, err)

		customTx, err := DecodeDynamicFeeTx(reader, rlpBytes, startPoint)

		assert.NoError(t, err)
		customTxFrom, err := customTx.From()
		assert.NoError(t, err)
		assert.Equal(t, from, customTxFrom)
	}
}

func TestCustomTx_From_AlreadySigned_Tx_Dynamic_None_RLP(t *testing.T) {
	client, err := ethclient.Dial("http://localhost:8080/bsc/rpc")
	if err != nil {
		t.Fatalf("Failed to connect to Ethereum node: %v", err)
	}
	defer client.Close()

	hashes := []common.Hash{
		common.HexToHash("0x7669c12696183cfcccc8d97c12aac89cbcd08fe5516c3522fc0058c8a33c153f"), //nonce 1
	}

	for _, h := range hashes {
		fmt.Println("testing hash ", h)
		tx, _, err := client.TransactionByHash(context.Background(), h)
		if err != nil {
			panic(err)
		}

		Init(tx.ChainId())

		var txBuf bytes.Buffer
		err = tx.EncodeRLP(&txBuf)
		if err != nil {
			t.Fatalf("Failed to RLP encode transaction: %v", err)
		}
		fmt.Println(fmt.Sprintf("%x", txBuf.Bytes()))
		rlpBytes := txBuf.Bytes()
		reader := reader2.NewReader(rlpBytes)

		// returns a slice of bytes with the tx type and the list of values of the tx
		startPoint := reader.Pos()
		// we already assume that this is a AccessList tx so we just read how many bytes it has
		_, err = reader.ReadValueSize()
		if err != nil {
			panic(err)
		}
		startPoint = reader.Pos() - startPoint
		if v, err := reader.ReadByte(); v != 0x2 || err != nil {
			if err != nil {
				panic(err)
			} else {
				panic("not a access list tx")
			}
		}

		from, err := types.Sender(types.LatestSignerForChainID(tx.ChainId()), tx)
		assert.NoError(t, err)

		customTx, err := DecodeDynamicFeeTx(reader, rlpBytes, startPoint)
		customTx.SignedRlpBytes = []byte{}
		assert.NoError(t, err)
		customTxFrom, err := customTx.From()
		fmt.Println(fmt.Sprintf("%x", customTx.UnsignedRlpBytes))
		assert.NoError(t, err)
		assert.Equal(t, from, customTxFrom)
	}
}

func TestCustomTx_Legacy_Sign_Tx(t *testing.T) {

	privKeyString := "0xa0a0a0a0a0a0a0a0a0a0a0a0a0a0a0a0a0a0a0a0a0a0a0a0a0a0a0a0a0a0a0a0" // Replace with your actual private key hex string

	// Remove the "0x" prefix if present
	cleanPrivKey := privKeyString
	if len(privKeyString) >= 2 && privKeyString[:2] == "0x" {
		cleanPrivKey = privKeyString[2:]
	}

	privateKey, err := crypto.HexToECDSA(cleanPrivKey)
	if err != nil {
		log.Fatalf("Failed to parse private key: %v", err)
	}

	// Print the private key (for demonstration; in production, secure this!)
	privateKeyBytes := crypto.FromECDSA(privateKey)
	fmt.Printf("Generated Private Key: %x\n", privateKeyBytes)

	// Get the public key and sender address
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		fmt.Println("Failed to cast public key to ECDSA")
		return
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	fmt.Printf("Sender Address: %s\n", fromAddress.Hex())

	chainId := new(big.Int).SetUint64(56)
	Init(chainId)

	tx := types.NewTx(&types.LegacyTx{
		Nonce:    1,
		GasPrice: big.NewInt(1000000000),
		Gas:      1000,
		Data:     []byte{0xaa, 0xbb, 0xcc, 0xdd, 0xee},
		To:       &common.Address{0x00, 0x01, 0x02, 0x03},
	})

	customTx := new(CustomTx)
	err = customTx.FromTx(tx)
	if err != nil {
		panic(err)
	}

	signer := types.LatestSignerForChainID(chainId)
	signedTx, err := types.SignTx(tx, signer, privateKey)
	if err != nil {
		fmt.Printf("Failed to sign transaction: %v\n", err)
		return
	}

	err = customTx.SignTx(privateKey)
	assert.NoError(t, err)

	v, r, s := signedTx.RawSignatureValues()

	assert.Equal(t, customTx.V, v)
	assert.Equal(t, customTx.R, r)
	assert.Equal(t, customTx.S, s)

	// check hash and rlp
	hash := signedTx.Hash()
	gotHash := customTx.Hash()
	assert.Equal(t, hash, gotHash)

	rlpBytes, err := rlp.EncodeToBytes(signedTx)
	if err != nil {
		panic(err)
	}

	buff := bytes.NewBuffer([]byte{})
	err = customTx.EncodeSignedRLP(buff, true)
	assert.NoError(t, err)
	assert.Equal(t, fmt.Sprintf("%x", rlpBytes), fmt.Sprintf("%x", buff.Bytes()))

	bufPackets := bytes.NewBuffer([]byte{})
	err = EncodeTxsPacket(bufPackets, []*CustomTx{customTx})

	rlpPackets, err := rlp.EncodeToBytes(types.Transactions{signedTx})
	if err != nil {
		panic(err)
	}

	assert.Equal(t, fmt.Sprintf("%x", rlpPackets), fmt.Sprintf("%x", bufPackets.Bytes()))
}

func TestCustomTx_Dynamic_Fees_Sign_Tx(t *testing.T) {
	privKeyString := "0xa0a0a0a0a0a0a0a0a0a0a0a0a0a0a0a0a0a0a0a0a0a0a0a0a0a0a0a0a0a0a0a0" // Replace with your actual private key hex string

	// Remove the "0x" prefix if present
	cleanPrivKey := privKeyString
	if len(privKeyString) >= 2 && privKeyString[:2] == "0x" {
		cleanPrivKey = privKeyString[2:]
	}

	privateKey, err := crypto.HexToECDSA(cleanPrivKey)
	if err != nil {
		log.Fatalf("Failed to parse private key: %v", err)
	}

	// Print the private key (for demonstration; in production, secure this!)
	privateKeyBytes := crypto.FromECDSA(privateKey)
	fmt.Printf("Generated Private Key: %x\n", privateKeyBytes)

	// Get the public key and sender address
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		fmt.Println("Failed to cast public key to ECDSA")
		return
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	fmt.Printf("Sender Address: %s\n", fromAddress.Hex())

	chainId := new(big.Int).SetUint64(56)
	Init(chainId)

	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   chainId,
		Nonce:     1,
		GasTipCap: big.NewInt(1000000000),
		GasFeeCap: big.NewInt(1000000000),
		Gas:       1000,
		Data:      []byte{0xaa, 0xbb, 0xcc, 0xdd, 0xee},
		To:        &common.Address{0x00, 0x01, 0x02, 0x03},
	})

	customTx := new(CustomTx)
	err = customTx.FromTx(tx)
	if err != nil {
		panic(err)
	}

	signer := types.LatestSignerForChainID(chainId)
	signedTx, err := types.SignTx(tx, signer, privateKey)
	if err != nil {
		fmt.Printf("Failed to sign transaction: %v\n", err)
		return
	}

	err = customTx.SignTx(privateKey)
	assert.NoError(t, err)

	v, r, s := signedTx.RawSignatureValues()

	assert.Equal(t, customTx.V, v)
	assert.Equal(t, customTx.R, r)
	assert.Equal(t, customTx.S, s)

	// check hash and rlp
	hash := signedTx.Hash()
	gotHash := customTx.Hash()
	assert.Equal(t, hash, gotHash)

	rlpBytes, err := rlp.EncodeToBytes(signedTx)
	if err != nil {
		panic(err)
	}

	buff := bytes.NewBuffer([]byte{})
	err = customTx.EncodeSignedRLP(buff, true)
	assert.NoError(t, err)
	assert.Equal(t, fmt.Sprintf("%x", rlpBytes), fmt.Sprintf("%x", buff.Bytes()))

	bufPackets := bytes.NewBuffer([]byte{})
	err = EncodeTxsPacket(bufPackets, []*CustomTx{customTx})

	rlpPackets, err := rlp.EncodeToBytes(types.Transactions{signedTx})
	if err != nil {
		panic(err)
	}

	assert.Equal(t, fmt.Sprintf("%x", rlpPackets), fmt.Sprintf("%x", bufPackets.Bytes()))
}

func compareLegacyTx(t *testing.T, got *CustomTx, want *types.Transaction) {
	assert.Equal(t, want.ChainId().String(), got.ChainID.String(), "chain id not equal")
	assert.Equal(t, want.Nonce(), got.Nonce, "nonce not equal")
	assert.Equal(t, want.Gas(), got.Gas, "gas not equal")
	assert.Equal(t, want.GasPrice().String(), got.GasPrice.String(), "gas price not equal")
	if want.To() != nil {
		assert.Equal(t, want.To().String(), got.To.String(), "to not equal")
	} else {
		assert.Nil(t, got.To)
	}

	if want.Value().String() != "0" {
		assert.Equal(t, want.Value().String(), got.Value.String(), "value not equal")
	} else {
		if got.Value == nil {
			assert.Nil(t, got.Value)
		} else {
			assert.Equal(t, want.Value().String(), got.Value.String(), "value not equal")
		}
	}

	if len(want.Data()) > 0 {
		assert.Equal(t, want.Data(), got.Data, "data not equal")
	} else {
		assert.Len(t, got.Data, 0)
	}
}
func compareAccessListTx(t *testing.T, got *CustomTx, want *types.Transaction) {
	assert.Equal(t, want.ChainId().String(), got.ChainID.String(), "chain id not equal")
	assert.Equal(t, want.Nonce(), got.Nonce, "nonce not equal")
	assert.Equal(t, want.Gas(), got.Gas, "gas not equal")
	assert.Equal(t, want.GasPrice().String(), got.GasPrice.String(), "gas price not equal")
	if want.Value().String() != "0" {
		assert.Equal(t, want.Value().String(), got.Value.String(), "value not equal")
	} else {
		if got.Value == nil {
			assert.Nil(t, got.Value)
		} else {
			assert.Equal(t, want.Value().String(), got.Value.String(), "value not equal")
		}
	}
	if want.To() != nil {
		assert.Equal(t, want.To().String(), got.To.String(), "to not equal")
	} else {
		assert.Nil(t, got.To)
	}

	if len(want.Data()) > 0 {
		assert.Equal(t, want.Data(), got.Data, "data not equal")
	} else {
		assert.Len(t, got.Data, 0)
	}
	assert.Equal(t, want.AccessList(), got.AccessList, "access list not equal")
}
func compareDynamicFeesTx(t *testing.T, got *CustomTx, want *types.Transaction) {
	assert.Equal(t, want.ChainId().String(), got.ChainID.String(), "chain id not equal")
	assert.Equal(t, want.Nonce(), got.Nonce, "nonce not equal")
	assert.Equal(t, want.Gas(), got.Gas, "gas not equal")
	assert.Equal(t, want.GasFeeCap().String(), got.GasFeeCap.String(), "gas fee cap not equal")
	assert.Equal(t, want.GasTipCap().String(), got.GasTipCap.String(), "gas tip cap not equal")
	if want.Value().String() != "0" {
		assert.Equal(t, want.Value().String(), got.Value.String(), "value not equal")
	} else {
		if got.Value == nil {
			assert.Nil(t, got.Value)
		} else {
			assert.Equal(t, want.Value().String(), got.Value.String(), "value not equal")
		}
	}
	if want.To() != nil {
		assert.Equal(t, want.To().String(), got.To.String(), "to not equal")
	} else {
		assert.Nil(t, got.To)
	}

	if len(want.Data()) > 0 {
		assert.Equal(t, want.Data(), got.Data, "data not equal")
	} else {
		assert.Len(t, got.Data, 0)
	}
	if want.AccessList() != nil {
		if len(want.AccessList()) == 0 {
			if got.AccessList != nil {
				assert.Len(t, got.AccessList, 0)
			} else {
				assert.Nil(t, got.AccessList)
			}
		} else {
			assert.Equal(t, want.AccessList(), got.AccessList, "access list not equal")
		}
	} else {
		if got.AccessList != nil {
			assert.Len(t, got.AccessList, 0)
		} else {
			assert.Nil(t, got.AccessList)
		}
	}
}
