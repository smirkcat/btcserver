# rpc 测试

## btc-test
IP=192.168.1.11
PORT=18332
USER=user
PWD=8e7OmEaXPalNB7kvoFnsuGhjfu1YI5ajVa4vKoPD

## bch-test
IP=192.168.1.32
PORT=18332
USER=wawawa
PWD=123456789

## ltc-test
IP=192.168.1.11
PORT=9333
USER=wawawa
PWD=123456789

IP=192.168.1.4
PORT=8333
USER=admin #wawawa
PWD=123456
METHOD='getinfo' getwalletinfo omni_getwalletbalances
PARAMS='[]'
curl -X POST http://$USER:$PWD@$IP:$PORT \
-d '{"jsonrpc":"1.0","id":"1111111","method":"'$METHOD'","params":'$PARAMS'}' \
-H 'Content-Type: application/json'
```json
{"result":{"version":130200,"protocolversion":70015,"walletversion":130000,"balance":0.06800171,"blocks":1664816,"timeoffset":-1,"connections":8,"proxy":"","difficulty":12563071.03178775,"testnet":true,"keypoololdest":1577340193,"keypoolsize":100,"paytxfee":0.00000000,"relayfee":0.00001000,"errors":"Warning: Unknown block versions being mined! It's possible unknown rules are in effect"},"error":null,"id":"222222w22"}
```

IP=127.0.0.1
PORT=8384
METHOD='getnewaddress'
PARAMS='["golang","p2sh-segwit"]'
curl -X POST http://$USER:$PWD@$IP:$PORT \
-d '{"jsonrpc":"2.0","id":"222222","method":"'$METHOD'","params":'$PARAMS'}' \
-H 'Content-Type: application/json'

{"id":"222222","jsonrpc":"2.0","result":"dfkdkfldlfkdlfk"}

IP=192.168.1.4
PORT=8333
USER=admin
PWD=123456
METHOD='getaddressesbyaccount' # V0.18 deprecated
PARAMS='["mytest"]'
curl -X POST http://$USER:$PWD@$IP:$PORT \
-d '{"jsonrpc":"2.0","id":"222222","method":"'$METHOD'","params":'$PARAMS'}' \
-H 'Content-Type: application/json'

```json mytest
{"result":["mpguyNJZL4rWXqKXQM8SZk3okSCM5dGWox","mt6i5AAA8J1PphJDshmT5BSK28ELfVASyy","n3xch7rUFE9iUbkd9Gmoe9BLwbhMo8bNWp"],"error":null,"id":"222222w22"}
```

```json 
{"result":["mmV4dJbhGFf7NN9aRiDMhvUGyHeEihsq8y","n4QyZxdUSg29pGPLwSVhTZFDfnXu8KkjMs"],"error":null,"id":"222222w22"}
```

```json golang
{"result":["mwAzHKfEEvtXHR7ZttDjLnBzV5yjP9b7nZ","n3V1FowaAxWV68LJp5xJhR1qEbyfsD29kC"],"error":null,"id":"222222w22"}
```

// mt6i5AAA8J1PphJDshmT5BSK28ELfVASyy n3xch7rUFE9iUbkd9Gmoe9BLwbhMo8bNWp
{"result":"mt6i5AAA8J1PphJDshmT5BSK28ELfVASyy","error":null,"id":"1111111"}

{"error":{"code":2,"data":null,"message":"Account not found"},"id":"222222","jsonrpc":"2.0"}

IP=127.0.0.1
PORT=8384
METHOD='validateaddress'
PARAMS='["n4QyZxdUSg29pGPLwSVhTZFDfnXu8KkjMs"]'
curl -X POST http://$USER:$PWD@$IP:$PORT \
-d '{"jsonrpc":"2.0","id":"222222","method":"'$METHOD'","params":'$PARAMS'}' \
-H 'Content-Type: application/json'

{"id":"222222","jsonrpc":"2.0","result":{"address":"n4QyZxdUSg29pGPLwSVhTZFDfnXu8KkjMs","ismine":false,"isvalid":true}}

