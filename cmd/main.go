package main

import (
	"flag"
	rpcx_code_generate "github.com/kinwyb/rpcx-code-generate"
)

var (
	sourcePath = flag.String("sourcePath", "./", "服务地址")
	outPath    = flag.String("outPath", "./", "输出地址")
)

func main() {
	flag.Parse()
	f := rpcx_code_generate.NewParseDir(*sourcePath, *outPath)
	err := f.Generate()
	if err != nil {
		println("err : " + err.Error())
	}
	err = f.OutputToFile()
	if err != nil {
		println("err : " + err.Error())
	}
}
