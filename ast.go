package rpcx_code_generate

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/token"
	"golang.org/x/tools/imports"
	"runtime"
	"sort"
)

// decls排序
type sortableDecls []ast.Decl

func (sd sortableDecls) Len() int {
	return len(sd)
}

func (sd sortableDecls) Less(i int, j int) bool {
	switch left := sd[i].(type) {
	case *ast.GenDecl:
		switch right := sd[j].(type) {
		default:
			return left.Tok == token.IMPORT
		case *ast.GenDecl:
			return left.Tok == token.IMPORT && right.Tok != token.IMPORT
		}
	}
	return false
}

func (sd sortableDecls) Swap(i int, j int) {
	sd[i], sd[j] = sd[j], sd[i]
}

func formatNode(fname string, node ast.Node) ([]byte, error) {
	if node == nil {
		return []byte{}, nil
	}
	if file, is := node.(*ast.File); is {
		sort.Stable(sortableDecls(file.Decls))
	}
	outfset := token.NewFileSet()
	buf := &bytes.Buffer{}
	err := format.Node(buf, outfset, node)
	if err != nil {
		return nil, err
	}
	var ostype = runtime.GOOS
	if ostype == "windows" && fname == "" {
		fname = "\\tmp.go"
	}
	imps, err := imports.Process(fname, buf.Bytes(), nil)
	if err != nil {
		return nil, err
	}
	return bytes.NewBuffer(imps).Bytes(), nil
}

func sel(ids ...*ast.Ident) ast.Expr {
	switch len(ids) {
	default:
		return &ast.SelectorExpr{
			X:   sel(ids[:len(ids)-1]...),
			Sel: ids[len(ids)-1],
		}
	case 1:
		return ids[0]
	case 0:
		panic("zero ids to Sel()")
	}
}

func fieldList(list ...*ast.Field) *ast.FieldList {
	return &ast.FieldList{List: list}
}

func field(n *ast.Ident, t ast.Expr) *ast.Field {
	return &ast.Field{
		Names: []*ast.Ident{n},
		Type:  t,
	}
}

// 函数结构体
func funcDecl() *ast.FuncDecl {
	return &ast.FuncDecl{
		Recv: &ast.FieldList{},
		Type: &ast.FuncType{
			Params: nil,
			//Results: &ast.FieldList{},
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{},
		},
	}
}

// importFor ast import代码
func importFor(is *ast.ImportSpec) *ast.GenDecl {
	return &ast.GenDecl{Tok: token.IMPORT, Specs: []ast.Spec{is}}
}