IP=192.168.1.4
PORT=8333
USER=admin
PWD=123456
METHOD='listaddressgroupings' # listaccounts ..btc V0.18 deprecated 
curl -X POST http://$USER:$PWD@$IP:$PORT \
-d '{"jsonrpc":"2.0","id":"222222","method":"'$METHOD'"}' \
-H 'Content-Type: application/json'

```json
{"result":{"":-0.11087091,"golang":0.05411381,"mt6i5AAA8J1PphJDshmT5BSK28ELfVASyy":0.00000000,"mytest":0.12475881},"error":null,"id":"222222w22"}
```

```json
{"result":[[["mmxL38k284XZvXKnwNTFNUTVRxHiZWEQ16",0.04999775],["mwAzHKfEEvtXHR7ZttDjLnBzV5yjP9b7nZ",0.00000000,"golang"],["n3V1FowaAxWV68LJp5xJhR1qEbyfsD29kC",0.01710396,"golang"],["n3xch7rUFE9iUbkd9Gmoe9BLwbhMo8bNWp",0.00000000,"mytest"]],[["mpguyNJZL4rWXqKXQM8SZk3okSCM5dGWox",0.00090000,"mytest"]]],"error":null,"id":"222222w22"}
```

omni_listtransactions
IP=127.0.0.1
PORT=8384
METHOD='listtransactions'
PARAMS='["*",100,0]'
curl -X POST http://$USER:$PWD@$IP:$PORT \
-d '{"jsonrpc":"2.0","id":"222222","method":"'$METHOD'","params":'$PARAMS'}' \
-H 'Content-Type: application/json'	
```json
{"id":"222222","jsonrpc":"2.0","result":[{"account":"zmtest123","address":"n3xch7rUFE9iUbkd9Gmoe9BLwbhMo8bNWp","amount":1.0900000000000001,"blockindex":139651751342000,"blocktime":1502576499,"category":"receive","confirmations":1,"fee":0,"time":1502576499,"timereceived":1502576499,"txid":"bc15205c20c499b71d4b9a1304bb8e2beff44c944db1d6be72de7b7fe78c7535","vout":1}]}
```
```json
{"result":[{"txid":"4e5d0beae5b592dc806b4c98c93fd00cdbe60e19a3a3a1adf9832673351ffb0e","fee":"0.00000257","sendingaddress":"n3xch7rUFE9iUbkd9Gmoe9BLwbhMo8bNWp","referenceaddress":"mwAzHKfEEvtXHR7ZttDjLnBzV5yjP9b7nZ","ismine":true,"version":0,"type_int":0,"type":"Simple Send","propertyid":1,"divisible":true,"amount":"6.00000000","valid":true,"blockhash":"00000000000124905900b46a20d4c912c393802ff6bf4c4fde65f749d0dba970","blocktime":1580902775,"positioninblock":395,"block":1665015,"confirmations":139}],"error":null,"id":"222222"}
```

IP=192.168.1.4
PORT=8333
USER=admin
PWD=123456
METHOD='listunspent'
PARAMS='[0,5,["2N4WvGSA5NkTtgpMMhbTf3xFeLrJmAuDSh3"]]'
curl -X POST http://$USER:$PWD@$IP:$PORT \
-d '{"jsonrpc":"2.0","id":"1111111","method":"'$METHOD'","params":'$PARAMS'}' \
-H 'Content-Type: application/json'

IP=192.168.1.4
PORT=39842
METHOD='eth_getBlockByNumber'
PARAMS='["0x2c0e37",true]'
curl -X POST http://$IP:$PORT \
-d '{"jsonrpc":"2.0","id":"1111111","method":"'$METHOD'","params":'$PARAMS'}' \
-H 'Content-Type: application/json'

METHOD='eth_getTransactionReceipt'
PARAMS='["0x2e18f1196e65dad15f3c0db1832ec4e90b7e8377cb3113966d047bfc7345906c"]'
