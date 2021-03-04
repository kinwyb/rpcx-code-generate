// 这是通过appTool自动生成的rpcx代码，请勿修改
package serviceClt

import (
	"context"
	"errors"

	"github.com/kinwyb/rpcx-code-generate/example/rpcx/serviceObj"
	"github.com/smallnest/rpcx/client"
)

type arithRpcxClient struct {
	client client.XClient
}

func (a *arithRpcxClient) Init(client client.XClient) {
	a.client = client
}
func (a *arithRpcxClient) ServicePath() string {
	return "arithSev"
}
func (a *arithRpcxClient) Mul(A1 int, B1 int) int {
	var err error
	req := &serviceObj.RpcxRequestArithMul{A1: A1, B1: B1}
	resp := &serviceObj.RpcxResponseArithMul{}
	err = a.client.Call(context.Background(), "Mul", req, resp)
	if err != nil {
		if lg != nil {
			lg.Errorf("[service - Arith.Mul]RPC调用错误:%s", err.Error())
		}
	}
	return resp.I
}
func (a *arithRpcxClient) One(d int) []int {
	var err error
	resp := &serviceObj.RpcxResponseArithOne{}
	err = a.client.Call(context.Background(), "One", d, resp)
	if err != nil {
		if lg != nil {
			lg.Errorf("[service - Arith.One]RPC调用错误:%s", err.Error())
		}
	}
	return resp.S
}
func (a *arithRpcxClient) TestContext(ctx context.Context) {
	var err error
	err = a.client.Call(ctx, "TestContext", nil, nil)
	if err != nil {
		if lg != nil {
			lg.Errorf("[service - Arith.TestContext]RPC调用错误:%s", err.Error())
		}
	}
	return
}
func ArithMulFun(A1 int, B1 int) int {
	var err error
	req := &serviceObj.RpcxRequestArithMulFun{A1: A1, B1: B1}
	resp := &serviceObj.RpcxResponseArithMulFun{}
	if gClient == nil {
		err = errors.New("gClient尚未初始化")
	} else {
		err = gClient.Call(context.Background(), "ArithMulFun", req, resp)
	}
	if err != nil {
		if lg != nil {
			lg.Errorf("[service - .ArithMulFun]RPC调用错误:%s", err.Error())
		}
	}
	return resp.I
}
