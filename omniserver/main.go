package main

import (
	"btcallserver/daemon"
	wallet "btcallserver/data"
	"btcallserver/db"
	"btcallserver/log"
	coin "btcallserver/wallet"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/shopspring/decimal"
	"github.com/spf13/cobra"
)

const (
	// name of the service
	name        = "omnirpc"
	description = "omni coin rpcservice"
)

var sigChan = make(chan os.Signal, 1) //用于系统信息接收处理的通道
var httpClose = make(chan struct{})   //用于接收server停止通道
var server = &http.Server{
	Addr: ":6543",
} // rpc httpserver
var exit = make(chan struct{})

// flag args
var (
	configfile string
	verbose    bool
	h          bool
)

var rootCmd = &cobra.Command{
	Use:   "",
	Short: "",
	//Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if verbose {
			printVersion()
			return
		}
		if h {
			cmd.Help()
		}
		preFun()
		defer sufFun()
		// 阻塞
		go daemon.HandleSystemSignal(sigChan, stop)
		<-httpClose
		close(sigChan)
	},
}

func initCmd() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "version", "v", false, "version omniserver")
	rootCmd.PersistentFlags().BoolVarP(&h, "help", "h", false, "help of omniserver")
	rootCmd.PersistentFlags().StringVarP(&configfile, "config", "c", "omni.toml", "启动文件")
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Version omniserver",
		Run: func(cmd *cobra.Command, args []string) {
			printVersion()
		},
	}
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(daemon.Cmds(namedesc)...)
}

func namedesc() (string, string) {
	return name, description
}

func init() {
	initCmd()
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var (
	BuildVersion string
	BuildDate    string
)

// Version .
const Version = "omnirpc Version --v1.2.1"

func timePrint() string {
	return time.Now().Local().Format("2006-01-02T15:04:05.000Z07:00")
}

func preFun() {
	fmt.Printf("omnirpc start, time=%s\n", timePrint())
	Init()
	Serv(server, httpClose)
}

func sufFun() {
	fmt.Printf("omnirpc exit, time=%s\n", timePrint())
}

//显示版本信息
func printVersion() {
	fmt.Println(Version)
	if BuildDate != "" {
		fmt.Println("omnirpc BuildDate --" + BuildDate)
	}
	if BuildVersion != "" {
		fmt.Println("omnirpc BuildVersion --" + BuildVersion)
	}
}

func stop() {
	server.Close()
	close(exit)
}

var curr = getCurrentDirectory() + `/`

func getCurrentDirectory() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return "."
	}
	return strings.Replace(dir, "\\", "/", -1)
}

// InitLog 初始化日志文件
func InitLog() {
	var logConfigInfoName, logConfigErrorName, logLevel string
	logConfigInfoName = curr + globalConf.Symbol + "rpc.log"
	logConfigErrorName = curr + globalConf.Symbol + "rpc-err.log"
	logLevel = globalConf.LogLevel
	log.Init(logConfigInfoName, logConfigErrorName, logLevel)
}

var globalConf Client

var dbengine *db.DB

var mainAddr string
var feeAddr string
var node *coin.Coin                           //钱包节点
var propertyid int = 31                       //代币符号
var feeMin = decimal.NewFromFloat32(0.000035) //最小单次手续费划转数量
var collectMin = decimal.NewFromFloat32(4)    //最小归集数量
var collectTimeInterval = 30                  //归集检测时间
var satoshiPerByte = 5                        //归集最小手续费单位判断

