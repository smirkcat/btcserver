package main

import (
	wallet "btcallserver/data"
	"btcallserver/db"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/semrush/zenrpc"
)

// go get -u -v github.com/semrush/zenrpc/zenrpc

// Service omni 钱包信息
type Service struct{ zenrpc.Service }

// Getinfo  获取钱包信息
func (as Service) Getinfo() wallet.Info {
	return walletinfo
}

// GetNewAddress  获取新地址
func (as Service) GetNewAddress(account string) (string, error) {
	return creataddress(account)
}

// ValidateAddress 校验地址
func (as Service) ValidateAddress(addr string) wallet.ValidateAddress {
	var resp wallet.ValidateAddress
	resp.IsValidate = validaddress(addr)
	return resp
}

// ListTransactions 获取指定地址最近的交易记录
//zenrpc:count=300
//zenrpc:skip=0
//zenrpc:addr="*"
func (as Service) ListTransactions(addr string, count, skip int) ([]wallet.Transactions, error) {
	return recentTransactions(addr, count, skip)
}

// SendToAddress .
func (as Service) SendToAddress(addr string, amount json.Number) (string, error) {
	return sendfromMainAddr(addr, amount)
}

// GetRecords . 归集交易记录 中转记录
func (as Service) GetRecords(sTime, eTime int64) ([]wallet.SummaryData, error) {
	return collectTransactions(sTime, eTime)
}

//go:generate zenrpc

// Serv 监听服务
func Serv(server *http.Server, httpClose chan struct{}) {
	rpc := zenrpc.NewServer(zenrpc.Options{ExposeSMD: true})
	rpc.Register("", Service{})
	//rpc.Use(zenrpc.Logger(log.New(os.Stderr, "", log.LstdFlags)))
	httpw := http.NewServeMux()
	httpw.Handle("/", rpc)
	server.Handler = httpw
	go func() {
		err := server.ListenAndServe()
		if err != nil {
			fmt.Println(err)
		}
		close(httpClose)
	}()
}

func creataddress(account string) (string, error) {
	account = "golang-uvbitpay" // 这里暂时固定
	addr, err := node.Getnewaddress(account)
	if err != nil {
		return addr, err
	}
	timenow := time.Now().Unix()
	acc := &db.Account{
		Address:  addr,
		User:     account,
		Ctime:    timenow,
		Lasttime: timenow,
	}
	_, err = dbengine.InsertAccount(acc)
	return addr, err
}

func validaddress(addr string) bool {
	flag, _ := node.Validateaddress(addr)
	return flag
}

func recentTransactions(addr string, count, skip int) ([]wallet.Transactions, error) {
	var trans []wallet.Transactions
	translog, err := node.Listtransactions(addr, count, skip)
	if err != nil {
		return trans, err
	}
	lens := len(translog)
	trans = make([]wallet.Transactions, lens)
	for i := 0; i < lens; i++ {
		trans[i].TxID = translog[i].TxID
		//trans[i].FromAddress = translog[i].Fromaddr
		trans[i].Address = translog[i].Address
		trans[i].Category = translog[i].Category
		trans[i].Amount = json.Number(translog[i].Amount.String())
		trans[i].Fee = json.Number(translog[i].Fee.String())
		trans[i].Confirmations = int64(translog[i].Confirmations)
		trans[i].Time = translog[i].Time
		trans[i].BlockIndex = blockHeightTop - trans[i].Confirmations
		trans[i].Account = "go-" + globalConf.Symbol + "-walletrpc"
	}
	return trans, nil
}

func collectTransactions(sTime, eTime int64) ([]wallet.SummaryData, error) {
	var collects = make([]wallet.SummaryData, 0)
	// trans, err := dbengine.GetTransactionsCollect(sTime, eTime)
	// if err != nil {
	// 	return collects, err
	// }
	// lens := len(trans)
	// collects = make([]wallet.SummaryData, lens)
	// for i := 0; i < lens; i++ {
	// 	collects[i].TxID = trans[i].TxID
	// 	collects[i].Address = trans[i].Toaddr
	// 	collects[i].FromAddress = trans[i].Fromaddr
	// 	collects[i].Category = trans[i].Category
	// 	collects[i].Amount = trans[i].Amount.String()
	// 	collects[i].Fee = trans[i].Fee.String()
	// 	collects[i].Time = trans[i].Timestamp
	// 	collects[i].BlockIndex = trans[i].Blockindex
	// 	collects[i].Account = "go-omni-walletrpc"
	// }
	return collects, nil
}
