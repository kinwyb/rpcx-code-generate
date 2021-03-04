// 这是通过appTool自动生成的rpcx代码，请勿修改
package serviceClt

import (
	"github.com/sirupsen/logrus"
	"github.com/smallnest/rpcx/client"
)

var (
	gClient client.XClient
	lg      = logrus.New()
)
var ArithSev *arithRpcxClient

func RpcxClients(discovery client.ServiceDiscovery, failMode client.FailMode, selectMode client.SelectMode, option client.Option) {
	if ArithSev == nil {
		ArithSev = &arithRpcxClient{client: client.NewXClient("arithSev", failMode, selectMode, discovery, option)}
	}
	if gClient == nil {
		gClient = client.NewXClient("globalFun", failMode, selectMode, discovery, option)
	}
}
func RpcxClose() {
	if ArithSev != nil {
		ArithSev.client.Close()
		ArithSev = nil
	}
	if gClient != nil {
		gClient.Close()
		gClient = nil
	}
}
func SetLogger(l *logrus.Logger) {
	lg = l
}