// Init 初始化相关参数
func Init() {
	if _, err := toml.DecodeFile(curr+configfile, &globalConf); err != nil {
		fmt.Println(err)
		_, err = toml.Decode(string(getConfig()), &globalConf)
		if err != nil {
			panic(err)
		}
	}
	globalConf.Symbol = strings.ToLower(globalConf.Symbol)
	InitLog()
	var err error
	dbengine, err = db.InitDB(globalConf.DBAddr)
	if err != nil {
		panic(err)
	}
	err = dbengine.Sync()
	if err != nil {
		panic(err)
	}
	node = coin.NewWalletCoin(globalConf.Symbol, globalConf.Host, globalConf.Port,
		globalConf.User, globalConf.Pwd, globalConf.WalletPwd)
	if globalConf.Propertyid > 0 {
		propertyid = globalConf.Propertyid
	}
	if globalConf.CollectMin.GreaterThan(decimal.Zero) {
		collectMin = globalConf.CollectMin
	}
	if globalConf.CollectTimeInterval > 0 {
		collectTimeInterval = globalConf.CollectTimeInterval
	}
	if globalConf.SatoshiPerByte > 5 {
		satoshiPerByte = globalConf.SatoshiPerByte
	}
	if globalConf.FeeMin.GreaterThan(decimal.Zero) {
		feeMin = globalConf.FeeMin
	}
	if globalConf.MonitorPort != 0 {
		server.Addr = ":" + strconv.Itoa(globalConf.MonitorPort)
	}
	mainAddr = globalConf.MainAddr
	feeAddr = globalConf.FeeAddr
	targetHeight = getlastBlock()
	getwalletinfo()
	log.Info(walletinfo)
	go task()
}

//获取默认的数据库配置
func getConfig() []byte {
	return []byte(`
rpc_host="127.0.0.1"
rpc_port="10001" 
rpc_user="admin"
rpc_pwd="123456"  
main_addr="mwAzHKfEEvtXHR7ZttDjLnBzV5yjP9b7nZ"
fee_addr="mmxL38k284XZvXKnwNTFNUTVRxHiZWEQ16" #有些手续费
propertyid=1
symbol="usdt"
precision=8
db_addr="C:/Users/smirkcat/go/src/btcallserver/omni.db"
fee_min=0.000035
monitor_port=6543
collect_min=4
collect_time_interval=30
scantrans_time_interval=60
satoshi_per_byte=5 # 归集每字节最花费检测单位 最小5 默认5
`)
}

func task() {
	var collectT = globalConf.CollectTimeInterval
	if collectT < 1 {
		collectT = 30
	}
	var scanT = globalConf.ScantransTimeInterval
	if scanT < 1 {
		scanT = 60
	}
	var timec = time.Duration(collectT) * time.Minute
	var times = time.Duration(scanT) * time.Second
	tikerc := time.NewTicker(timec)
	tikers := time.NewTicker(times)
	defer tikerc.Stop()
	defer tikers.Stop()
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		for {
			select {
			case <-exit:
				wg.Done()
				return
			case <-tikerc.C:
				collect()
			}
		}
	}()
	go func() {
		for {
			select {
			case <-exit:
				wg.Done()
				return
			case <-tikers.C:
				getwalletinfo()
				scanomnitranslog()
				scanbtctrans()
			}
		}
	}()
	wg.Wait()
}

var btcdanwei = decimal.NewFromInt(10000_0000)
var omnito = decimal.NewFromFloat(0.00000546)

// 0.00035

var minScanBlock int64 = 0 // 最小扫描高度
var targetHeight int64     // 扫描到区块高度
var blockHeightTop int64   //最新区块高度

func getlastBlock() int64 {
	block, err := dbengine.LoadLastBlockHeight()
	if err != nil || block < minScanBlock {
		block = minScanBlock
	}
	return block
}

