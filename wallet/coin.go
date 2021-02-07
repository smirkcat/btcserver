package coin

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/shopspring/decimal"
)

// Coin 新的rpc接口 采用golang编写 后期替换全部使用golang编写 参考这个接口
type Coin struct {
	Name      string
	WalletPwd string
	rpc       *JsonRpc
}

// NewWalletCoin 初始化币结构体
func NewWalletCoin(coinname, host, port, account, password, walletPwd string) *Coin {
	return &Coin{coinname, walletPwd, NewJsonRpc(host, port, account, password)}
}

// {"fastestFee":16,"halfHourFee":16,"hourFee":4}
//  https://bitcoinfees.earn.com/api/v1/fees/recommended
// 总手续费 size*HalfHourFee

// Fee 手续费每字节单位获取
type Fee struct {
	FastestFee  int `json:"fastestFee"`  // 最快手续费
	HalfHourFee int `json:"halfHourFee"` // 半个小时确认手续费
	HourFee     int `json:"hourFee"`     // 一个小时手续费
}

// GetFee 获取手续费
func (b *Coin) GetFee() (Fee, error) {
	var resp = Fee{20, 20, 20}
	if b.Name == "btc" || b.Name == "usdt" {
		res, err := http.DefaultClient.Get("https://bitcoinfees.earn.com/api/v1/fees/recommended")
		if err != nil {
			return resp, err
		}
		defer res.Body.Close()
		if res.StatusCode != 200 {
			return resp, err
		}
		tmp, _ := ioutil.ReadAll(res.Body)
		err = json.Unmarshal(tmp, &resp)
		if err != nil {
			return Fee{20, 20, 20}, err
		}
	} else if b.Name == "bch" {
		//  https://txstreet.com/
		// https://fork.lol/tx/fee
		// https://bitinfocharts.com/bitcoin%20cash/
		resp = Fee{2, 1, 1}
	} else if b.Name == "ltc" {
		// https://bitinfocharts.com/litecoin/
		// https://medium.com/the-litecoin-school-of-crypto/understanding-minimum-fees-525877ac8f5f
		// 最小 0.0001 LTC/kB
		// https://github.com/litecoin-project/litecoin/blob/master/src/wallet/wallet.h#L51 DEFAULT_TRANSACTION_MINFEE
		resp = Fee{110, 110, 110} // 0.0011
	}
	return resp, nil
}

//  listunspent
// [{
// 	"txid": "65b8b7f155515e17e09364cac0b2d40ce49f6fdf36b3d648ac1245b347ee0559",
// 	"vout": 0,
// 	"address": "mmxL38k284XZvXKnwNTFNUTVRxHiZWEQ16",
// 	"scriptPubKey": "76a914469d4dbec5b7b0dec9b6ab8a6ff6192f1674c4a388ac",
// 	"amount": 0.04999775,
// 	"confirmations": 23127,
// 	"spendable": true,
// 	"solvable": true
// }]

// Addressutxos ..
type Addressutxos struct {
	Address       string          `json:"address"`
	Txid          string          `json:"txid"`
	Vout          int             `json:"vout"`
	Script        string          `json:"scriptPubKey"`
	Amount        decimal.Decimal `json:"amount"`
	Confirmations int64           `json:"confirmations"`
}

// Prevtxs .
type Prevtxs struct {
	Txid   string          `json:"txid"`
	Vout   int             `json:"vout"`
	Script string          `json:"scriptPubKey"`
	Value  decimal.Decimal `json:"value"`
}

// Listunspent 获取utox
// https://chainquery.com/bitcoin-cli/listunspent
// https://bitcoin.org/en/developer-reference#listunspent
func (b *Coin) Listunspent(addrs []string) ([]Addressutxos, error) {
	var resp []Addressutxos
	err := b.rpc.CallMethod(&resp, "listunspent", 6, 99999999, addrs)
	return resp, err
}

