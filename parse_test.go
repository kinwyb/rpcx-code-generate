package rpcx_code_generate

import "testing"

func TestParseFile_Generate(t *testing.T) {
	filePath := "./example/server"
	f := &parseDir{
		dirPath: filePath,
		outpath: "./example/rpcx",
	}
	err := f.Generate()
	if err != nil {
		t.Fatal(err)
	}
	err = f.OutputToFile()
	if err != nil {
		t.Fatal(err)
	}
}
