package rpcx_code_generate

import "go/ast"

// 代码解析存储对象
type SourceContext struct {
	Pkg        *ast.Ident
	Imports    []*ast.ImportSpec
	Types      []*StructType
	Funcs      []*fun
	StructFunc map[string][]*fun //结构体关联的函数
	Prefix     string
	Common     []*ast.Comment
}

// 导出源文件中所有import包去重
func (sc *SourceContext) ImportDecls() (decls []ast.Decl) {
	have := map[string]struct{}{}
	notHave := func(is *ast.ImportSpec) bool {
		if _, has := have[is.Path.Value]; has {
			return false
		}
		have[is.Path.Value] = struct{}{}
		return true
	}
	for _, is := range sc.Imports {
		if notHave(is) {
			decls = append(decls, importFor(is))
		}
	}
	return
}