// Listunspent05 获取utox 未被确认的
func (b *Coin) Listunspent05(addrs []string) ([]Addressutxos, error) {
	var resp []Addressutxos
	err := b.rpc.CallMethod(&resp, "listunspent", 0, 5, addrs)
	return resp, err
}

// CalcBalance 计算输入可用总和
func CalcBalance(utxo []Addressutxos) (balance decimal.Decimal) {
	for i := range utxo {
		balance = utxo[i].Amount.Add(balance)
	}
	return
}

// CalcAmount 计算要转账总和
func CalcAmount(mapamount map[string]decimal.Decimal) (balance decimal.Decimal) {
	for i := range mapamount {
		balance = mapamount[i].Add(balance)
	}
	return
}

// CalcUtxo 转化为最终可用签名数据
func CalcUtxo(utxo []Addressutxos, mapamount map[string]decimal.Decimal) {

}

// [
//  [
//      {
//          "txid":"1fab5888cfb00b4123ef4e44f55ae230ee24382ebefcbff8e565cb16e6354156",
//          "vout":1
//      },
//      {
//          "txid":"d965a36c71abc5c02a117c602afa88f499c7b66e40964bb441f58b600aec1964",
//          "vout":1
//      }
//  ],
//  {
//  "RGd77ChrgHVNgYU4ptXaofCCBYnkho1xJr":8,
//  "RWAcLSVK742UVvStNVZEVFaV8gDFjrxGJt":1.9994
//  }
// ]

// UtxosMap .
type UtxosMap struct {
	Txid string `json:"txid"`
	Vout int    `json:"vout"`
}

// size:=input*148+34*out+10

// CreateTransactionList 创建批量交易 utxosMap 可用输入 mapamount 输出包含手续费
func (b *Coin) CreateTransactionList(utxosMap []UtxosMap, mapamount map[string]decimal.Decimal) (string, error) {
	var resp string
	err := b.rpc.CallMethod(&resp, "createrawtransaction", utxosMap, mapamount)
	return resp, err
}

// signrawtransaction

// SignResult .
type SignResult struct {
	Hex      string `json:"hex"`
	Complete bool   `json:"complete"`
}

// Signrawtransaction 创建签名hash ltc btc bch 有新接口替代
// func (b *Coin) Signrawtransaction(hex string) (SignResult, error) {
// 	var resp SignResult
// 	err := b.rpc.CallMethod(&resp, "signrawtransaction", hex)
// 	return resp, err
// }

// 500 Internal Server Error {"result":null,"error":{"code":-32,"message":"signrawtransaction is deprecated
// and will be fully removed in v0.18. To use signrawtransaction in v0.17, restart bitcoind with -deprecatedrpc=signrawtransaction.\n
// Projects should transition to using signrawtransactionwithkey and signrawtransactionwithwallet before upgrading to v0.18"},"id":6}

// Signrawtransactionwithwallet 创建签名hash 新版btc钱包已经修改为这个接口
func (b *Coin) Signrawtransactionwithwallet(hex string) (SignResult, error) {
	var resp SignResult
	err := b.rpc.CallMethod(&resp, "signrawtransactionwithwallet", hex)
	return resp, err
}

// Walletpassphrase 解锁钱包
func (b *Coin) Walletpassphrase(timeout int) error {
	err := b.rpc.CallMethod(nil, "walletpassphrase", b.WalletPwd, timeout)
	return err
}

// sendrawtransaction

// Sendrawtransaction  广播签名hash
func (b *Coin) Sendrawtransaction(sign string) (string, error) {
	var resp string
	err := b.rpc.CallMethod(&resp, "sendrawtransaction", sign)
	return resp, err
}

// getaddressesbyaccount

// Getaddressesbyaccount  获取钱包账户所有地址 高版本被删除
func (b *Coin) Getaddressesbyaccount() ([]string, error) {
	var resp []string
	err := b.rpc.CallMethod(&resp, "getaddressesbyaccount", "")
	return resp, err
}

