// 这是通过appTool自动生成的rpcx代码，请勿修改
package serviceSev

import (
	"context"

	"github.com/kinwyb/rpcx-code-generate/example/rpcx/serviceObj"
	"github.com/kinwyb/rpcx-code-generate/example/server/service"
)

type arithRpcxService struct {
	Serv *service.Arith
}

func (a *arithRpcxService) ServicePath() string {
	return "arithSev"
}
func (a *arithRpcxService) Mul(reqCtx context.Context, arg serviceObj.RpcxRequestArithMul, resp *serviceObj.RpcxResponseArithMul) error {
	resp.I = a.Serv.Mul(arg.A1, arg.B1)
	return nil
}
func (a *arithRpcxService) One(reqCtx context.Context, arg int, resp *serviceObj.RpcxResponseArithOne) error {
	resp.S = a.Serv.One(arg)
	return nil
}
func (a *arithRpcxService) TestContext(reqCtx context.Context, arg string, resp *string) error {
	a.Serv.TestContext(reqCtx)
	return nil
}
func ArithMulFun(reqCtx context.Context, arg serviceObj.RpcxRequestArithMulFun, resp *serviceObj.RpcxResponseArithMulFun) error {
	resp.I = service.ArithMulFun(arg.A1, arg.B1)
	return nil
}
