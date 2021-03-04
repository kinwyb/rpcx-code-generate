// 这是通过appTool自动生成的rpcx代码，请勿修改
package serviceSev

import (
	"github.com/kinwyb/rpcx-code-generate/example/server/service"
	rpcxServer "github.com/smallnest/rpcx/server"
)

type RpcxServiceGroup struct {
	arithSev *service.Arith
}

func (r *RpcxServiceGroup) SetArith(arithSev *service.Arith) {
	r.arithSev = arithSev
}
func (r *RpcxServiceGroup) Services(rpcxSev *rpcxServer.Server, meta string) {
	var err error
	if r.arithSev == nil {
		panic("RpcxServiceGroup[Arith]尚未初始化")
	}
	arithsev := &arithRpcxService{Serv: r.arithSev}
	err = rpcxSev.RegisterName("arithSev", arithsev, meta)
	if err != nil {
		panic("rpcx服务[arithSev]注册错误:" + err.Error())
	}
	err = rpcxSev.RegisterFunctionName("globalFun", "ArithMulFun", ArithMulFun, meta)
	if err != nil {
		panic("rpcx函数[ArithMulFun]注册错误:" + err.Error())
	}
	err = rpcxSev.RegisterFunctionName("globalFun", "NoResult", NoResult, meta)
	if err != nil {
		panic("rpcx函数[NoResult]注册错误:" + err.Error())
	}
	err = rpcxSev.RegisterFunctionName("globalFun", "GoFunc", GoFunc, meta)
	if err != nil {
		panic("rpcx函数[GoFunc]注册错误:" + err.Error())
	}
	err = rpcxSev.RegisterFunctionName("globalFun", "ContextFunc", ContextFunc, meta)
	if err != nil {
		panic("rpcx函数[ContextFunc]注册错误:" + err.Error())
	}
	err = rpcxSev.RegisterFunctionName("globalFun", "OneParamFunc", OneParamFunc, meta)
	if err != nil {
		panic("rpcx函数[OneParamFunc]注册错误:" + err.Error())
	}
}
