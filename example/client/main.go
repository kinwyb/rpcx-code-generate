package main

import (
	"context"
	"fmt"
	"github.com/kinwyb/rpcx-code-generate/example/rpcx/serviceClt"
	"github.com/smallnest/rpcx/client"
)

func main() {
	d, _ := client.NewPeer2PeerDiscovery("tcp@localhost:8798", "")
	serviceClt.RpcxClients(d, client.Failtry, client.RandomSelect, client.DefaultOption)
	defer serviceClt.RpcxClose()
	fmt.Printf("5+6 = %d\n", serviceClt.ArithSev.Mul(5, 6))
	fmt.Printf("7+8 = %d\n", serviceClt.ArithMulFun(7, 8))
	serviceClt.NoResult()
	serviceClt.GoFunc()
	fmt.Printf("Funcs.GoFunc \n")
	serviceClt.ContextFunc(context.Background())
	r := serviceClt.ArithSev.One(6)
	fmt.Printf("Arith One %v", r)
}
