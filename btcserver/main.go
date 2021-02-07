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

var (
	// name of the service
	name        = "rpc"
	description = " coin rpcservice"
)

var sigChan = make(chan os.Signal, 1) //用于系统信息接收处理的通道
var httpClose = make(chan struct{})   //用于接收server停止通道
var server = &http.Server{
	Addr: ":10333",
} // rpc httpserver
var exit = make(chan struct{})

// flag args
var (
	configfile  string
	verbose     bool
	h           bool
	servicename string
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
	rootCmd.PersistentFlags().BoolVarP(&verbose, "version", "v", false, "version btcserver")
	rootCmd.PersistentFlags().StringVarP(&servicename, "service", "s", "btc", "systemd service prefix")
	rootCmd.PersistentFlags().BoolVarP(&h, "help", "h", false, "help of btcserver")
	rootCmd.PersistentFlags().StringVarP(&configfile, "config", "c", "btc.toml", "启动文件")
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Version btcserver",
		Run: func(cmd *cobra.Command, args []string) {
			printVersion()
		},
	}
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(daemon.Cmds(namedesc)...)
}

func namedesc() (string, string) {
	return servicename + name, servicename + description
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
const Version = "btc-all-rpc Version --v1.2.1"

func timePrint() string {
	return time.Now().Local().Format("2006-01-02T15:04:05.000Z07:00")
}

func preFun() {
	fmt.Printf("btcallrpc start, time=%s\n", timePrint())
	Init()
	Serv(server, httpClose)
}

func sufFun() {
	fmt.Printf("btcallrpc exit, time=%s\n", timePrint())
}

func printVersion() {
	fmt.Println(Version)
	if BuildDate != "" {
		fmt.Println("btcallrpc BuildDate --" + BuildDate)
	}
	if BuildVersion != "" {
		fmt.Println("btcallrpc BuildVersion --" + BuildVersion)
	}
}

func stop() {
	c.Stop()
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

var globalConf Client

var dbengine *db.DB

var mainAddr string
var node *coin.Coin                            // 钱包节点
var collectMin = decimal.NewFromFloat32(0.001) // 单次最小归集总数量

// InitLog 初始化日志文件
func InitLog() {
	var logConfigInfoName, logConfigErrorName, logLevel string
	logConfigInfoName = curr + globalConf.Symbol + "rpc.log"
	logConfigErrorName = curr + globalConf.Symbol + "rpc-err.log"
	logLevel = globalConf.LogLevel
	log.Init(logConfigInfoName, logConfigErrorName, logLevel)
}

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

	if globalConf.CollectMin.GreaterThan(decimal.Zero) {
		collectMin = globalConf.CollectMin
	}
	if globalConf.MonitorPort != 0 {
		server.Addr = ":" + strconv.Itoa(globalConf.MonitorPort)
	}
	mainAddr = globalConf.MainAddr
	//collect()
	getwalletinfo()
	log.Info(walletinfo)
	go task()
}

//获取默认的数据库配置
func getConfig() []byte {
	return []byte(`
rpc_host="192.168.1.11"
rpc_port="18332" 
rpc_user="user"
rpc_pwd="8e7OmEaXPalNB7kvoFnsuGhjfu1YI5ajVa4vKoPD"
wallet_pwd="789456123"
main_addr="2MvyN7B8pyYdD2g8nj7owinoeDiv7yyxfKu"
symbol="btc"
precision=8
db_addr="D:/go/src/btcallserver/btc.db"
monitor_port=10332
collect_min=0.001
scantrans_time_interval=60
log_level="info"
collect_cron="@hourly"
`)
}

func task() {
	cronCollect()
	var scanT = globalConf.ScantransTimeInterval
	if scanT < 1 {
		scanT = 60
	}
	var times = time.Duration(scanT) * time.Second
	tikers := time.NewTicker(times)
	defer tikers.Stop()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		for {
			select {
			case <-exit:
				wg.Done()
				return
			case <-tikers.C:
				getwalletinfo()
			}
		}
	}()
	wg.Wait()
}

var btcdanwei = decimal.NewFromInt(10000_0000)

var blockHeightTop int64 //最新区块高度

var walletinfo wallet.Info

// 获取钱包信息
func getwalletinfo() {
	info, err := node.GetWalletInfo()
	if err != nil {
		log.Errorf("GetWalletInfo err %v", err)
	} else {
		walletinfo.Blocks = info.Blocks
		blockHeightTop = info.Blocks
		walletinfo.Connections = info.Connections
		walletinfo.Difficulty = int64(info.Difficulty)
		walletinfo.ProtocolVersion = info.ProtocolVersion
		walletinfo.Time = info.Time
		walletinfo.TimeOffset = info.TimeOffset
		walletinfo.Version = info.Version
		walletinfo.Balance = info.Balance
	}
}

// 都是从主账户提币 需要排对提币
var locksend sync.Mutex

func sendfromMainAddr(addr string, amount json.Number) (string, error) {
	if addr == mainAddr {
		return "", fmt.Errorf("非法的提币到主地址")
	}
	locksend.Lock()
	defer locksend.Unlock()
	utxo, err := node.Listunspent([]string{mainAddr})
	if err != nil {
		return "", err
	}
	lenutxo := len(utxo)
	if lenutxo < 0 {
		return "", fmt.Errorf("主账户没有余额")
	}
	var mapbtc = make(map[string]decimal.Decimal, 0) // 需要转手续费的地址
	mapbtc[addr], _ = decimal.NewFromString(amount.String())
	sizeout := 34*len(mapbtc) + 34 + 10 // 手续费找零
	// input*148+34*out+10
	fee, err := node.GetFee()
	if err != nil {
		log.Error(err)
	}

	//fmt.Println(utxo)

	maxoutfee := decimal.New(int64(sizeout*fee.FastestFee), 0).Div(btcdanwei)
	perutxofee := decimal.New(int64(148*fee.FastestFee), 0).Div(btcdanwei)
	utoxfee := decimal.Zero
	vouttotal := decimal.Zero
	total := coin.CalcAmount(mapbtc)
	shouldtotal := total.Add(maxoutfee)
	utxosMap := make([]coin.UtxosMap, 0)
	for i := 0; i < lenutxo; i++ {
		v := utxo[i]
		tmp := v.Amount
		utoxfee = utoxfee.Add(perutxofee)
		vouttotal = vouttotal.Add(tmp)
		tmputxo := coin.UtxosMap{
			Txid: v.Txid,
			Vout: v.Vout,
		}
		utxosMap = append(utxosMap, tmputxo)
		if vouttotal.Sub(utoxfee).GreaterThanOrEqual(shouldtotal) {
			break
		}
	}
	if vouttotal.Sub(utoxfee).LessThan(shouldtotal) {
		return "", fmt.Errorf("可用余额%s不足,应该有%s,请检查钱包主地址余额", vouttotal.String(), shouldtotal.String())
	}
	btcfee := utoxfee.Add(maxoutfee)
	mapbtc[mainAddr] = vouttotal.Sub(btcfee).Sub(total)
	txid, err := node.Sends(utxosMap, mapbtc)
	if err != nil {
		return "", err
	}
	return txid, nil
}