// Getnewaddress 创建钱包的支付地址
// https://bitcoin.org/en/developer-reference#listunspent
func (b *Coin) Getnewaddress(account string) (string, error) {
	var resp string
	var err error
	if b.Name == "bch" {
		err = b.rpc.CallMethod(&resp, "getnewaddress", account)
	} else {
		// 高版本默认就是这个参数 p2sh-segwit legacy bech32
		err = b.rpc.CallMethod(&resp, "getnewaddress", account, "p2sh-segwit")
	}
	return resp, err
}

// Sends 综合返回
func (b *Coin) Sends(utxosMap []UtxosMap, mapamount map[string]decimal.Decimal) (string, error) {
	hex, err := b.CreateTransactionList(utxosMap, mapamount)
	if err != nil {
		return "", err
	}
	if b.WalletPwd != "" {
		err := b.Walletpassphrase(60)
		if err != nil {
			return "", errors.New(("解锁钱包失败 " + err.Error()))
		}
	}
	var sign SignResult

	sign, err = b.Signrawtransactionwithwallet(hex)

	if err != nil {
		return "", errors.New(("创建交易签名失败 " + err.Error()))
	}
	if !sign.Complete {
		return "", errors.New("创建交易签名失败 状态false 请检查是否合法")
	}
	txid, err := b.Sendrawtransaction(sign.Hex)
	if err != nil {
		return "", errors.New("广播签名交易失败 " + err.Error())
	}
	return txid, nil
}

// Transactions btc交易记录
type Transactions struct {
	Address       string          `json:"address"`
	FromAddress   string          `json:"fromaddress"`
	Category      string          `json:"category"`
	Amount        decimal.Decimal `json:"amount"`
	Confirmations int             `json:"confirmations"`
	Fee           decimal.Decimal `json:"fee"`
	Generated     bool            `json:"generated"`
	BlockHash     string          `json:"blockhash"`
	BlockIndex    int64           `json:"blockindex"` // 块高度 有可能是当前块索引 不能以这个作为参考
	BlockTime     int64           `json:"blocktime"`  // 没有用到
	TxID          string          `json:"txid"`
	Time          int64           `json:"time"` // txid 时间
	TimeReceived  int64           `json:"timereceived"`
}

// TransactionsList .
type TransactionsList = []Transactions

// Listtransactions 获取钱包btc交易记录
func (b *Coin) Listtransactions(args ...interface{}) (TransactionsList, error) {
	var resp TransactionsList
	err := b.rpc.CallMethod(&resp, "listtransactions", args...)
	return resp, err
}

// ValidateAddress 钱包合法检测
type ValidateAddress struct {
	IsValidate bool `json:"isvalid"`
}

// Validateaddress 验证钱包地址的合法性
func (b *Coin) Validateaddress(address string) (bool, error) {
	var resp ValidateAddress
	err := b.rpc.CallMethod(&resp, "validateaddress", address)
	return resp.IsValidate, err
}

// Info  钱包信息
type Info struct {
	Version         int         `json:"version"`
	ProtocolVersion int         `json:"protocolversion"`
	WalletVersion   int         `json:"walletversion"`
	Balance         json.Number `json:"balance"`
	Difficulty      float64     `json:"difficulty"`
	Blocks          int64       `json:"blocks"`
	Connections     int64       `json:"connections"`
	TimeOffset      int64       `json:"timeoffset"`
	Time            int64       `json:"time"`
}

// GetWalletInfo 钱包信息
// func (b *Coin) GetWalletInfo() (Info, error) {
// var resp Info
// err := b.rpc.CallMethod(&resp, "getinfo")
// return resp, err
// }