// 扫描omni交易记录
func scanomnitranslog() {
	log.Info("开始扫描交易记录，扫描高度", targetHeight)
	count := 1000
	from := 0
	startHeight := targetHeight
	for {
		re, err := node.OmniListtransactions("*", count, from, startHeight, 999999999)
		if err != nil {
			log.Errorf("OmniListtransactions get err: %v", err)
		}
		lens := len(re)
		for i := 0; i < lens; i++ {
			// 没被验证不是有效转账
			if !re[i].Generated {
				continue
			}
			tmplog := &db.Transferlog{
				TxID:       re[i].TxID,
				Fromaddr:   re[i].FromAddress,
				Toaddr:     re[i].Address,
				Blockindex: re[i].BlockIndex,
				Category:   strings.ToLower(re[i].Category),
				Timestamp:  re[i].BlockTime,
			}
			// 扫描到的最高区块高度 下次从里开始 可能会重复跳过
			if tmplog.Blockindex > targetHeight {
				targetHeight = tmplog.Blockindex
				dbengine.InsertLastBlockHeight(targetHeight)
			}
			tmplog.Amount, _ = decimal.NewFromString(re[i].Amount)
			tmplog.Fee, _ = decimal.NewFromString(re[i].Fee)

			// 这里不可能有手续费地址参与
			if tmplog.Fromaddr == mainAddr {
				ok, _ := dbengine.SearchAccount(tmplog.Toaddr)
				if ok == nil {
					tmplog.Category = db.Send
				} else {
					tmplog.Category = db.Collect
				}
			} else {
				ok, _ := dbengine.SearchAccount(tmplog.Fromaddr)
				if ok != nil {
					// 这里应该进行详细判断 一般是归集
					tmplog.Category = db.Collect
				} else {
					ok1, _ := dbengine.SearchAccount(tmplog.Toaddr)
					if ok1 != nil {
						tmplog.Category = db.Receive
					} else {
						// 这里应该进行详细判断 这里有可能是充值或者归集
						tmplog.Category = db.Collect
					}
				}
			}
			// Send Receive
			ret, _ := dbengine.SearchTransactions(tmplog.TxID,
				tmplog.Fromaddr, tmplog.Toaddr, tmplog.Category)
			if ret == nil {
				_, err = dbengine.InsertTransactions(tmplog)
				if err != nil {
					log.Errorf("InsertTransactions %v err: %v\n", tmplog, err)
				}
			}
		}
		if lens < count {
			break
		}
		from = from + count
	}
}

// 扫描btc交易记录
func scanbtctrans() {
	count := 1000
	from := 0
	flag := false
	for {
		re, err := node.Listtransactions("*", count, from)
		if err != nil {
			log.Infof("Listtransactions get err: %v", err)
		}
		lens := len(re)
		for i := 0; i < lens; i++ {
			tmplog := &db.Transfers{
				TxID:       re[i].TxID,
				Address:    re[i].Address,
				Blockindex: blockHeightTop - int64(re[i].Confirmations),
				Category:   strings.ToLower(re[i].Category),
				Timestamp:  re[i].Time,
				Amount:     re[i].Amount,
				Fee:        re[i].Fee,
			}
			if tmplog.Blockindex < targetHeight {
				flag = true
			}
			if tmplog.Address == "" {
				continue
			}
			// move 不记录
			if tmplog.Category != db.Send && tmplog.Category != db.Receive {
				continue
			}
			// Send Receive
			ret, _ := dbengine.SearchBtcTransactions(re[i].TxID, re[i].Address, tmplog.Category)
			if ret == nil {
				_, err = dbengine.InsertBtcTransactions(tmplog)
				if err != nil {
					log.Infof("InsertTransactions err: %v", err)
				}
			}
		}
		if lens < count || flag {
			break
		}
		from = from + count
	}
}

var walletinfo wallet.Info

// 获取钱包信息
func getwalletinfo() {
	info, err := node.GetWalletInfo()
	if err != nil {
		log.Infof("GetWalletInfo err %v", err)
	} else {
		walletinfo.Blocks = info.Blocks
		blockHeightTop = info.Blocks
		walletinfo.Connections = info.Connections
		walletinfo.Difficulty = int64(info.Difficulty)
		walletinfo.ProtocolVersion = info.ProtocolVersion
		walletinfo.Time = info.Time
		walletinfo.TimeOffset = info.TimeOffset
		walletinfo.Version = info.Version
	}
	balance, err := node.OmniGetwalletbalances()
	if err != nil {
		log.Infof("GetWalletInfo err %v", err)
	} else {
		for k := range balance {
			if balance[k].Propertyid == propertyid {
				walletinfo.Balance = json.Number(balance[k].Balance)
				break
			}
		}
	}
}
