package starknet

import (
	"encoding/json"
	"math/big"
)

const (
	//todo put your account class hash
	AccountClassHash = "0x"

	//todo put your proxy account class hash
	ProxyAccountClassHash = "0x"
)

func NewKeyPair(curve StarkCurve) (priv, pub string, err error) {
	privateKey, err := curve.GetRandomPrivateKey()
	if err != nil {
		return "", "", err
	}
	publicKey, err := curve.PrivateToPublic(privateKey)
	if err != nil {
		return "", "", err
	}

	return BigToHex(privateKey), BigToHex(publicKey), nil
}

func GetPubKey(curve StarkCurve, privKey string) (string, error) {
	privKeyBN := HexToBN(privKey)
	publicKey, err := curve.PrivateToPublic(privKeyBN)
	if err != nil {
		return "", err
	}
	return BigToHex(publicKey), nil
}

func CalculateContractAddressFromHash(starkPub string) (hash *big.Int, err error) {
	salt := HexToBN(starkPub)
	classHash := HexToBN(ProxyAccountClassHash)
	accountClassHash := HexToBN(AccountClassHash)
	deployerAddress := big.NewInt(0)

	calldate := []*big.Int{big.NewInt(2), salt, big.NewInt(0)}

	constructorCallData := []*big.Int{accountClassHash, GetSelectorFromName("initialize")}
	constructorCallData = append(constructorCallData, calldate...)

	constructorCalldataHash, _ := computeHashOnElements(constructorCallData)
	ContractAddressPrefix := HexToBN("0x535441524b4e45545f434f4e54524143545f41444452455353")

	ele := []*big.Int{
		ContractAddressPrefix,
		deployerAddress,
		salt,
		classHash,
		constructorCalldataHash,
	}
	return computeHashOnElements(ele)
}

func GetPubKeyPoint(curve StarkCurve, privKey string) (string, error) {
	x, y, err := curve.PrivateToPoint(HexToBN(privKey))
	if err != nil {
		return "", err
	}
	point, err := json.Marshal(struct {
		X string `json:"publicKey"`
		Y string `json:"publicKeyY"`
	}{BigToHexWithPadding(x), BigToHexWithPadding(y)})

	return string(point), err
}
