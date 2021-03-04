package rpcx_code_generate

import (
	"fmt"
	"go/ast"
)

type (
	// 解析对象
	parseVisitor struct {
		fileName    string         //文件名称
		packageName string         //包名称
		src         *SourceContext //源文件解析内容存储
	}

	// 公共对象解析参考 ast.GenDecl
	genDelVisitor struct {
		comments *ast.CommentGroup
		ps       *parseVisitor
		src      *SourceContext
	}

	// type解析参考 ast.TypeSpec
	typeSpecVisitor struct {
		src     *SourceContext
		gen     *genDelVisitor
		node    *ast.TypeSpec
		structs *StructType
		name    *ast.Ident
	}

	// 接口解析
	interfaceVisitor struct {
		node    *ast.TypeSpec
		ts      *typeSpecVisitor
		methods []*fun
	}

	// 结构体解析
	structVisitor struct {
		node    *ast.TypeSpec
		ts      *typeSpecVisitor
		methods []*fun
	}

	// 方法解析
	methodVisitor struct {
		depth           int
		node            *ast.TypeSpec
		list            *[]*fun
		name            *ast.Ident
		params, results *[]Arg
		isMethod        bool
		comments        *ast.CommentGroup
	}

	// 参数解析
	argVisitor struct {
		node  *ast.TypeSpec
		parts []ast.Expr
		list  *[]Arg
	}
)

func (v *parseVisitor) Visit(n ast.Node) ast.Visitor {
	switch rn := n.(type) {
	default:
		return v
	case *ast.GenDecl:
		return &genDelVisitor{comments: rn.Doc, ps: v, src: v.src}
	case *ast.File:
		v.packageName = rn.Name.Name
		v.src.Pkg = rn.Name
		return v
	case *ast.ImportSpec:
		v.src.Imports = append(v.src.Imports, rn)
		return nil
	case *ast.Comment:
		v.src.Common = append(v.src.Common, rn)
		return nil
	case *ast.FuncDecl:
		method := &fun{
			FileName:    v.fileName,
			PackageName: v.packageName,
			Name:        rn.Name,
			Params:      nil,
			Results:     nil,
			Comments:    rn.Doc,
		}
		if rn.Recv != nil {
			var tp []Arg
			for _, v := range rn.Recv.List {
				arg := &argVisitor{list: &tp}
				ast.Walk(arg, v)
			}
			if len(tp) > 0 {
				method.Recv = &tp[0]
			}
		}
		if rn.Type != nil {
			if rn.Type.Params != nil {
				method.Params = []Arg{}
				for _, v1 := range rn.Type.Params.List {
					arg := &argVisitor{list: &method.Params}
					ast.Walk(arg, v1)
				}
			}
			if rn.Type.Results != nil {
				method.Results = []Arg{}
				for _, v1 := range rn.Type.Results.List {
					arg := &argVisitor{list: &method.Results}
					ast.Walk(arg, v1)
				}
			}
		}
		method.initData()
		structName := method.StructName
		if structName != "" { //如果是结构体相关的函数归类到结构体下面
			if v.src.StructFunc == nil {
				v.src.StructFunc = map[string][]*fun{}
			}
			v.src.StructFunc[structName] = append(v.src.StructFunc[structName], method)
		} else {
			v.src.Funcs = append(v.src.Funcs, method)
		}
		return nil
	case *ast.TypeSpec:
		return &typeSpecVisitor{src: v.src, node: rn}
	}
}

func (v *genDelVisitor) Visit(n ast.Node) ast.Visitor {
	switch rn := n.(type) {
	default:
		return v
	case *ast.File:
		v.src.Pkg = rn.Name
		return v
	case *ast.ImportSpec:
		v.src.Imports = append(v.src.Imports, rn)
		return nil
	case *ast.Comment:
		v.src.Common = append(v.src.Common, rn)
		return nil
	case *ast.TypeSpec:
		return &typeSpecVisitor{src: v.src, node: rn, gen: v}
	}
}

