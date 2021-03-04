package main

import (
	"github.com/kinwyb/rpcx-code-generate/example/rpcx/serviceSev"
	"github.com/kinwyb/rpcx-code-generate/example/server/service"
	"github.com/smallnest/rpcx/server"
)

//go:generate rpcx-code-generate --sourcePath ./service --outPath ../rpcx
func main() {
	s := server.NewServer()
	g := serviceSev.RpcxServiceGroup{}
	g.SetArith(&service.Arith{})
	g.Services(s, "")
	s.Serve("tcp", "localhost:8798")
}
