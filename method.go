package rpcx_code_generate

import (
	"go/ast"
	"strings"
)

// 函数类型接口
type methodType interface {
	Reciever(suffix string) *ast.FieldList
	StructName() *ast.Ident
}

// 标记结构体是否被导出
//
// 注解方式 @rpcxService 服务名
// 返回 (是否导出,导出服务名)
func structIsExport(c *ast.CommentGroup) (bool, string) {
	if c != nil {
		for _, text := range c.List {
			str := strings.TrimSpace(strings.TrimLeft(text.Text, "//"))
			if strings.HasPrefix(str, "@rpcxService ") {
				serviceNames := strings.Split(str, " ")
				if len(serviceNames) > 1 {
					return true, serviceNames[1]
				}
			}
		}
	}
	return false, ""
}

// 方法导出信息
type funcExportInfo struct {
	isExport    bool   // 是否导出
	serviceName string // 服务名称
	isGoRun     bool   // 是否使用go xxx()方式运行
}

// 函数是否标记成导出
//
// 注解方式 @rpcxMethod 服务名
// 注解方式 @rpcxMethod.go 服务名 使用go标记已go fun()方式运行
func funcIsExport(c *ast.CommentGroup) funcExportInfo {
	ret := funcExportInfo{}
	if c != nil {
		for _, text := range c.List {
			str := strings.TrimSpace(strings.TrimLeft(text.Text, "//"))
			if strings.HasPrefix(str, "@rpcxMethod") {
				serviceNames := strings.Split(str, " ")
				ret.isExport = true
				ret.isGoRun = serviceNames[0] == "@rpcxMethod.go"
				if len(serviceNames) > 1 {
					ret.serviceName = serviceNames[1]
				}
				return ret
			}
		}
	}
	return ret
}

// 函数结构对象
type fun struct {
	FileName        string            //所在文件名称
	PackageName     string            //包名称
	Name            *ast.Ident        //方法名称
	StructName      string            //结构体名称
	Recv            *Arg              //结构体对象
	Params          []Arg             //请求参数
	Results         []Arg             //返回参数
	Comments        *ast.CommentGroup //注释信息
	StructsResolved bool              //标记是否已经解析,防止参数重复生成
	ExportInfo      funcExportInfo    // 导出信息
}

// 解析函数数据
func (m *fun) initData() {
	if m.StructsResolved {
		return
	}
	m.StructsResolved = true
	scope := ast.NewScope(nil)
	for i, p := range m.Params {
		p.Name = p.chooseName(scope)
		m.Params[i] = p
	}
	scope = ast.NewScope(nil)
	for i, r := range m.Results {
		r.Name = r.chooseName(scope)
		m.Results[i] = r
	}
	if m.Recv != nil {
		structName := ""
		if recv, ok := m.Recv.Typ.(*ast.Ident); ok {
			structName = recv.String()
		} else if startRecv, ok := m.Recv.Typ.(*ast.StarExpr); ok {
			if recv, ok := startRecv.X.(*ast.Ident); ok {
				structName = recv.String()
			}
		}
		m.StructName = structName
	}
	m.ExportInfo = funcIsExport(m.Comments)
}

// 生成请求结构体
func (m *fun) RequestStruct(nameToLower bool) ast.Decl {
	if len(m.Params) < 2 { // 没有参数直接返回
		return nil
	}
	reqFields := argListToFieldList(func(a Arg) *ast.Field {
		return a.Field(nil, true)
	}, true, m.Params...)
	return StructDecl(m.RequestStructType(nameToLower, ""), reqFields)
}

func (m *fun) ResponseStruct(nameToLower bool) ast.Decl {
	fields := m.ResponseStructFields()
	if len(fields.List) < 1 {
		return nil
	}
	return StructDecl(m.ResponseStructName(nameToLower, ""), fields)
}