func (v *typeSpecVisitor) Visit(n ast.Node) ast.Visitor {
	switch rn := n.(type) {
	default:
		return v
	case *ast.Ident:
		if v.name == nil {
			v.name = rn
		}
		return v
	case *ast.StructType:
		return &structVisitor{ts: v, methods: []*fun{}}
	case *ast.InterfaceType:
		return &interfaceVisitor{ts: v, methods: []*fun{}}
	case nil:
		switch v.node.Type.(type) {
		case *ast.StructType:
			if v.structs != nil {
				if v.gen != nil && v.gen.ps != nil {
					v.structs.FileName = v.gen.ps.fileName
				}
				v.src.Types = append(v.src.Types, v.structs)
			}
		}
		return nil
	}
}

func (v *interfaceVisitor) Visit(n ast.Node) ast.Visitor {
	switch n.(type) {
	default:
		return v
	case *ast.Field:
		return &methodVisitor{list: &v.methods}
	case nil:
		return nil
	}
}

func (v *structVisitor) Visit(n ast.Node) ast.Visitor {
	switch n.(type) {
	default:
		return v
	case *ast.Field:
		return &methodVisitor{list: &v.methods}
	case nil:
		v.ts.structs = &StructType{
			Methods: v.methods,
		}
		v.ts.structs.Name = v.ts.name
		if v.ts.gen != nil {
			v.ts.structs.Comments = v.ts.gen.comments
		}
	}
	return nil
}

func (v *methodVisitor) Visit(n ast.Node) ast.Visitor {
	switch rn := n.(type) {
	default:
		v.depth++
		return v
	case *ast.CommentGroup:
		v.comments = rn
		v.depth++
		return v
	case *ast.Ident:
		if rn.IsExported() {
			v.name = rn
		}
		v.depth++
		return v
	case *ast.FuncType:
		v.depth++
		v.isMethod = true
		return v
	case *ast.FieldList:
		if v.params == nil {
			v.params = &[]Arg{}
			return &argVisitor{list: v.params}
		}
		if v.results == nil {
			v.results = &[]Arg{}
		}
		return &argVisitor{list: v.results}
	case nil:
		v.depth--
		if v.depth == 0 && v.isMethod && v.name != nil {
			method := &fun{Name: v.name, Comments: v.comments}
			if v.results != nil {
				method.Results = *v.results
			}
			if v.params != nil {
				method.Params = *v.params
			}
			method.initData()
			*v.list = append(*v.list, method)
		}
		return nil
	}
}

func (v *argVisitor) Visit(n ast.Node) ast.Visitor {
	switch t := n.(type) {
	case *ast.Field: //field对象
		return &argVisitor{list: v.list}
	case *ast.CommentGroup, *ast.BasicLit:
		return nil
	case *ast.Ident: //Expr -> everything, but clarity
		if t.Name != "_" {
			v.parts = append(v.parts, t)
		}
	case ast.Expr:
		v.parts = append(v.parts, t)
	case nil:
		names := v.parts[:len(v.parts)-1]
		tp := v.parts[len(v.parts)-1]
		iserr, isstar := v.parseTpIsError(tp)
		if len(names) == 0 { //如果没有参数增加一个错误类型参数
			*v.list = append(*v.list, Arg{Typ: tp, IsError: iserr, IsStar: isstar})
			return nil
		}
		for _, n := range names { //最后取到的名称,转换成参数结构体
			*v.list = append(*v.list, Arg{
				Name:    n.(*ast.Ident),
				Typ:     tp,
				IsError: iserr,
				IsStar:  isstar,
			})
		}
	}
	return nil
}

//是否是错误类型
func (v *argVisitor) parseTpIsError(tp ast.Expr) (bool, bool) {
	if v, ok := tp.(*ast.SelectorExpr); ok {
		return fmt.Sprintf("%s", v.Sel.Name) == "Error", false
	} else if _, ok := tp.(*ast.StarExpr); ok {
		return false, true
	}
	return false, false
}
