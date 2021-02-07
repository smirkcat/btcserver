package coin

import (
	"fmt"
	"testing"

	"github.com/shopspring/decimal"
)

// rpcuser=user4181334432
// rpcpassword=pass493b8ca5c1d1e752b40e59aabd3e12db558b99a91107b4c3f3331050c6fb92bc0c
// rpcport=63220

// 172.104.113.147

// var CoinName = "btc"
// var Host = "192.168.1.11" //"192.168.1.4"
// var Port = "18332"        //"8333"
// var Password = "8e7OmEaXPalNB7kvoFnsuGhjfu1YI5ajVa4vKoPD"
// var Account = "user"
// var WalletPwd = "789456123"

// var CoinName = "ltc"
// var Host = "192.168.1.11" //"192.168.1.4"
// var Port = "9333"         //"8333"
// var Password = "123456789"
// var Account = "wawawa"
// var mainaddr = "QTEKcDAdG4sWbB6PSkgtUuEbbbFjnwRzHK" // QiRLSzxW9P5HbFcP1maTjZNQpsRLpP8n58

var CoinName = "bch"
var Host = "192.168.1.32" //"192.168.1.4"
var Port = "18332"        //"8333"
var Password = "123456789"
var WalletPwd = "123qweasdzxc"
var Account = "wawawa"
var mainaddr = "bchtest:qzfmw68rkvxzfzqqe7jxnsx4jcqf8wp7u5tsf0797c" // bchtest:qpp7d2axp9vyyla7dna2kuzfus6hula8658dleer8e

var co = NewWalletCoin(CoinName, Host, Port, Account, Password, WalletPwd)

// NewWalletCoin 初始化币结构体 3.47103997004466e+07
func TestNewWalletCoin(t *testing.T) {
	// 2Msjm7evZgePaEpxNrNxUizhMn64V3bdpvi
	re, err := co.Listunspent([]string{"bchtest:qzfmw68rkvxzfzqqe7jxnsx4jcqf8wp7u5tsf0797c"})
	if err != nil {
		fmt.Println(err)
		return
	}
	// 0.03185881
	fmt.Println(re)
}

func TestWalletpassphrase(t *testing.T) {
	err := co.Walletpassphrase(60)
	if err != nil {
		fmt.Println(err)
		return
	}
}

// CreateWalletAddress 创建钱包的支付地址
func TestCreateWalletAddress(t *testing.T) {
	re, err := co.Getnewaddress("golang")
	if err != nil {
		fmt.Println(err)
		return
	}
	// n3V1FowaAxWV68LJp5xJhR1qEbyfsD29kC 先里转0.01
	// mwAzHKfEEvtXHR7ZttDjLnBzV5yjP9b7nZ 先里转0.01
	fmt.Println(re)
}

// input*148+34*out+10
// var btcfee = decimal.NewFromFloat(0.00164)
var btcdanwei = decimal.NewFromInt(10000_0000)

// https://bitcoinfees.earn.com/  获取当前手续费单位
//  再 Which fee should I use? 下面第一行 当前默认 设为 22 satoshis/byte
//  单位 satoshis/byte

// https://bitcoinfees.earn.com/api 获取交易手续费
// https://bitcoinfees.earn.com/api/v1/fees/recommended

// https://coinfaucet.eu/en/btc-testnet/ 获取btc测试币
// https://testnet-faucet.mempool.co/
// https://developer.bitcoin.com/faucets/bch/
// https://testnet.help/en/bchfaucet/testnet#log
// https://coinfaucet.eu/en/bch-testnet/
// https://kuttler.eu/en/bitcoin/ltc/faucet/
// https://tltc.bitaps.com/
// http://faucet.thonguyen.net/ltc
// 获取测试1 token 发送测试币到 moneyqMan7uh8FqdCA2BV5yZ8qVrc9ikLP

