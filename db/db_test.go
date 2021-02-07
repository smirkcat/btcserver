package db

import (
	"fmt"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
)

var url = "D:/go/src/btcallserver/bch.db"

func TestRunDb(t *testing.T) {
	InitDBTest()
}

func InitDBTest() {
	re, err := InitDB(url)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = re.Sync()

	if err != nil {
		fmt.Println(err)
		return
	}
	// ret, err := re.SearchTransactions("4e5d0beae5b592dc806b4c98c93fd00cdbe60e19a3a3a1adf9832673351ffb0e",
	// 	"n3xch7rUFE9iUbkd9Gmoe9BLwbhMo8bNWp", "mwAzHKfEEvtXHR7ZttDjLnBzV5yjP9b7nZ", "simple send")
	// fmt.Println(ret, err)
}

var urlbtc = "D:/go/src/btcallserver/btc.db"
var urladdr = "uvbitpay:7q4a1z@tcp(192.168.1.10:3306)/uvbitpay?charset=utf8"

func TestDBMoveAddress(t *testing.T) {
	re, err := InitDB(urlbtc)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = re.Sync()
	if err != nil {
		fmt.Println(err)
		return
	}

	engine, err := xorm.NewEngine("mysql", urladdr)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("开始处理")
	skip := 0
	for {
		var tmp []Account
		err := engine.Table("bgd_address_pool").Cols("address").Asc("id").
			Where("coin_name = 'btc'").Limit(1000, skip).Find(&tmp)
		if err != nil {
			fmt.Println(err)
			return
		}

		lens := len(tmp)
		if lens < 1 {
			break
		}
		fmt.Printf("拉取数量%d 跳过总量%d\n", lens, skip)
		skip += 1000
		ds := re.NewSession()
		for i := 0; i < lens; i++ {
			tmp := &Account{
				Address: tmp[i].Address,
			}
			_, err := ds.Insert(tmp)
			if err != nil {
				fmt.Println(err)
			}
		}
		ds.Commit()
		ds.Close()
	}
	fmt.Println("结束处理")
}
