package rpcx_code_generate

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"
	"unicode"
)

// 参数结构对象
type Arg struct {
	Name    *ast.Ident //参数名称
	Typ     ast.Expr   //参数类型
	IsStar  bool       //是否是指针
	IsError bool       //是错误
}

// 获取参数ast结构
func (a Arg) Field(scope *ast.Scope, isExport bool) *ast.Field {
	ret := &ast.Field{
		Names: []*ast.Ident{a.chooseName(scope)},
		Type:  a.Typ,
	}
	if isExport { //如果需要导出,首字母大写
		for _, v := range ret.Names {
			v.Name = strings.Title(v.Name)
		}
	}
	return ret
}

// 返回参数结构对象,不包含名称. 主要用于method返回
func (a Arg) Result() *ast.Field {
	return &ast.Field{
		Names: nil,
		Type:  a.Typ,
	}
}

func (a Arg) toDecl() ast.Decl {
	return &ast.GenDecl{
		Doc: nil,
		Tok: token.TYPE,
		Specs: []ast.Spec{
			&ast.TypeSpec{
				Name: a.Name,
				Type: a.Typ,
			},
		},
	}
}

// 获取参数名称
func (a Arg) chooseName(scope *ast.Scope) *ast.Ident {
	if scope != nil {
		if a.Name == nil || scope.Lookup(a.Name.Name) != nil {
			return inventName(a.Typ, scope)
		}
	}
	return a.Name
}

// 根据结构体生成名称,并注入到scope的作用域中
func inventName(t ast.Expr, scope *ast.Scope) *ast.Ident {
	n := baseName(t)
	for try := 0; ; try++ {
		nstr := pickName(n, try) //增加try计数包装名称防止在scope中重复
		obj := ast.NewObj(ast.Var, nstr)
		if alt := scope.Insert(obj); alt == nil { //注入成功
			return ast.NewIdent(nstr)
		}
	}
}

// 根据t对应的类型生成相应的字符串
func baseName(t ast.Expr) string {
	switch tt := t.(type) {
	default:
		panic(fmt.Sprintf("don't know how to choose a base name for %T (%[1]v)", tt))
	case *ast.MapType:
		return "map"
	case *ast.InterfaceType:
		return "inf"
	case *ast.ArrayType:
		return "slice"
	case *ast.Ident:
		return tt.Name
	case *ast.SelectorExpr:
		return tt.Sel.Name
	case *ast.StarExpr:
		return baseName(tt.X)
	}
}

// 包装对象名称
func pickName(base string, idx int) string {
	if idx == 0 {
		switch base {
		default:
			return strings.Split(base, "")[0]
		case "Context":
			return "ctx"
		case "error":
			return "err"
		}
	}
	return fmt.Sprintf("%s%d", base, idx)
}

// 首字母消息
func nameFirstWordToLower(str string) string {
	runs := []rune(str)
	if len(runs) < 1 {
		return ""
	}
	runs[0] = unicode.ToLower(runs[0])
	return string(runs)
}

// 首字母大写
func nameFirstWordToUpper(str string) string {
	runs := []rune(str)
	if len(runs) < 1 {
		return ""
	}
	runs[0] = unicode.ToUpper(runs[0])
	return string(runs)
}