func TestSend(t *testing.T) {
	mainaddr := []string{"bchtest:qzfmw68rkvxzfzqqe7jxnsx4jcqf8wp7u5tsf0797c"}
	utxo, err := co.Listunspent(mainaddr)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(utxo)
	var mapamount = make(map[string]decimal.Decimal, 0)
	mapamount["bchtest:qpp7d2axp9vyyla7dna2kuzfus6hula8658dleer8e"] = decimal.NewFromFloat(0.001)
	//mapamount["mwAzHKfEEvtXHR7ZttDjLnBzV5yjP9b7nZ"] = decimal.NewFromFloat(0.01)
	total := decimal.NewFromFloat(0.001)
	vouttotal := decimal.Zero
	sizeout := 34*len(mapamount) + 34 + 10 // 手续费找零
	lenutxo := len(utxo)
	fee, err := co.GetFee()
	if err != nil {
		fmt.Println(err)
	}
	maxoutfee := decimal.New(int64(sizeout*fee.HalfHourFee), 0).Div(btcdanwei)
	perutxofee := decimal.New(int64(148*fee.HalfHourFee), 0).Div(btcdanwei)
	//peroutfee := decimal.New(int64(34*fee.HalfHourFee), 0).Div(btcdanwei)
	utoxfee := decimal.Zero
	shouldtotal := total.Add(maxoutfee)
	utxosMap := make([]UtxosMap, 0)
	for i := 0; i < lenutxo; i++ {
		v := utxo[i]
		tmp := v.Amount
		utoxfee = utoxfee.Add(perutxofee)
		vouttotal = vouttotal.Add(tmp)
		tmputxo := UtxosMap{
			Txid: v.Txid,
			Vout: v.Vout,
		}
		utxosMap = append(utxosMap, tmputxo)
		if vouttotal.Sub(utoxfee).GreaterThanOrEqual(shouldtotal) {
			break
		}
	}
	if vouttotal.Sub(utoxfee).LessThan(shouldtotal) {
		fmt.Printf("可用余额 %s 不足 应该有 %s 请检查钱包主地址余额", vouttotal.String(), shouldtotal.String())
		return
	}

	btcfee := utoxfee.Add(maxoutfee)
	mapamount["bchtest:qzfmw68rkvxzfzqqe7jxnsx4jcqf8wp7u5tsf0797c"] = vouttotal.Sub(btcfee).Sub(total)

	fmt.Println("计算结果")
	fmt.Println(vouttotal)
	fmt.Println(maxoutfee)
	fmt.Println(utoxfee)
	fmt.Println(btcfee)
	fmt.Println(utxosMap)
	fmt.Println(mapamount)
	txid, err := co.Sends(utxosMap, mapamount)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(txid)
}

// TestGetOmniGetbalance 获取地址代币余额
func TestGetOmniGetbalance(t *testing.T) {
	// n3V1FowaAxWV68LJp5xJhR1qEbyfsD29kC
	re, err := co.OmniGetbalance("mwAzHKfEEvtXHR7ZttDjLnBzV5yjP9b7nZ", 1)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(re)
}

// OmniListtransactions 获取钱包交易记录
func TestOmniListtransactions(t *testing.T) {
	re, err := co.OmniListtransactions("*", 100, 0, 0, 99999999)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(re)
	fmt.Println(re[0].BlockIndex)
}

// SendWalletTransfer 从交易平台提币转帐 R9JRJD7YwdmdnJvw1F97o4qb6hC9iy1GFZ
func TestGetAllbalance(t *testing.T) {
	re, err := co.OmniGetwalletbalances()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(re)
}

func TestCollect(t *testing.T) {
	mainAddr := "2Msjm7evZgePaEpxNrNxUizhMn64V3bdpvi"
	// 这个地方检测余额 暂时不并行处理，一般本地连接处理很快的 后期可通过交易记录处理
	collectAddr := []string{"2NEaeffbnv5zjV4jseVAP5UwCmE67gLBTSd"}
	utxos, err := co.Listunspent(collectAddr)
	if err != nil {
		fmt.Println(err)
		return
	}
	lensutxo := len(utxos)
	if lensutxo < 1 {
		return
	}
	vouttotal := decimal.Zero
	var utxosMap []UtxosMap
	for i := 0; i < lensutxo; i++ {
		vouttotal = vouttotal.Add(utxos[i].Amount)
		tmp := UtxosMap{
			Txid: utxos[i].Txid, // 后面检验txid来源地址是否可靠
			Vout: utxos[i].Vout,
		}
		utxosMap = append(utxosMap, tmp)
	}
	// input*148+34*out+10
	fee, err := co.GetFee()
	if err != nil {
		fmt.Println(err)
	}

	sizein := 148 * lensutxo
	sizeout := 34 + 10 // 手续费找零
	sizeall := sizein + sizeout
	shuoldfee := decimal.New(int64(sizeall*fee.HalfHourFee), 0).Div(btcdanwei)

	var mapbtc = make(map[string]decimal.Decimal, 0) // 需要转手续费的地址
	mapbtc[mainAddr] = vouttotal.Sub(shuoldfee)

	txid, err := co.Sends(utxosMap, mapbtc)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("此次%v归集转账返回txid:%s,等待确认中", collectAddr, txid)
}
