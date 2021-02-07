package wallet

import (
	"encoding/json"
)

// Info  钱包信息
type Info struct {
	Version         int         `json:"version"`
	ProtocolVersion int         `json:"protocolversion"`
	WalletVersion   int         `json:"walletversion"`
	Balance         json.Number `json:"balance"`
	Difficulty      int64       `json:"difficulty"`
	Blocks          int64       `json:"blocks"`
	Connections     int64       `json:"connections"`
	TimeOffset      int64       `json:"timeoffset"`
	Time            int64       `json:"time"`
}

// Transactions ..
type Transactions struct {
	Account       string      `json:"account"`
	TxID          string      `json:"txid"`
	Address       string      `json:"address"`
	FromAddress   string      `json:"fromaddress"`
	Category      string      `json:"category"`
	Amount        json.Number `json:"amount"`
	Fee           json.Number `json:"fee"`
	Vout          int         `json:"vout"`
	Confirmations int64       `json:"confirmations"`
	Generated     bool        `json:"generated"`
	BlockHash     string      `json:"blockhash"`
	BlockIndex    int64       `json:"blockindex"`
	BlockTime     int64       `json:"blocktime"`
	Time          int64       `json:"time"`
	TimeReceived  int64       `json:"timereceived"`
}

// SummaryData 归集中转记录
type SummaryData struct {
	TxID        string `json:"txid"`
	Account     string `json:"account"`
	Address     string `json:"address"`
	FromAddress string `json:"fromaddress"`
	Amount      string `json:"amount"`
	BlockIndex  int64  `json:"blockindex"`
	Blocktime   int64  `json:"blocktime"`
	Category    string `json:"category"`
	Fee         string `json:"fee"`
	Time        int64  `json:"time"`
}

// ValidateAddress 钱包合法检测
type ValidateAddress struct {
	IsValidate bool `json:"isvalid"`
}

// Address 钱包地址
type Address struct {
	StandardAddress string `json:"standard_address"`
	PaymentID       string `json:"payment_id"`
}

//IntegrateAddress string `json:"integrated_address"`
