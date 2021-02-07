package main

import "github.com/shopspring/decimal"

// Client 钱包节点
type Client struct {
	Host                  string          `toml:"rpc_host"`                // 钱包节点url
	Port                  string          `toml:"rpc_port"`                // omnirpc端口
	User                  string          `toml:"rpc_user"`                // user
	Pwd                   string          `toml:"rpc_pwd"`                 // pwd
	MainAddr              string          `toml:"main_addr"`               // 主钱包地址
	Symbol                string          `toml:"symbol"`                  // btc
	Precision             uint8           `toml:"precision"`               // 精度 默认8
	DBAddr                string          `toml:"db_addr"`                 // sqlite 地址
	CollectMin            decimal.Decimal `toml:"collect_min"`             // 最小归集代币数量
	CollectCron           string          `toml:"collect_cron"`            // 归集检测间隔cron表达式
	MonitorPort           int             `toml:"monitor_port"`            // 提供给清算系统rpc端口
	ScantransTimeInterval int             `toml:"scantrans_time_interval"` // 扫描交易记录间隔时间
	LogLevel              string          `toml:"log_level"`               // 日志等级
	WalletPwd             string          `toml:"wallet_pwd"`              // 钱包密码
}
