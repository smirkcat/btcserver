// Package jsonrpc 调用封装
package coin

import (
	jRpc "github.com/ethereum/go-ethereum/rpc"
)

// JsonRpc .
type JsonRpc struct {
	host string
	port string
	user string
	pass string
	conn *jRpc.Client
}

// NewJsonRpc 创建json-rpc 请求
func NewJsonRpc(host, port, user, password string) *JsonRpc {
	rpc := &JsonRpc{host: host, port: port, user: user, pass: password}
	rpc.initJsonRpcConn()
	return rpc
}

var err error

//初始化JSON-RPC连接
func (j *JsonRpc) initJsonRpcConn() {
	j.conn, err = jRpc.Dial(j.getRPCAddress())
	if err != nil {
		panic("Init Json rpc conn client failed : " + err.Error())
	}
}

// CallMethod 呼叫JSON-RPC服务方案
func (j *JsonRpc) CallMethod(result interface{}, method string, args ...interface{}) error {

	err := j.conn.Call(result, method, args...)

	return err
}

// Close JSON-RPC 关闭
func (j *JsonRpc) Close() {
	j.conn.Close()
}

// 生成json-rpc的访问的服务地址
func (j *JsonRpc) getRPCAddress() string {
	start := `http://`
	if j.user != "" && j.pass != "" {
		return start + j.user + `:` + j.pass + `@` + j.host + `:` + j.port
	}
	return start + j.host + `:` + j.port
}
