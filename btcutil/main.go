package main

import (
	"encoding/hex"
	"flag"
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
)

func main() {
	flagProcess()

	wif, err := btcutil.DecodeWIF(pwd)

	var btcprivateKey *btcec.PrivateKey = wif.PrivKey

	btcprivateKey.Serialize()
	publicKey := btcprivateKey.PubKey()

	publicKeyBytes := publicKey.SerializeUncompressed()
	publicCKeyBytes := publicKey.SerializeCompressed()

	fmt.Println("公钥(去掉前两位)：" + hex.EncodeToString(publicKeyBytes))
	fmt.Println("压缩公钥：" + hex.EncodeToString(publicCKeyBytes))

	// privKeyWif, err := btcutil.NewWIF(btcprivateKey, &chaincfg.RegressionNetParams, false)
	// if err != nil {
	// 	return nil, err
	// }
	// 采用压缩公钥
	pubKey, err := btcutil.NewAddressPubKey(publicCKeyBytes, &chaincfg.MainNetParams)
	if err != nil {
		fmt.Println(err)
	}
	pub2Hash, err := btcutil.NewAddressScriptHash(publicCKeyBytes, &chaincfg.MainNetParams)
	if err != nil {
		fmt.Println(err)
	}
	pubKeyHash, err := btcutil.NewAddressWitnessPubKeyHash(btcutil.Hash160(publicCKeyBytes), &chaincfg.MainNetParams)
	if err != nil {
		fmt.Println(err)
	}
	//pubSKeyHash, err := btcutil.NewAddressWitnessScriptHash(btcutil.Hash160(publicCKeyBytes), &chaincfg.TestNet3Params)
	// if err != nil {
	// 	fmt.Println(err)
	// }

	address := pubKey.EncodeAddress()
	fmt.Println("btc主网地址(其他网络地址需要另外计算)：           " + address)
	address = pub2Hash.EncodeAddress()
	fmt.Println("btc主网隔离见证地址(其他网络地址需要另外计算)：    " + address)
	address = pubKeyHash.EncodeAddress()
	fmt.Println("btc主网隔离见证原生地址(其他网络地址需要另外计算)：" + address)
}

var (
	addr string
	file string
	pwd  string
)

//显示版本信息
func flagProcess() {
	//de97fdbdb823a197603e1f2cb8b1bded3824147e88ebd47367ba82d4b5600d73
	flag.StringVar(&pwd, "p", "cPK16gEe5QCDN2zne5ZDZjMai83HpewZ1RbVcSUXrARKE17AR3ai", "私钥")
	flag.Parse()
}
