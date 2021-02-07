package main

import "github.com/shopspring/decimal"

// Client 钱包节点
type Client struct {
	Host                  string          `toml:"rpc_host"`                // 钱包节点url
	Port                  string          `toml:"rpc_port"`                // omnirpc端口
	User                  string          `toml:"rpc_user"`                // user
	Pwd                   string          `toml:"rpc_pwd"`                 // pwd
	MainAddr              string          `toml:"main_addr"`               // 主钱包地址
	FeeAddr               string          `toml:"fee_addr"`                // 手续费地址
	Propertyid            int             `toml:"propertyid"`              // 代币id 默认31
	Symbol                string          `toml:"symbol"`                  // usdt
	Precision             uint8           `toml:"precision"`               // 精度 默认8
	DBAddr                string          `toml:"db_addr"`                 // sqlite 地址
	FeeMin                decimal.Decimal `toml:"fee_min"`                 // 最小单次转账手续费
	CollectMin            decimal.Decimal `toml:"collect_min"`             // 最小归集代币数量
	CollectTimeInterval   int             `toml:"collect_time_interval"`   // 归集检测间隔时间单位分钟
	MonitorPort           int             `toml:"monitor_port"`            // 提供给清算系统rpc端口
	ScantransTimeInterval int             `toml:"scantrans_time_interval"` // 扫描交易记录间隔时间
	LogLevel              string          `toml:"log_level"`               // 日志等级
	SatoshiPerByte        int             `toml:"satoshi_per_byte"`        // 归集每字节最花费检测单位 最小5 默认5
	WalletPwd             string          `toml:"wallet_pwd"`              // 钱包密码
}