// GetWalletInfo 获取比特币新版本的钱包信息,必须是要支持比特币RPC接口才能启用
func (b *Coin) GetWalletInfo() (Info, error) {
	var resp Info
	walletInfo, err := b.GetNewWalletInfo()
	resp.Balance = walletInfo.Balance
	resp.WalletVersion = walletInfo.WalletVersion
	blockInfo, err := b.GetWalletBlockChainInfo()
	resp.Blocks = blockInfo.Blocks
	resp.Difficulty = blockInfo.Difficulty
	networkInfo, err := b.GetWalletNetworkInfo()
	resp.Connections = networkInfo.Connections
	resp.TimeOffset = networkInfo.TimeOffset
	resp.ProtocolVersion = networkInfo.ProtocolVersion
	resp.Version = networkInfo.Version
	return resp, err
}

// GetNewWalletInfo 新钱包信息
func (b *Coin) GetNewWalletInfo() (Info, error) {
	var resp Info
	err := b.rpc.CallMethod(&resp, "getwalletinfo")
	return resp, err
}

// GetWalletBlockChainInfo 获取钱包区块链信息
func (b *Coin) GetWalletBlockChainInfo() (Info, error) {
	var resp Info
	err := b.rpc.CallMethod(&resp, "getblockchaininfo")
	return resp, err
}

// GetWalletNetworkInfo 获取钱包网络信息
func (b *Coin) GetWalletNetworkInfo() (Info, error) {
	var resp Info
	err := b.rpc.CallMethod(&resp, "getnetworkinfo")
	return resp, err
}

// getbalance 所有账户余额

// {
// 	"txid": "fca4488ad8daccbaaf63fa78bfdf462b9da6bd30e2e935a29bc479f4fe010b60",
// 	"fee": "0.00001620",
// 	"sendingaddress": "1LAnF8h3qMGx3TSwNUHVneBZUEpwE4gu3D",
// 	"referenceaddress": "1254FSzQ7ULQYhCkJ9mdvfQbpN1AhqdShn",
// 	"ismine": true,
// 	"version": 0,
// 	"type_int": 0,
// 	"type": "Simple Send",
// 	"propertyid": 31,
// 	"divisible": true,
// 	"amount": "229.23860000",
// 	"valid": true,
// 	"blockhash": "0000000000000000002a4851b915765fc5346a122c581ce19dceb1cc06cba8b4",
// 	"blocktime": 1525588078,
// 	"positioninblock": 496,
// 	"block": 521447,
// 	"confirmations": 251
// }

// OmniTransactions omni代币交易记录
type OmniTransactions struct {
	Address       string `json:"referenceaddress"`
	FromAddress   string `json:"sendingaddress"`
	Propertyid    int    `json:"propertyid"` //代币
	Category      string `json:"type"`
	Amount        string `json:"amount"`
	Confirmations int    `json:"confirmations"`
	Fee           string `json:"fee"`
	Generated     bool   `json:"valid"`
	BlockHash     string `json:"blockhash"`
	BlockIndex    int64  `json:"block"`     // 块高度 btc高度
	BlockTime     int64  `json:"blocktime"` // 没有用到
	TxID          string `json:"txid"`
	// Time          int64  `json:"time"` // txid 时间  blocktime替代
}

// OmniTransactionsList .
type OmniTransactionsList = []OmniTransactions

// OmniListtransactions 获取代币交易记录
func (b *Coin) OmniListtransactions(args ...interface{}) (OmniTransactionsList, error) {
	var resp OmniTransactionsList
	err := b.rpc.CallMethod(&resp, "omni_listtransactions", args...)
	return resp, err
}

// omni_listblocktransactions

// [
// 			"LN6YuQmnASrAZ4NKgWytWRdCudKKNapzQn",
// 			"13X5s8Kz9a5rLrXXis7wXQYbY1DCSeRZ6i",
// 			31,
// 			"0.2"
// 		]

// omni_send
// https://github.com/OmniLayer/omnicore/issues/760  omni_funded_sendall

