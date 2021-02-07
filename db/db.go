package db

import (
	"strconv"

	"github.com/go-xorm/xorm"
	_ "github.com/mattn/go-sqlite3"
	"github.com/shopspring/decimal"
)

const (
	Send    = "send"    // 提币 即 主地址提币到其他地址
	Receive = "receive" // 本平台地址 分配的用户地址收到的
	Collect = "collect" // 本平台地址归集到主地址
)

var dbengine *DB

// OtherParam .
type OtherParam struct {
	Key   string `xorm:"'key' unique"`
	Value string `xorm:"'value'"`
}

// TableName 表名
func (fh OtherParam) TableName() string {
	return "param"
}

// Account .
type Account struct {
	ID       int64           `xorm:"'id' pk autoincr"`
	Address  string          `xorm:"'address' text unique(addressid)"`
	Amount   decimal.Decimal `xorm:"'amount' real default '0'"`
	User     string          `xorm:"'user' text"`
	Ctime    int64           `xorm:"'ctime' integer"`
	Lasttime int64           `xorm:"'lasttime' integer"`
}

// TableName 表名
func (fh Account) TableName() string {
	return "account"
}

// Transfers btc交易记录
type Transfers struct {
	ID         int64           `xorm:"'id' pk autoincr" json:"-"`
	TxID       string          `xorm:"'txid' text index" json:"txid"`
	Blockindex int64           `xorm:"'blockindex' integer" json:"blockindex"`
	Address    string          `xorm:"'address' text index" json:""`
	Blockhash  string          `xorm:"'blockhash' text" json:"blockhash"`
	Amount     decimal.Decimal `xorm:"'amount' real" json:"amount"`
	Fee        decimal.Decimal `xorm:"'fee' real" json:"fee"` // 保留字段
	Timestamp  int64           `xorm:"'time' integer"`
	Category   string          `xorm:"'category' text index"`
}

// TableName 表名
func (fh Transfers) TableName() string {
	return "transfers"
}

// Transferlog omni代币交易记录
type Transferlog struct {
	ID         int64           `xorm:"'id' pk autoincr" json:"-"`
	TxID       string          `xorm:"'txhash'" json:"txid"`
	Fromaddr   string          `xorm:"'fromaddr' index" json:""`
	Toaddr     string          `xorm:"'toaddr' index" json:""`
	Blockindex int64           `xorm:"'blockindex' default '0'"`
	Amount     decimal.Decimal `xorm:"'amount' real" json:"amount"`
	Fee        decimal.Decimal `xorm:"'fee' real default '0'" json:"fee"` // 保留字段
	Timestamp  int64           `xorm:"'time' integer"`
	Category   string          `xorm:"'category' default 'send'"` // send recive collect
}

// TableName 表名
func (fh Transferlog) TableName() string {
	return "transferlog"
}

// DB .
type DB struct {
	*xorm.Engine
}

// Close 关闭数据库引擎
func (db *DB) Close() {
	db.Close()
}

// Session 创建事务
func (db *DB) Session() *xorm.Session {
	return db.NewSession()
}

// InitDB 初始化数据库
func InitDB(url string) (*DB, error) {
	engine, err := xorm.NewEngine("sqlite3", url)
	//设置连接池的空闲数大小
	engine.SetMaxIdleConns(10)
	//设置最大打开连接数
	engine.SetMaxOpenConns(300)
	return &DB{
		Engine: engine,
	}, err
}

// Sync 同步数据库结构
func (db *DB) Sync() error {
	return db.Sync2(new(Account), new(Transfers), new(Transferlog), new(OtherParam))
}

// SearchAccount 搜索账户是否存在
func (db *DB) SearchAccount(addr string) (*Account, error) {
	var tmp Account
	ok, err := db.Where("address = ?", addr).Limit(1).Get(&tmp)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}
	return &tmp, err
}

// GetAccount 获取所有账户
func (db *DB) GetAccount(count, skip int) ([]Account, error) {
	var tmp = make([]Account, 0)
	if skip < 0 {
		skip = 0
	}
	if count < 1 || count > 1000 {
		count = 1000
	}

	err := db.Limit(count, skip).Find(&tmp)
	return tmp, err
}

// GetTransactions 获取最近交易记录
func (db *DB) GetTransactions(addr string, count, skip int) ([]Transferlog, error) {
	var tmp = make([]Transferlog, 0)

	if count < 1 || count > 1000 {
		count = 300
	}
	if skip < 0 {
		skip = 0
	}
	tmpdb := db.Limit(count, skip).Where("category=? OR category=?", Send, Receive)
	if addr != "*" && addr != "" {
		tmpdb = tmpdb.Where("address = ?", addr)
	}
	err := tmpdb.Desc("time").Find(&tmp)
	return tmp, err
}

// GetTransactionsCollect 获取指定时间段归集
func (db *DB) GetTransactionsCollect(sTime, eTime int64) ([]Transferlog, error) {
	var tmp = make([]Transferlog, 0)
	tmpdb := db.Where("time >= ? and time <=? and type=?", sTime, eTime, Collect)
	err := tmpdb.Desc("time").Find(&tmp)
	return tmp, err
}

// SearchTransactions 搜索交易记录是否存在
func (db *DB) SearchTransactions(txid string, fromaddr, toaddr, category string) (*Transferlog, error) {
	var tmp Transferlog
	ok, err := db.Table(Transferlog{}.TableName()).Where("txhash = ? and fromaddr = ? and toaddr = ? and category=?",
		txid, fromaddr, toaddr, category).Exist()
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}
	return &tmp, err
}

// SearchBtcTransactions 搜索交易记录是否存在
func (db *DB) SearchBtcTransactions(txid string, addr, category string) (*Transfers, error) {
	var tmp Transfers
	ok, err := db.Table(Transfers{}.TableName()).Where("txid = ? and address = ? and category = ?",
		txid, addr, category).Exist()
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}
	return &tmp, err
}

// InsertAccount 插入数据
func (db *DB) InsertAccount(account *Account) (int64, error) {
	return db.Insert(account)
}

// InsertTransactions 插入数据
func (db *DB) InsertTransactions(transactions *Transferlog) (int64, error) {
	return db.Insert(transactions)
}

// UpdateTransactions 插入数据 暂时没用
func (db *DB) UpdateTransactions(transactions *Transferlog) (int64, error) {
	return db.Cols("").Insert(transactions)
}

// InsertBtcTransactions 插入数据
func (db *DB) InsertBtcTransactions(transactions *Transfers) (int64, error) {
	return db.Insert(transactions)
}

// LoadLastBlockHeight 获取最后一次扫描高度 已经扫描到这个高度
func (db *DB) LoadLastBlockHeight() (int64, error) {
	var tmp OtherParam
	ok, err := db.Where("key='block'").Limit(1).Get(&tmp)
	if err != nil || !ok {
		return 0, err
	}
	var un int64
	un, err = strconv.ParseInt(tmp.Value, 10, 0)
	return un, err
}

// InsertLastBlockHeight 更新最后一次扫描高度
func (db *DB) InsertLastBlockHeight(num int64) (err error) {
	var ok bool
	var tmp = OtherParam{
		Key: "block",
	}
	ok, err = db.Exist(&tmp)
	if err != nil || !ok {
		_, err = db.Insert(&tmp)
	} else {
		tmp.Value = strconv.FormatInt(num, 10)
		_, err = db.Where("key='block'").Cols("value").Update(&tmp)
	}
	return
}
