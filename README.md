## BTC系列钱包代理程序

1. 包含 btc bch ltc
2. omni-usdt

### 一、 OMNI-USDT
OMNI 协议的USDT 钱包代理程序


#### ubuntu 切换镜像源
```shell
sudo cp /etc/apt/sources.list /etc/apt/sources.list.back
vim /etc/apt/sources.list
:%s/security.ubuntu/mirrors.aliyun/g
:%s/archive.ubuntu/mirrors.aliyun/g
sudo apt update
```


##### 编译  
```shell
cd omniserver
./make.sh
```
##### 直接启动
```shell
./omniserver
# 或者
nohup ./omniserver -c omni.toml  >omni-run.log 2>&1 &
```

##### 服务方式
服务名为 omnirpc
1. 安装
```shell
sudo ./omniserver install -c omni.toml
```

2. 卸载服务
```shell
sudo ./omniserver remove
```
3. 启动服务
```shell
sudo ./omniserver start  # 当前路径下面必须有 omni.toml配置文件
# 或者
sudo systemctl start omnirpc
```

4. 停止服务
```shell
sudo ./omniserver stop
或者
sudo systemctl stop omnirpc
```
5. 服务状态
```shell
./omniserver status
或者
systemctl status omnirpc
```
以上  -c omni.toml 参数可忽略，默认参数

### 二、BTC BCH LTC

1. 交易记录转发，不记录
2. 记录地址，用于归集检测

#### 编译
```
cd btcserver
./make.sh
```

#### 直接启动
```shell
./btcserver -c [btc|ltc|bch].toml
# 或者
nohup ./btcserver -c [btc|ltc|bch].toml  >[btc|ltc|bch]-run.log 2>&1 &
```

#### 服务方式

服务名以-s 参数确定 默认btc 则服务名为btcrpc
1. 安装服务
```shell
sudo ./btcserver install -c [btc|ltc|bch].toml -s [btc|ltc|bch]
```
2. 卸载服务
```shell
sudo ./btcserver remove -s [btc|ltc|bch]
```
3. 启动服务
```shell
sudo ./btcserver start -s [btc|ltc|bch]
# 或者
sudo systemctl start [btc|ltc|bch]rpc
```
4. 停止服务
```shell
sudo ./btcserver stop -s [btc|ltc|bch]
或者
sudo systemctl stop [btc|ltc|bch]rpc
```
5. 服务状态
```shell
./btcserver status -s [btc|ltc|bch]
或者
systemctl status [btc|ltc|bch]rpc
```

[*Change Log*](CHANGELOG.md)