// OmniSend 发送代币 本地址保证有btc有余额
func (b *Coin) OmniSend(from, to string, propertyid int, amount string) (string, error) {
	var resp string
	err := b.rpc.CallMethod(&resp, "omni_send", from, to, propertyid, amount)
	return resp, err
}

// {
// 	"balance" : "n.nnnnnnnn",  // (string) the available balance of the address
// 	"reserved" : "n.nnnnnnnn", // (string) the amount reserved by sell offers and accepts
// 	"frozen" : "n.nnnnnnnn"    // (string) the amount frozen by the issuer (applies to managed properties only)
//   }

// omni_getbalance

// Omnibalance 地址代笔余额
type Omnibalance struct {
	Address    string `json:"address"`    // 地址
	Name       string `json:"name"`       // 代币名称
	Propertyid int    `json:"propertyid"` // 代币id
	Balance    string `json:"balance"`    // 余额
	Reserved   string `json:"reserved"`   // 保留数量
	Frozen     string `json:"frozen"`     // 被发行人冻结的数量
}

// OmniGetbalance 余额
func (b *Coin) OmniGetbalance(addr string, propertyid int) (Omnibalance, error) {
	var resp Omnibalance
	err := b.rpc.CallMethod(&resp, "omni_getbalance", addr, propertyid)
	return resp, err
}

// omni_getallbalancesforid

// OmnibalanceList 。
type OmnibalanceList = []Omnibalance

// OmniGetbalancesforid 线上所有指定代币余额
func (b *Coin) OmniGetbalancesforid(propertyid int) (OmnibalanceList, error) {
	var resp OmnibalanceList
	err := b.rpc.CallMethod(&resp, "omni_getallbalancesforid", propertyid)
	return resp, err
}

// omni_getallbalancesforaddress

// OmniGetwalletbalances 钱包所有代币余额
func (b *Coin) OmniGetwalletbalances() (OmnibalanceList, error) {
	var resp OmnibalanceList
	err := b.rpc.CallMethod(&resp, "omni_getwalletbalances")
	return resp, err
}

// omni_getallbalancesforaddress
// omni_getwalletaddressbalances

// omni_createpayload_simplesend

// OmniCreatepayloadSimplesend 创建简单交易
func (b *Coin) OmniCreatepayloadSimplesend(propertyid int, amount string) (string, error) {
	var resp string
	err := b.rpc.CallMethod(&resp, "omni_createpayload_simplesend", propertyid, amount)
	return resp, err
}

// omni_createrawtx_opreturn 签名btc后 负载简单交易

// OmniCreaterawtxOpreturn 负载简单交易
func (b *Coin) OmniCreaterawtxOpreturn(txBase, payload string) (string, error) {
	var resp string
	err := b.rpc.CallMethod(&resp, "omni_createrawtx_opreturn", txBase, payload)
	return resp, err
}

// omni_createrawtx_reference 添加转出地址到 负载交易中

// OmniCreaterawtxReference 添加转出地址到 负载交易中
func (b *Coin) OmniCreaterawtxReference(txBaseWithPayload, to string) (string, error) {
	var resp string
	err := b.rpc.CallMethod(&resp, "omni_createrawtx_reference", txBaseWithPayload, to)
	return resp, err
}

// v 创建最后交易 增加手续费 以及找零地址

// OmniCreaterawtxChange 增加手续费 以及找零地址
func (b *Coin) OmniCreaterawtxChange(txBaseWithPayloadAndTo, destination, fee string, prevtxs []Prevtxs) (string, error) {
	var resp string
	err := b.rpc.CallMethod(&resp, "omni_createrawtx_change", txBaseWithPayloadAndTo, prevtxs, destination, json.Number(fee))
	return resp, err
}

// SendsOmni 综合返回
func (b *Coin) SendsOmni(prevtxs []Prevtxs, mapamount map[string]decimal.Decimal, from, to, amount, fee string) (string, error) {
	var resp string
	return resp, nil
}
