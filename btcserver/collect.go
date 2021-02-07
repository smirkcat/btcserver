package main

import (
	"btcallserver/log"
	coin "btcallserver/wallet"
	"sync"

	"github.com/robfig/cron/v3"
	"github.com/shopspring/decimal"
)

var c *cron.Cron

func cronCollect() {
	c = cron.New(cron.WithLocation(log.CSTZone))
	_, err := c.AddFunc(globalConf.CollectCron, collect) //用法 https://godoc.org/github.com/robfig/cron
	if err != nil {
		log.Fatal(err)
	}
	c.Start()
}

var isruncollect = false
var collectclock sync.Mutex

func collect() {
	if isruncollect {
		log.Info("上次归集未完成")
		return
	}
	collectclock.Lock()
	isruncollect = true
	log.Info("开始处理归集检测")
	defer func() {
		log.Info("结束处理归集检测")
		isruncollect = false
		collectclock.Unlock()
	}()
	from := 0
	count := 700
	for {
		infotmp1, _ := node.GetNewWalletInfo()
		balancetmp1, _ := decimal.NewFromString(infotmp1.Balance.String())
		totalfee := decimal.Zero
		log.Infof("开始处理账户数量%d,页码%d", count, from/count+1)
		account, err := dbengine.GetAccount(count, from)
		if err != nil {
			log.Error(err)
		}
		lens := len(account)
		for i := 0; i < lens; i++ {
			// 这个地方检测余额 暂时不并行处理，一般本地连接处理很快的 后期可通过交易记录处理
			utxos, err := node.Listunspent([]string{account[i].Address})
			if err != nil {
				log.Errorf("获取当前第%d个地址%s可花费输入失败,err:%v", i, account[i].Address, err)
				continue
			}
			lensutxo := len(utxos)
			if lensutxo < 1 {
				continue
			}
			vouttotal := decimal.Zero
			var utxosMap []coin.UtxosMap
			for i := 0; i < lensutxo; i++ {
				vouttotal = vouttotal.Add(utxos[i].Amount)
				tmp := coin.UtxosMap{
					Txid: utxos[i].Txid, // 后面检验txid来源地址是否可靠
					Vout: utxos[i].Vout,
				}
				utxosMap = append(utxosMap, tmp)
			}
			// input*148+34*out+10
			fee, err := node.GetFee()
			if err != nil {
				log.Error(err)
			}

			sizein := 148 * lensutxo
			sizeout := 34 + 10 // 手续费找零
			sizeall := sizein + sizeout
			shuoldfee := decimal.New(int64(sizeall*fee.HalfHourFee), 0).Div(btcdanwei)

			var mapbtc = make(map[string]decimal.Decimal, 0) // 需要转手续费的地址
			mapbtc[mainAddr] = vouttotal.Sub(shuoldfee)

			log.Infof("归集地址%s输入条数%d,总额%s,手续费%s,主地址%s实际收入%s. 归集最小值为%s",
				account[i].Address, lensutxo, vouttotal, shuoldfee, mainAddr, mapbtc[mainAddr], collectMin)
			if vouttotal.LessThanOrEqual(shuoldfee.Add(collectMin)) {
				log.Infof("归集地址%s撤销归集不满足最小归集总量", account[i].Address)
				continue
			}

			txid, err := node.Sends(utxosMap, mapbtc)
			if err != nil {
				log.Errorf("此次%s归集转账失败，%v", account[i].Address, err)
				return
			}
			totalfee = totalfee.Add(shuoldfee)
			log.Infof("此次%s归集转账返回txid:%s,等待确认中", account[i].Address, txid)
		}
		infotmp2, _ := node.GetNewWalletInfo()
		balancetmp2, _ := decimal.NewFromString(infotmp2.Balance.String())
		balancetmp := balancetmp1.Sub(balancetmp2)
		log.Infof("处理完成账户数量%d,页码%d,总共手续费%s,钱包余额减少%s,账户余额是否匹配%v",
			lens, from/count+1, totalfee, balancetmp, balancetmp.Equal(totalfee))
		// 结束所有账户
		if lens < count {
			break
		}
		from = from + count
	}
}
