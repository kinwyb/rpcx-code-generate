// 这是通过appTool自动生成的rpcx代码，请勿修改
package serviceClt

import (
	"context"
	"errors"

	"github.com/kinwyb/rpcx-code-generate/example/rpcx/serviceObj"
)

func (a *arithRpcxClient) MulDouble(D int, ctx context.Context) []int {
	var err error
	req := &serviceObj.RpcxRequestArithMulDouble{D: D}
	resp := &serviceObj.RpcxResponseArithMulDouble{}
	err = a.client.Call(ctx, "MulDouble", req, resp)
	if err != nil {
		if lg != nil {
			lg.Errorf("[service2 - Arith.MulDouble]RPC调用错误:%s", err.Error())
		}
	}
	return resp.S
}
func NoResult() {
	var err error
	if gClient == nil {
		err = errors.New("gClient尚未初始化")
	} else {
		err = gClient.Call(context.Background(), "NoResult", nil, nil)
	}
	if err != nil {
		if lg != nil {
			lg.Errorf("[service2 - .NoResult]RPC调用错误:%s", err.Error())
		}
	}
	return
}
func GoFunc() {
	var err error
	if gClient == nil {
		err = errors.New("gClient尚未初始化")
	} else {
		err = gClient.Call(context.Background(), "GoFunc", nil, nil)
	}
	if err != nil {
		if lg != nil {
			lg.Errorf("[service2 - .GoFunc]RPC调用错误:%s", err.Error())
		}
	}
	return
}
func ContextFunc(ctx context.Context) {
	var err error
	if gClient == nil {
		err = errors.New("gClient尚未初始化")
	} else {
		err = gClient.Call(ctx, "ContextFunc", nil, nil)
	}
	if err != nil {
		if lg != nil {
			lg.Errorf("[service2 - .ContextFunc]RPC调用错误:%s", err.Error())
		}
	}
	return
}
func OneParamFunc(x bool) {
	var err error
	if gClient == nil {
		err = errors.New("gClient尚未初始化")
	} else {
		err = gClient.Call(context.Background(), "OneParamFunc", x, nil)
	}
	if err != nil {
		if lg != nil {
			lg.Errorf("[service2 - .OneParamFunc]RPC调用错误:%s", err.Error())
		}
	}
	return
}