// 生成请求结构体名称,(结构体名称,结构体类型)
func (m *fun) RequestStructType(nameToLower bool, pkgName string) *ast.Ident {
	if len(m.Params) < 1 {
		return nil
	}
	if len(m.Params) == 1 {
		if isContext(m.Params[0].Typ) {
			return nil
		}
		switch v := m.Params[0].Typ.(type) {
		case *ast.StarExpr:
			return v.X.(*ast.Ident)
		case *ast.Ident:
			return v
		default:
			return nil
		}
	}
	name := "RpcxRequest" + m.StructName + nameFirstWordToUpper(m.Name.Name)
	if pkgName != "" {
		pkgName = pkgName + "."
	} else if nameToLower {
		name = nameFirstWordToLower(name)
	}
	return ast.NewIdent(pkgName + name)
}

func (m *fun) RequestStructTypeExpr(nameToLower bool, pkgName string) ast.Expr {
	if len(m.Params) < 1 {
		return nil
	}
	if len(m.Params) == 1 {
		if isContext(m.Params[0].Typ) {
			return nil
		}
		return m.Params[0].Typ
	}
	name := "RpcxRequest" + m.StructName + nameFirstWordToUpper(m.Name.Name)
	if pkgName != "" {
		pkgName = pkgName + "."
	} else if nameToLower {
		name = nameFirstWordToLower(name)
	}
	return ast.NewIdent(pkgName + name)
}

// 生成请求结构体名称,(结构体名称,结构体类型)
func (m *fun) RequestParamName(nameToLower bool, pkgName string) *ast.Ident {
	if len(m.Params) < 1 {
		return nil
	}
	if len(m.Params) == 1 {
		if isContext(m.Params[0].Typ) {
			return nil
		}
		return m.Params[0].Name
	}
	name := "RpcxRequest" + m.StructName + nameFirstWordToUpper(m.Name.Name)
	if pkgName != "" {
		pkgName = pkgName + "."
	} else if nameToLower {
		name = nameFirstWordToLower(name)
	}
	return ast.NewIdent(pkgName + name)
}

// 生成返回结构体名称
func (m *fun) ResponseStructName(nameToLower bool, pkgName string) *ast.Ident {
	if len(m.Results) < 1 {
		return nil
	}
	name := "RpcxResponse" + m.StructName + nameFirstWordToUpper(m.Name.Name)
	if pkgName != "" {
		pkgName = pkgName + "."
	} else if nameToLower {
		name = nameFirstWordToLower(name)
	}
	return ast.NewIdent(pkgName + name)
}

// 获取返回结构体的字段
func (m *fun) ResponseStructFields() *ast.FieldList {
	return argListToFieldList(func(a Arg) *ast.Field {
		return a.Field(nil, true)
	}, false, m.Results...)
}

// 包装请求参数
func (m *fun) WrapRequest(ignoreContext bool, nameToLower bool) ast.Expr {
	var kvs []ast.Expr
	for _, a := range m.Params {
		if ignoreContext && isContext(a.Typ) {
			continue
		}
		kvs = append(kvs, &ast.KeyValueExpr{
			Key:   ast.NewIdent(nameFirstWordToUpper(a.Name.Name)),
			Value: ast.NewIdent(a.Name.Name),
		})
	}
	name := m.RequestStructType(nameToLower, "")
	return &ast.CompositeLit{
		Type: name,
		Elts: kvs,
	}
}

// 是否包含上下文参数
func (m *fun) HasContextParam() (bool, int) {
	for index, v := range m.Params {
		if isContext(v.Typ) {
			return true, index
		}
	}
	return false, 0
}

// 是否返回错误
func (m *fun) HasErrorResult() (bool, int) {
	for index, v := range m.Results {
		if v.IsError {
			return true, index
		}
	}
	return false, 0
}

// Arg结构转换成ast.Field结构
func argListToFieldList(fn func(Arg) *ast.Field, ignoreContext bool, args ...Arg) *ast.FieldList {
	fl := &ast.FieldList{List: []*ast.Field{}}
	for _, a := range args {
		if ignoreContext && isContext(a.Typ) {
			continue
		}
		fl.List = append(fl.List, fn(a))
	}
	return fl
}

// 判断是否是context.Context对象
func isContext(exp ast.Expr) bool {
	if sel, is := exp.(*ast.SelectorExpr); is && sel.Sel.Name == "Context" { //判断结构体是Context
		if id, is := sel.X.(*ast.Ident); is && id.Name == "context" { //判断包是context
			return true
		}
	}
	return false
}
