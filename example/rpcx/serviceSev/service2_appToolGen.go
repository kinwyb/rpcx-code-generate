// 这是通过appTool自动生成的rpcx代码，请勿修改
package serviceSev

import (
	"context"

	"github.com/kinwyb/rpcx-code-generate/example/rpcx/serviceObj"
	"github.com/kinwyb/rpcx-code-generate/example/server/service"
)

func (a *arithRpcxService) MulDouble(reqCtx context.Context, arg serviceObj.RpcxRequestArithMulDouble, resp *serviceObj.RpcxResponseArithMulDouble) error {
	resp.S = a.Serv.MulDouble(arg.D, reqCtx)
	return nil
}
func NoResult(reqCtx context.Context, arg string, resp *string) error {
	service.NoResult()
	return nil
}
func GoFunc(reqCtx context.Context, arg string, resp *string) error {
	go service.GoFunc()
	return nil
}
func ContextFunc(reqCtx context.Context, arg string, resp *string) error {
	service.ContextFunc(reqCtx)
	return nil
}
func OneParamFunc(reqCtx context.Context, arg bool, resp *string) error {
	service.OneParamFunc(arg)
	return nil
}
