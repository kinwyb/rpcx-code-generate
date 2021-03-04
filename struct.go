package rpcx_code_generate

import (
	"go/ast"
	"go/token"
)

// 结构体
type StructType struct {
	FileName string            //文件名
	Name     *ast.Ident        //结构名称
	Methods  []*fun            //相关函数
	Comments *ast.CommentGroup //注释
}

func (i StructType) StructName() *ast.Ident {
	return i.Name
}

func (i StructType) Reciever(suffix string) *ast.FieldList {
	ret := &ast.FieldList{
		List: []*ast.Field{
			{
				Names: []*ast.Ident{
					ast.NewIdent(nameFirstWordToLower(i.Name.String()[:1])),
				},
				Type: nil,
			},
		},
	}
	ret.List[0].Type = &ast.StarExpr{
		X: &ast.Ident{Name: i.Name.String() + suffix},
	}
	return ret
}

func StructDecl(name *ast.Ident, fields *ast.FieldList, comment ...*ast.CommentGroup) ast.Decl {
	if name == nil { // 没有名称信息直接返回空
		return nil
	}
	ret := &ast.GenDecl{
		Tok: token.TYPE,
		Specs: []ast.Spec{&ast.TypeSpec{
			Name: name,
			Type: &ast.StructType{
				Fields: fields,
			},
		}},
	}
	if len(comment) > 0 {
		ret.Doc = comment[0]
	}
	return ret
}
