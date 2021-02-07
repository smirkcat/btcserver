// +build !collect

package main

import (
	"btcallserver/log"
	coin "btcallserver/wallet"

	"github.com/shopspring/decimal"
)

// 检测是否有足够的手续费且是否能被花费 0 没有足够余额 1 有足够余额 但是有未被确认的等待被确认 2 有足够余额且被确认
func isenoughbtc(addr string, feePerbyte int) (status int) {
	utxo, err := node.Listunspent([]string{addr})
	if err != nil {
		log.Error(err)
	}
	utxo05, err := node.Listunspent05([]string{addr})
	if err != nil {
		log.Error(err)
	}
	sizeout := 34*2 + 34 + 10 + 40 // 手续费找零 多40包含可能字节变多
	if len(utxo) > 0 {
		sizein := len(utxo) * 148
		sizeall := sizein + sizeout
		shuoldfee := decimal.New(int64(sizeall*feePerbyte), 0).Div(btcdanwei).Add(omnito)
		balancebtc := coin.CalcBalance(utxo)
		log.Infof("%s 可用手续费 %s 实际归集花费需要 %s", addr, balancebtc, shuoldfee)
		if shuoldfee.LessThanOrEqual(balancebtc) {
			status = 2
			return
		}
		// 如果上面不满足估算的 且没有未确认的 则满足一个最小值 则也可以传输
		// 这里satoshiPerByte设置最小 5 satoshi/byte 设置2 返回200 错误 不断尝试
		if len(utxo05) < 1 {
			shuoldfeemin := decimal.New(int64(sizeall*satoshiPerByte), 0).Div(btcdanwei).Add(omnito)
			log.Infof("%s 可用手续费 %s 实际归集花费最小需要 %s", addr, balancebtc, shuoldfeemin)
			if shuoldfeemin.LessThanOrEqual(balancebtc) {
				status = 2
				return
			}
		}
	}

	if len(utxo05) > 0 {
		sizein := len(utxo05)*148 + len(utxo)*148
		sizeall := sizein + sizeout
		shuoldfee := decimal.New(int64(sizeall*feePerbyte), 0).Div(btcdanwei)
		balancebtc := coin.CalcBalance(utxo05).Add(coin.CalcBalance(utxo))
		log.Infof("%s 还有未确认手续费 %s 未到6个确认", addr, coin.CalcBalance(utxo05))
		log.Infof("%s 将可用手续费 %s 实际归集花费需要 %s (此处直接划转不在判断余额是否满足，每个字节手续费是在变化的)", addr, balancebtc, shuoldfee)
		status = 1
		return
	}
	status = 0
	return
}

// 传输 btc 最大100 个
func tranferbtc(mapbtc map[string]decimal.Decimal, feePerbyte int) {
	utxo, err := node.Listunspent([]string{feeAddr})
	if err != nil {
		log.Error(err)
	}
	log.Info("开始手续费转账")
	defer func() {
		// 不管成功或者失败都需要清空
		for k := range mapbtc {
			delete(mapbtc, k)
		}
	}()
	lenutxo := len(utxo)
	if lenutxo < 0 {
		log.Info("没有足够的手续费")
	}
	sizeout := 34*len(mapbtc) + 34 + 10 // 手续费找零
	maxoutfee := decimal.New(int64(sizeout*feePerbyte), 0).Div(btcdanwei)
	perutxofee := decimal.New(int64(148*feePerbyte), 0).Div(btcdanwei)
	utoxfee := decimal.Zero
	vouttotal := decimal.Zero
	total := coin.CalcAmount(mapbtc).Add(decimal.New(int64(len(mapbtc)), 0).Mul(omnito))
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
		log.Infof("可用btc余额 %s 不足 应该有 %s 请检查钱包主地址btc余额\n", vouttotal.String(), shouldtotal.String())
		return
	}
	btcfee := utoxfee.Add(maxoutfee)
	mapbtc[feeAddr] = vouttotal.Sub(btcfee).Sub(total)

	txid, err := node.Sends(utxosMap, mapbtc)
	if err != nil {
		log.Infof("此次手续费转账失败，%v\n", err)
		return
	}
	log.Infof("此次手续费转账返回txid:%s,等待确认中\n", txid)
}

func collect() {
	log.Info("开始处理归集检测")
	defer log.Info("结束处理归集检测")

	sizeout := 34*2 + 34 + 10 + 40 // 手续费找零 多40包含可能字节变多
	sizein := 148
	sizeall := sizein + sizeout
	fee, err := node.GetFee()
	if err != nil {
		log.Error(err)
	}
	shuoldfee := decimal.New(int64(sizeall*fee.FastestFee), 0).Div(btcdanwei)
	if shuoldfee.LessThan(feeMin) {
		shuoldfee = feeMin
	}
	from := 0
	count := 700
	var mapbtc = make(map[string]decimal.Decimal, 0) // 需要转手续费的地址
	for {
		log.Infof("开始处理账户数量%d,页码%d", count, from/count+1)
		account, err := dbengine.GetAccount(count, from)
		if err != nil {
			log.Error(err)
		}
		lens := len(account)
		for i := 0; i < lens; i++ {
			// 这个地方检测余额 暂时不并行处理，一般本地连接处理很快的 后期可通过交易记录处理
			balance, err := node.OmniGetbalance(account[i].Address, propertyid)
			if err != nil {
				log.Infof("获取当前第%d个地址%s余额失败err:%v", i, account[i].Address, err)
				continue
			}
			amount, _ := decimal.NewFromString(balance.Balance)
			if amount.GreaterThanOrEqual(collectMin) {
				ret := isenoughbtc(account[i].Address, fee.FastestFee)
				if ret == 0 {
					log.Infof("归集地址%s到主账户地址%s数量%s,手续费不够等待手续费转账",
						account[i].Address, mainAddr, amount)
					// 这里设置一个固定值
					mapbtc[account[i].Address] = shuoldfee
					if len(mapbtc) >= 10 {
						tranferbtc(mapbtc, fee.FastestFee)
					}
				} else if ret == 2 {
					// 开始归集
					txid, err := node.OmniSend(account[i].Address, mainAddr, propertyid, amount.String())
					if err != nil {
						log.Errorf("归集地址%s到主账户地址%s数量%s,返回err:%s",
							account[i].Address, mainAddr, amount, err.Error())
					} else {
						log.Infof("归集地址%s到主账户地址%s数量%s,返回txid:%s",
							account[i].Address, mainAddr, amount, txid)
					}
				} else {
					log.Infof("归集地址%s到主账户地址%s数量%s,手续费还在等待被确认",
						account[i].Address, mainAddr, amount)
				}
			}
		}
		// 结束所有账户
		if lens < count {
			break
		}
		from = from + count
	}
	// 处理剩下的btc手续费
	if len(mapbtc) > 0 {
		tranferbtc(mapbtc, fee.FastestFee)
	}
}
