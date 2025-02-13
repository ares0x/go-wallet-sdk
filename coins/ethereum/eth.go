package ethereum

import (
	"encoding/hex"
	"fmt"
	"github.com/okx/go-wallet-sdk/crypto"
	"math/big"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/okx/go-wallet-sdk/util"
	"golang.org/x/crypto/sha3"
)

type EthTransaction struct {
	Nonce    *big.Int `json:"nonce"`
	GasPrice *big.Int `json:"gasPrice"`
	GasLimit *big.Int `json:"gas"`
	To       []byte   `json:"to"`
	Value    *big.Int `json:"value"`
	Data     []byte   `json:"data"`
	// Signature values
	V *big.Int `json:"v"`
	R *big.Int `json:"r"`
	S *big.Int `json:"s"`
}

func (tx *EthTransaction) SignTransaction(chainId *big.Int, prvKey *btcec.PrivateKey) string {
	tx.V = chainId
	rawTransaction, _ := rlp.EncodeToBytes([]interface{}{
		tx.Nonce,
		tx.GasPrice,
		tx.GasLimit,
		tx.To,
		tx.Value,
		tx.Data,
		chainId, uint(0), uint(0),
	})
	sig := SignMessage(rawTransaction, prvKey)
	tx.V = big.NewInt(chainId.Int64()*2 + sig.V.Int64() + 8)
	tx.R = sig.R
	tx.S = sig.S
	value, _ := rlp.EncodeToBytes(tx)
	return "0x" + hex.EncodeToString(value)
}

func (tx *EthTransaction) UnSignedTx(chainId *big.Int) string {
	tx.V = chainId
	rawTransaction, _ := rlp.EncodeToBytes([]interface{}{
		tx.Nonce,
		tx.GasPrice,
		tx.GasLimit,
		tx.To,
		tx.Value,
		tx.Data,
		chainId, uint(0), uint(0),
	})
	return hex.EncodeToString(rawTransaction)
}

func (tx *EthTransaction) GetSigningHash(chainId *big.Int) (string, string, error) {
	unSignedTx := tx.UnSignedTx(chainId)
	raw, _ := hex.DecodeString(unSignedTx)
	h := sha3.NewLegacyKeccak256()
	h.Write(raw)
	msgHash := h.Sum(nil)
	return hex.EncodeToString(msgHash), unSignedTx, nil
}

func (tx *EthTransaction) SignedTx(chainId *big.Int, sig *SignatureData) string {
	tx.V = big.NewInt(chainId.Int64()*2 + sig.V.Int64() + 8)
	tx.R = sig.R
	tx.S = sig.S
	value, _ := rlp.EncodeToBytes(tx)
	return "0x" + hex.EncodeToString(value)
}

func SignMessage(message []byte, prvKey *btcec.PrivateKey) *SignatureData {
	hash256 := sha3.NewLegacyKeccak256()
	hash256.Write(message)
	messageHash := hash256.Sum(nil)
	return SignAsRecoverable(messageHash, prvKey)
}

func NewEthTransaction(nonce, gasLimit, gasPrice, value *big.Int, to, data string) *EthTransaction {
	toBytes := util.RemoveZeroHex(to)
	dataBytes := util.RemoveZeroHex(data)
	return &EthTransaction{
		Nonce:    nonce,
		GasPrice: gasPrice,
		GasLimit: gasLimit,
		To:       toBytes,
		Value:    value,
		Data:     dataBytes,
	}

}

func NewTransactionFromRaw(raw string) (*EthTransaction, error) {
	bytes := util.RemoveZeroHex(raw)
	t := new(EthTransaction)
	err := rlp.DecodeBytes(bytes, &t)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func SignAsRecoverable(value []byte, prvKey *btcec.PrivateKey) *SignatureData {
	sig, _ := ecdsa.SignCompact(prvKey, value, false)
	V := sig[0]
	R := sig[1:33]
	S := sig[33:65]
	return &SignatureData{
		V:     new(big.Int).SetBytes([]byte{V}),
		R:     new(big.Int).SetBytes(R),
		S:     new(big.Int).SetBytes(S),
		ByteV: V,
		ByteR: R,
		ByteS: S,
	}
}

type SignatureData struct {
	V *big.Int
	R *big.Int
	S *big.Int

	ByteV byte
	ByteR []byte
	ByteS []byte
}

func NewSignatureData(msgHash []byte, publicKey string, r, s *big.Int) (*SignatureData, error) {
	// Calculate v, r, and s
	pubBytes, err := hex.DecodeString(publicKey)
	if err != nil {
		return nil, err
	}

	pubKey, _ := btcec.ParsePubKey(pubBytes)
	sig, err := crypto.SignCompact(btcec.S256(), r, s, *pubKey, msgHash, false)
	if err != nil {
		return nil, err
	}

	V := sig[0]
	R := sig[1:33]
	S := sig[33:65]
	return &SignatureData{
		V:     new(big.Int).SetBytes([]byte{V}),
		R:     new(big.Int).SetBytes(R),
		S:     new(big.Int).SetBytes(S),
		ByteV: V,
		ByteR: R,
		ByteS: S,
	}, nil
}

func (sd *SignatureData) ToHex() string {
	return hex.EncodeToString(sd.ToBytes())
}

func (sd SignatureData) ToBytes() []byte {
	bytes := []byte{}
	bytes = append(bytes, sd.ByteR...)
	bytes = append(bytes, sd.ByteS...)
	bytes = append(bytes, sd.ByteV)
	return bytes
}

func GetNewAddress(pubKey *btcec.PublicKey) string {
	pubBytes := pubKey.SerializeUncompressed()
	hash := sha3.NewLegacyKeccak256()
	hash.Write(pubBytes[1:])
	addressByte := hash.Sum(nil)
	return "0x" + hex.EncodeToString(addressByte[12:])
}

func GetEthereumMessagePrefix(message string) string {
	return fmt.Sprintf(MessagePrefixTmp, len(message))
}
