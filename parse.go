package rpcx_code_generate

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"golang.org/x/tools/go/ssa/interp/testdata/src/errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var (
	rpcxServiceStructSuffix = "RpcxService"
	rpcxClientStructSuffix  = "RpcxClient"
)

// 解析目录
type parseDir struct {
	dirPath                 string   //目录路径
	outpath                 string   //输出路径
	pkgName                 string   //包名
	globalFuns              []string //导出的公共函数
	files                   []*parseFile
	exportStructServiceName map[string]string //导出函数的结构体
	serviceGroupFile        *ast.File
	clientGroupFile         *ast.File
}

func NewParseDir(sourcePath, outPath string) *parseDir {
	return &parseDir{
		dirPath: sourcePath,
		outpath: outPath,
	}
}

// 生成文件
func (p *parseDir) Generate() error {
	if p.dirPath == "" {
		return errors.New("路径空")
	}
	p.dirPath, _ = filepath.Abs(p.dirPath)
	fileInfo, err := os.Stat(p.dirPath)
	if err != nil {
		return err
	}
	if fileInfo.IsDir() { //如果是文件夹,遍历解析所有文件
		err = filepath.Walk(p.dirPath, func(path string, info os.FileInfo, err error) error {
			if info.IsDir() && path != p.dirPath {
				return filepath.SkipDir
			}
			return p.parseFile(path)
		})
	} else {
		err = p.parseFile(p.dirPath)
	}
	if err != nil {
		return nil
	}
	p.exportServiceGroupStruct()
	p.exportClientGroupStruct()
	return nil
}

func (p *parseDir) parseFile(path string) error {
	if strings.HasSuffix(path, rpcxServiceStructSuffix+"_appToolGen.go") {
		// 移除自动生成文件
		os.Remove(path)
		return nil
	} else if ext := filepath.Ext(path); ext != ".go" {
		// 不是go文件忽略
		return nil
	} else if strings.HasSuffix(path, "_test.go") {
		// 忽略test文件
		return nil
	}
	println("解析文件:" + path)
	pfile := &parseFile{filePath: path}
	err := pfile.Generate()
	if err == nil {
		if p.exportStructServiceName == nil {
			p.exportStructServiceName = map[string]string{}
		}
		p.globalFuns = append(p.globalFuns, pfile.globalFuns...)
		for k, v := range pfile.exportStructServiceName {
			p.exportStructServiceName[k] = v
		}
		if len(p.exportStructServiceName) > 0 || len(p.globalFuns) > 0 {
			p.files = append(p.files, pfile)
		}
	}
	return err
}

// rpcx导出组结构体生成
func (p *parseDir) exportServiceGroupStruct() {
	if len(p.files) < 1 {
		return
	} else if p.pkgName == "" {
		p.pkgName = p.files[0].packageName
	}
	srcFile := &ast.File{
		Name: ast.NewIdent(p.pkgName + "Sev"),
		Decls: []ast.Decl{
			importFor(&ast.ImportSpec{
				Doc:  nil,
				Name: ast.NewIdent("rpcxServer"),
				Path: &ast.BasicLit{
					ValuePos: 0,
					Kind:     token.STRING,
					Value:    "\"github.com/smallnest/rpcx/server\"",
				},
				Comment: nil,
				EndPos:  0,
			}),
		},
	}
	name := rpcxServiceStructSuffix + "Group"
	name = strings.ToTitle(name[:1]) + name[1:]
	typeFieldList := &ast.FieldList{}
	// 生成服务接口
	// 生成服务集合结构体
	ds := StructDecl(ast.NewIdent(name), typeFieldList, nil)
	srcFile.Decls = append(srcFile.Decls, ds)
	recvName := nameFirstWordToLower(name[:1])
	initFunc := funcDecl()
	initFunc.Recv = fieldList(field(ast.NewIdent(recvName), nil))
	initFunc.Recv.List[0].Type = &ast.StarExpr{
		X: &ast.Ident{Name: name},
	}
	initFunc.Name = ast.NewIdent("Services")
	initFunc.Type.Params = fieldList(
		field(ast.NewIdent("rpcxSev"), &ast.StarExpr{X: ast.NewIdent("rpcxServer.Server")}),
		field(ast.NewIdent("meta"), ast.NewIdent("string")),
	)
	initFunc.Type.Results = nil
	initFunc.Body.List = []ast.Stmt{
		&ast.DeclStmt{
			Decl: &ast.GenDecl{
				Tok: token.VAR,
				Specs: []ast.Spec{
					&ast.ValueSpec{
						Names: []*ast.Ident{
							ast.NewIdent("err"),
						},
						Type: ast.NewIdent("error"),
					},
				},
			},
		},
	}
	// 遍历结构体
	for structName, serviceName := range p.exportStructServiceName {
		if serviceName == "rpcxFuncs" {
			serviceName = "Funcs"
		}
		typeFieldList.List = append( //rpcxGroup 结构体增加字段
			typeFieldList.List,
			field(ast.NewIdent(nameFirstWordToLower(serviceName)),
				&ast.StarExpr{
					X: ast.NewIdent(p.pkgName + "." + structName),
				},
			),
		)
		// 设置set函数
		setFunc := funcDecl()
		setFunc.Recv = fieldList(field(ast.NewIdent(recvName), nil))
		setFunc.Recv.List[0].Type = &ast.StarExpr{
			X: &ast.Ident{Name: name},
		}
		setFunc.Name = ast.NewIdent("Set" + structName)
		setFunc.Type.Params = fieldList(field(ast.NewIdent(serviceName),
			&ast.StarExpr{
				X: ast.NewIdent(p.pkgName + "." + structName),
			}))
		setFunc.Type.Results = nil
		setFunc.Body.List = []ast.Stmt{
			&ast.AssignStmt{
				Lhs: []ast.Expr{
					ast.NewIdent(recvName + "." + nameFirstWordToLower(serviceName)),
				},
				Tok: token.ASSIGN,
				Rhs: []ast.Expr{
					ast.NewIdent(serviceName),
				},
			},
		}
		srcFile.Decls = append(srcFile.Decls, setFunc)
		//set函数结束
		// 判断结构体是否初始化，没初始化panic
		ifStmt := &ast.IfStmt{
			Cond: &ast.BasicLit{
				Kind:  token.TYPE,
				Value: recvName + "." + nameFirstWordToLower(serviceName) + " == nil",
			},
			Body: &ast.BlockStmt{
				List: []ast.Stmt{
					&ast.ExprStmt{
						X: &ast.CallExpr{
							Fun: ast.NewIdent("panic"),
							Args: []ast.Expr{
								ast.NewIdent("\"" + rpcxServiceStructSuffix + "Group[" + structName + "]尚未初始化\""),
							},
						},
					},
				},
			},
		}
		initFunc.Body.List = append(initFunc.Body.List, ifStmt)
		// 初始化结构体
		rpcxStructName := structName + rpcxServiceStructSuffix
		rpcxStructName = nameFirstWordToLower(rpcxStructName)
		structObject := &ast.UnaryExpr{
			Op: token.AND,
			X: &ast.CompositeLit{
				Type: ast.NewIdent(rpcxStructName),
			},
		}
		if o, ok := structObject.X.(*ast.CompositeLit); ok {
			o.Elts = []ast.Expr{
				&ast.KeyValueExpr{
					Key:   ast.NewIdent("Serv"),
					Value: ast.NewIdent(recvName + "." + nameFirstWordToLower(serviceName)),
				},
			}
		}
		stmt := []ast.Stmt{
			&ast.AssignStmt{
				Lhs: []ast.Expr{
					ast.NewIdent(strings.ToLower(serviceName)),
				},
				Tok: token.DEFINE,
				Rhs: []ast.Expr{
					structObject,
				},
			},
			&ast.AssignStmt{
				Lhs: []ast.Expr{
					ast.NewIdent("err"),
				},
				Tok: token.ASSIGN,
				Rhs: []ast.Expr{
					&ast.CallExpr{
						Fun: ast.NewIdent("rpcxSev.RegisterName"),
						Args: []ast.Expr{
							ast.NewIdent("\"" + serviceName + "\""),
							ast.NewIdent(strings.ToLower(serviceName)),
							ast.NewIdent("meta"),
						},
					},
				},
			},
			&ast.IfStmt{
				Cond: &ast.BasicLit{
					Kind:  token.TYPE,
					Value: "err != nil",
				},
				Body: &ast.BlockStmt{
					List: []ast.Stmt{
						&ast.ExprStmt{
							X: &ast.CallExpr{
								Fun: ast.NewIdent("panic"),
								Args: []ast.Expr{
									ast.NewIdent("\"rpcx服务[" + serviceName + "]注册错误:\"+err.Error()"),
								},
							},
						},
					},
				},
			},
		}
		initFunc.Body.List = append(initFunc.Body.List, stmt...)
	}
	for _, v := range p.globalFuns {
		stmt := []ast.Stmt{
			&ast.AssignStmt{
				Lhs: []ast.Expr{
					ast.NewIdent("err"),
				},
				Tok: token.ASSIGN,
				Rhs: []ast.Expr{
					&ast.CallExpr{
						Fun: ast.NewIdent("rpcxSev.RegisterFunctionName"),
						Args: []ast.Expr{
							ast.NewIdent("\"globalFun\""),
							ast.NewIdent("\"" + v + "\""),
							ast.NewIdent(nameFirstWordToUpper(v)),
							ast.NewIdent("meta"),
						},
					},
				},
			},
			&ast.IfStmt{
				Cond: &ast.BasicLit{
					Kind:  token.TYPE,
					Value: "err != nil",
				},
				Body: &ast.BlockStmt{
					List: []ast.Stmt{
						&ast.ExprStmt{
							X: &ast.CallExpr{
								Fun: ast.NewIdent("panic"),
								Args: []ast.Expr{
									ast.NewIdent("\"rpcx函数[" + v + "]注册错误:\"+err.Error()"),
								},
							},
						},
					},
				},
			},
		}
		initFunc.Body.List = append(initFunc.Body.List, stmt...)
	}
	srcFile.Decls = append(srcFile.Decls, initFunc)
	p.serviceGroupFile = srcFile
}

func (p *parseDir) exportClientGroupStruct() {
	if len(p.files) < 1 {
		return
	} else if p.pkgName == "" {
		p.pkgName = p.files[0].packageName
	}
	srcFile := &ast.File{
		Name: ast.NewIdent(p.pkgName + "Clt"),
		Decls: []ast.Decl{
			importFor(&ast.ImportSpec{
				Path: &ast.BasicLit{
					ValuePos: 0,
					Kind:     token.STRING,
					Value:    "\"github.com/smallnest/rpcx/client\"",
				},
			}),
		},
	}
	name := rpcxClientStructSuffix + "Group"
	name = nameFirstWordToUpper(name)
	srcFile.Decls = append(srcFile.Decls, &ast.GenDecl{
		Doc: nil,
		Tok: token.VAR,
		Specs: []ast.Spec{
			&ast.ValueSpec{
				Doc:     nil,
				Names:   []*ast.Ident{ast.NewIdent("gClient")},
				Type:    ast.NewIdent("client.XClient"),
				Values:  nil,
				Comment: nil,
			},
			&ast.ValueSpec{
				Doc:   nil,
				Names: []*ast.Ident{ast.NewIdent("lg")},
				Values: []ast.Expr{
					&ast.CallExpr{
						Fun: ast.NewIdent("logrus.New"),
					},
				},
				Comment: nil,
			},
		},
	})
	initFunc := funcDecl()
	initFunc.Recv = nil
	initFunc.Name = ast.NewIdent("RpcxClients")
	initFunc.Type.Params = fieldList(
		field(ast.NewIdent("discovery"), ast.NewIdent("client.ServiceDiscovery")),
		field(ast.NewIdent("failMode"), ast.NewIdent("client.FailMode")),
		field(ast.NewIdent("selectMode"), ast.NewIdent("client.SelectMode")),
		field(ast.NewIdent("option"), ast.NewIdent("client.Option")),
	)
	initFunc.Type.Results = nil
	// 遍历结构体
	for structName, serviceName := range p.exportStructServiceName {
		srcFile.Decls = append(srcFile.Decls, &ast.GenDecl{
			Doc: nil,
			Tok: token.VAR,
			Specs: []ast.Spec{
				&ast.ValueSpec{
					Doc:   nil,
					Names: []*ast.Ident{ast.NewIdent(nameFirstWordToUpper(serviceName))},
					Type: &ast.StarExpr{
						X: ast.NewIdent(nameFirstWordToLower(structName) + rpcxClientStructSuffix),
					},
					Values:  nil,
					Comment: nil,
				},
			},
		})
		var stmt ast.Stmt
		stmt = &ast.IfStmt{
			Cond: &ast.BasicLit{
				Kind:  token.TYPE,
				Value: nameFirstWordToUpper(serviceName) + " == nil",
			},
			Body: &ast.BlockStmt{
				List: []ast.Stmt{
					&ast.AssignStmt{
						Lhs: []ast.Expr{
							ast.NewIdent(nameFirstWordToUpper(serviceName)),
						},
						Tok: token.ASSIGN,
						Rhs: []ast.Expr{
							&ast.UnaryExpr{
								Op: token.AND,
								X: &ast.CompositeLit{
									Type: ast.NewIdent(nameFirstWordToLower(structName) + rpcxClientStructSuffix),
									Elts: []ast.Expr{
										&ast.KeyValueExpr{
											Key: ast.NewIdent("client"),
											Value: &ast.CallExpr{
												Fun: ast.NewIdent("client.NewXClient"),
												Args: []ast.Expr{
													ast.NewIdent("\"" + serviceName + "\""),
													ast.NewIdent("failMode"),
													ast.NewIdent("selectMode"),
													ast.NewIdent("discovery"),
													ast.NewIdent("option"),
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}
		initFunc.Body.List = append(initFunc.Body.List, stmt)
	}
	if len(p.globalFuns) > 0 {
		stmt := &ast.IfStmt{
			Cond: &ast.BasicLit{
				Kind:  token.TYPE,
				Value: " gClient == nil",
			},
			Body: &ast.BlockStmt{
				List: []ast.Stmt{
					&ast.AssignStmt{
						Lhs: []ast.Expr{
							ast.NewIdent("gClient"),
						},
						Tok: token.ASSIGN,
						Rhs: []ast.Expr{
							&ast.CallExpr{
								Fun: ast.NewIdent("client.NewXClient"),
								Args: []ast.Expr{
									ast.NewIdent("\"globalFun\""),
									ast.NewIdent("failMode"),
									ast.NewIdent("selectMode"),
									ast.NewIdent("discovery"),
									ast.NewIdent("option"),
								},
							},
						},
					},
				},
			},
		}
		initFunc.Body.List = append(initFunc.Body.List, stmt)
	}
	srcFile.Decls = append(srcFile.Decls, initFunc)
	// close函数
	closeFun := funcDecl()
	closeFun.Recv = nil
	closeFun.Name = ast.NewIdent("RpcxClose")
	closeFun.Type.Params = nil
	for _, serviceName := range p.exportStructServiceName {
		var stmt ast.Stmt
		stmt = &ast.IfStmt{
			Cond: &ast.BasicLit{
				Kind:  token.TYPE,
				Value: nameFirstWordToUpper(serviceName) + " != nil",
			},
			Body: &ast.BlockStmt{
				List: []ast.Stmt{
					&ast.ExprStmt{
						X: &ast.CallExpr{
							Fun: ast.NewIdent(nameFirstWordToUpper(serviceName) + ".client.Close"),
						},
					},
					&ast.AssignStmt{
						Lhs: []ast.Expr{
							ast.NewIdent(nameFirstWordToUpper(serviceName)),
						},
						Tok: token.ASSIGN,
						Rhs: []ast.Expr{
							ast.NewIdent("nil"),
						},
					},
				},
			},
		}
		closeFun.Body.List = append(closeFun.Body.List, stmt)
	}
	if len(p.globalFuns) > 0 {
		stmt := &ast.IfStmt{
			Cond: &ast.BasicLit{
				Kind:  token.TYPE,
				Value: " gClient != nil",
			},
			Body: &ast.BlockStmt{
				List: []ast.Stmt{
					&ast.ExprStmt{
						X: &ast.CallExpr{
							Fun:  ast.NewIdent("gClient.Close"),
							Args: nil,
						},
					},
					&ast.AssignStmt{
						Lhs: []ast.Expr{
							ast.NewIdent("gClient"),
						},
						Tok: token.ASSIGN,
						Rhs: []ast.Expr{
							ast.NewIdent("nil"),
						},
					},
				},
			},
		}
		closeFun.Body.List = append(closeFun.Body.List, stmt)
	}
	srcFile.Decls = append(srcFile.Decls, closeFun)
	// setLog函数
	logFun := funcDecl()
	logFun.Recv = nil
	logFun.Name = ast.NewIdent("SetLogger")
	logFun.Type.Params = fieldList(field(ast.NewIdent("l"), &ast.StarExpr{
		X: ast.NewIdent("logrus.Logger"),
	}))
	logFun.Body.List = []ast.Stmt{
		&ast.AssignStmt{
			Lhs: []ast.Expr{
				ast.NewIdent("lg"),
			},
			Tok: token.ASSIGN,
			Rhs: []ast.Expr{
				ast.NewIdent("l"),
			},
		},
	}
	srcFile.Decls = append(srcFile.Decls, logFun)
	p.clientGroupFile = srcFile
}

func (p *parseDir) OutputToFile() error {
	for _, v := range p.files {
		err := v.OutputToFile(p.outpath)
		if err != nil {
			return err
		}
	}
	if p.serviceGroupFile != nil {
		filedata, err := outputFileBytes(p.serviceGroupFile)
		if err != nil {
			return err
		}
		outDir := p.outpath + "/" + p.pkgName + "Sev"
		_, err = os.Stat(outDir)
		if os.IsNotExist(err) {
			os.MkdirAll(outDir, 0777)
		}
		name := "rpcxGroup_appToolGen.go"
		path := filepath.Join(outDir, name)
		err = ioutil.WriteFile(path, filedata, os.ModePerm)
		if err != nil {
			fmt.Printf("[%s]文件保存错误:%s\n", name, err.Error())
			return err
		} else {
			fmt.Printf("[%s]文件保存成功\n", name)
		}
	}
	if p.clientGroupFile != nil {
		filedata, err := outputFileBytes(p.clientGroupFile)
		if err != nil {
			return err
		}
		outDir := p.outpath + "/" + p.pkgName + "Clt"
		_, err = os.Stat(outDir)
		if os.IsNotExist(err) {
			os.MkdirAll(outDir, 0777)
		}
		name := "rpcxGroup_appToolGen.go"
		path := filepath.Join(outDir, name)
		err = ioutil.WriteFile(path, filedata, os.ModePerm)
		if err != nil {
			fmt.Printf("[%s]文件保存错误:%s\n", name, err.Error())
			return err
		} else {
			fmt.Printf("[%s]文件保存成功\n", name)
		}
	}
	return nil
}

// 解析文件
type parseFile struct {
	sctx                    SourceContext
	packageName             string            //包名
	filePath                string            //文件名
	fileName                string            //文件名
	globalFuns              []string          //公共函数名
	exportStructServiceName map[string]string //导出函数的结构体
	serviceFile             *ast.File         //生成的服务端文件
	clientFile              *ast.File         //生成客户端文件
	structFile              *ast.File         //生成结构体文件
}

// 生成文件
func (p *parseFile) Generate() error {
	filedata, err := os.ReadFile(p.filePath)
	if err != nil {
		return err
	}
	_, p.fileName = filepath.Split(p.filePath)
	ext := filepath.Ext(p.fileName)
	p.fileName = strings.TrimSuffix(p.fileName, ext)
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", bytes.NewReader(filedata), parser.DeclarationErrors|parser.ParseComments)
	if err != nil {
		return err
	}
	visitor := &parseVisitor{src: &p.sctx, fileName: p.fileName}
	ast.Walk(visitor, f)
	return p.transformAST()
}

func (p *parseFile) Output() (map[string][]byte, error) {
	var err error
	ret := map[string][]byte{}
	if p.serviceFile != nil {
		ret["sev"], err = outputFileBytes(p.serviceFile)
		if err != nil {
			return nil, err
		}
	}
	if p.clientFile != nil {
		ret["clt"], err = outputFileBytes(p.clientFile)
		if err != nil {
			return nil, err
		}
	}
	if p.structFile != nil {
		ret["obj"], err = outputFileBytes(p.structFile)
		if err != nil {
			return nil, err
		}
	}
	return ret, nil
}

func (p *parseFile) OutputToFile(outpath string) error {
	fileMap, err := p.Output()
	if err != nil {
		return err
	}
	for fileName, filedata := range fileMap {
		outDir := outpath + "/" + p.packageName + nameFirstWordToUpper(fileName)
		_, err := os.Stat(outDir)
		if os.IsNotExist(err) {
			os.MkdirAll(outDir, 0777)
		}
		name := p.fileName + "_appToolGen.go"
		path := filepath.Join(outDir, name)
		err = ioutil.WriteFile(path, filedata, os.ModePerm)
		if err != nil {
			fmt.Printf("[%s]文件保存错误:%s\n", name, err.Error())
			return err
		} else {
			fmt.Printf("[%s]文件保存成功\n", name)
		}
	}
	return nil
}

// 导出文件到bytes
func outputFileBytes(file *ast.File) ([]byte, error) {
	filedata, err := formatNode("", file)
	if err != nil {
		return nil, err
	}
	return append([]byte("// 这是通过appTool自动生成的rpcx代码，请勿修改\n"), filedata...), nil
}

// 创建导出文件
func (p *parseFile) createFile(packName string) *ast.File {
	astFile := &ast.File{
		Name:  ast.NewIdent(p.packageName + nameFirstWordToUpper(packName)),
		Decls: []ast.Decl{},
	}
	astFile.Decls = append(astFile.Decls, p.sctx.ImportDecls()...) //import
	astFile.Imports = append(astFile.Imports, &ast.ImportSpec{
		Doc:  nil,
		Name: nil,
		Path: &ast.BasicLit{
			ValuePos: 0,
			Kind:     token.STRING,
			Value:    "bjmes/src/appTool",
		},
		Comment: nil,
		EndPos:  0,
	})
	return astFile
}

func (p *parseFile) transformAST() error {
	p.packageName = strings.TrimSuffix(p.sctx.Pkg.Name, "_test")
	var exportStruct []*StructType                  //所有导出结构体
	p.exportStructServiceName = map[string]string{} //结构体和服务名称对应map
	//遍历所有结构体，找出暴露出服务的结构体
	for _, v := range p.sctx.Types {
		hasService, serviceName := structIsExport(v.Comments)
		if hasService {
			methods := p.sctx.StructFunc[v.Name.String()]
			for _, method := range methods {
				if method.ExportInfo.isExport {
					v.Methods = append(v.Methods, method)
				}
			}
			if len(v.Methods) > 0 {
				exportStruct = append(exportStruct, v)
				p.exportStructServiceName[v.Name.String()] = serviceName
			}
		}
	}
	// 检测是否有公共函数导出类型
	var exportFuncs []*fun //所有导出的函数
	for _, v1 := range p.sctx.StructFunc {
		for _, v := range v1 {
			if v.ExportInfo.isExport {
				exportFuncs = append(exportFuncs, v)
			}
		}
	}
	for _, v := range p.sctx.Funcs {
		if v.ExportInfo.isExport {
			exportFuncs = append(exportFuncs, v)
		}
	}
	// 没有需要导出的数据
	if len(exportStruct) < 1 && len(exportFuncs) < 1 {
		return nil
	}
	p.serviceFile = p.createFile("sev")
	p.clientFile = p.createFile("clt")
	p.structFile = p.createFile("obj")
	// 导出结构体
	if len(exportStruct) > 0 {
		for _, v := range exportStruct {
			p.addClientStruct(v.Name.Name)
			p.addServiceStruct(v.Name.Name)
			p.exportServicePublicFunc(v, p.exportStructServiceName[v.Name.String()])
			p.exportClientPublicFunc(v, p.exportStructServiceName[v.Name.String()])
		}
	}
	// 导出函数
	for _, v := range exportFuncs {
		//生成请求结构
		request := v.RequestStruct(false)
		if request != nil {
			p.structFile.Decls = append(p.structFile.Decls, request)
		}
		//生成返回结果结构
		result := v.ResponseStruct(false)
		if result != nil {
			p.structFile.Decls = append(p.structFile.Decls, result)
		}
		var ifc methodType
		if v.StructName != "" {
			ifc = StructType{
				Name: ast.NewIdent(v.StructName),
			}
		} else {
			p.globalFuns = append(p.globalFuns, v.Name.Name)
		}
		p.addServiceMethod(ifc, v)
		p.addClientMethod(ifc, v)
	}
	return nil
}

// 导出服务公共方法名
func (p *parseFile) exportServicePublicFunc(ifc methodType, rpcxService string) {
	rpcxServiceNameFun := funcDecl()
	rpcxServiceNameFun.Recv = ifc.Reciever(rpcxServiceStructSuffix)
	if v, ok := rpcxServiceNameFun.Recv.List[0].Type.(*ast.StarExpr); ok {
		if vn := v.X.(*ast.Ident); ok {
			v.X = ast.NewIdent(nameFirstWordToLower(vn.Name))
		}
	}
	rpcxServiceNameFun.Name = ast.NewIdent("ServicePath")
	rpcxServiceNameFun.Type.Params = nil
	rpcxServiceNameFun.Type.Results = fieldList(&ast.Field{
		Names: nil,
		Type:  ast.NewIdent("string"),
	})
	rpcxServiceNameFun.Body.List = []ast.Stmt{
		&ast.ReturnStmt{
			Results: []ast.Expr{
				ast.NewIdent("\"" + rpcxService + "\""),
			},
		},
	}
	p.serviceFile.Decls = append(p.serviceFile.Decls, rpcxServiceNameFun)
}

// 生成客户端结构体
func (p *parseFile) addClientStruct(structName string) {
	name := structName + rpcxClientStructSuffix
	name = nameFirstWordToLower(name)
	ds := StructDecl(ast.NewIdent(name), &ast.FieldList{
		List: []*ast.Field{{
			Names: []*ast.Ident{ast.NewIdent("client")},
			Type:  ast.NewIdent("client.XClient"),
		}},
	})
	p.clientFile.Decls = append(p.clientFile.Decls, ds)
}

// 生成服务端结构体
func (p *parseFile) addServiceStruct(structName string) {
	name := structName + rpcxServiceStructSuffix
	name = nameFirstWordToLower(name)
	fieldList := &ast.FieldList{
		List: []*ast.Field{{
			Names: []*ast.Ident{ast.NewIdent("Serv")},
			Type: &ast.StarExpr{
				X: ast.NewIdent(p.packageName + "." + structName),
			},
		}},
	}
	ds := StructDecl(ast.NewIdent(name), fieldList)
	p.serviceFile.Decls = append(p.serviceFile.Decls, ds)
}

//生成服务端函数代码
func (p *parseFile) addServiceMethod(ifc methodType, m *fun) {
	notImpl := funcDecl()
	notImpl.Name = m.Name
	if ifc != nil {
		notImpl.Recv = ifc.Reciever(rpcxServiceStructSuffix)
		if v, ok := notImpl.Recv.List[0].Type.(*ast.StarExpr); ok {
			if vn := v.X.(*ast.Ident); ok {
				v.X = ast.NewIdent(nameFirstWordToLower(vn.Name))
			}
		}
	} else {
		notImpl.Recv = nil
	}
	//生成请求参数
	parms := &ast.FieldList{}
	resultTp := m.ResponseStructName(false, p.packageName+"Obj")
	if resultTp == nil {
		resultTp = ast.NewIdent("string")
	}
	var paramTp ast.Expr
	// 这里有一个nil值的坑：当一个interface的type和value都是nil的时候，这个interface才等于nil
	// 而这里调用m.RequestStructType返回的是一个*ast.Ident,当转换为ast.Expr这个interface时
	// ast.Expr的type就不是nil了，这就导致如果m.RequestStructType返回直接赋值给reqStruct，
	// reqStruct == nil 这个条件是不成立的。所以这里引入一个param来判断nil值
	param := m.RequestStructType(false, p.packageName+"Obj")
	if param == nil {
		paramTp = ast.NewIdent("string")
	} else if strings.HasPrefix(param.Name, "RpcxRequest") {
		paramTp = &ast.StarExpr{
			X: param,
		}
	} else {
		paramTp = m.RequestStructTypeExpr(false, p.packageName+"Obj")
	}
	parms.List = []*ast.Field{{
		Names: []*ast.Ident{ast.NewIdent("reqCtx")},
		Type:  sel(ast.NewIdent("context"), ast.NewIdent("Context")),
	}, {
		Names: []*ast.Ident{ast.NewIdent("arg")},
		Type:  paramTp,
	}, {
		Names: []*ast.Ident{ast.NewIdent("resp")},
		Type: &ast.StarExpr{
			X: resultTp,
		},
	}}
	notImpl.Type.Params = parms
	notImpl.Type.Results = &ast.FieldList{
		List: []*ast.Field{{
			Type: ast.NewIdent("error"),
		}},
	}
	ret := &ast.CallExpr{}
	if ifc == nil {
		ret.Fun = ast.NewIdent(m.PackageName + "." + m.Name.Name)
	} else {
		ret.Fun = ast.NewIdent(notImpl.Recv.List[0].Names[0].Name + ".Serv." + m.Name.Name)
	}
	var args []ast.Expr
	for _, v := range m.Params {
		if isContext(v.Typ) {
			args = append(args, ast.NewIdent("reqCtx"))
		} else {
			if len(m.Params) == 1 {
				args = append(args, &ast.BasicLit{
					Kind:  token.TYPE,
					Value: "arg",
				})
			} else {
				args = append(args, &ast.BasicLit{
					Kind:  token.TYPE,
					Value: "arg." + strings.Title(v.Name.Name),
				})
			}
		}
	}
	ret.Args = args
	// 赋值返回结果
	responses := m.ResponseStructFields()
	var r ast.Stmt
	if len(responses.List) > 0 {
		assign := &ast.AssignStmt{
			Tok: token.ASSIGN,
			Rhs: []ast.Expr{ret},
		}
		for _, v := range responses.List {
			if len(v.Names) > 0 {
				assign.Lhs = append(assign.Lhs, ast.NewIdent("resp."+nameFirstWordToUpper(v.Names[0].Name)))
			}
		}
		r = assign
	} else if m.ExportInfo.isGoRun {
		r = &ast.GoStmt{Call: ret}
	} else {
		r = &ast.ExprStmt{X: ret}
	}
	notImpl.Body.List = append(notImpl.Body.List, r)
	// 处理分页函数
	//if ok, index := m.HasPageObjectParam(); ok {
	//	pgArg := m.Params[index]
	//	stmt := &ast.AssignStmt{
	//		Lhs: []ast.Expr{
	//			ast.NewIdent("resp." + strings.Title(pgArg.Name.Name)),
	//		},
	//		Tok: token.ASSIGN,
	//		Rhs: []ast.Expr{
	//			ast.NewIdent("arg." + strings.Title(pgArg.Name.Name)),
	//		},
	//	}
	//	notImpl.Body.List = append(notImpl.Body.List, stmt)
	//}
	//返回 nil
	rnil := ast.ReturnStmt{
		Results: []ast.Expr{
			ast.NewIdent("nil"),
		},
	}
	notImpl.Body.List = append(notImpl.Body.List, &rnil)
	p.serviceFile.Decls = append(p.serviceFile.Decls, notImpl)
}

//生成客户端函数代码
func (p *parseFile) addClientMethod(ifc methodType, m *fun) {
	var reqStruct, respStruct *ast.Ident   //请求和返回的结构体
	var reqParamName, respParamName string //请求和返回的参数名
	reqStruct = m.RequestStructType(true, p.packageName+"Obj")
	if reqStruct != nil {
		reqParamName = m.RequestParamName(true, "").Name
		if len(m.Params) > 1 {
			reqParamName = "req"
		}
	}
	respStruct = m.ResponseStructName(true, p.packageName+"Obj")
	if respStruct != nil {
		respParamName = "resp"
	}
	notImpl := funcDecl()
	notImpl.Name = m.Name
	var structName string
	if ifc != nil {
		notImpl.Recv = ifc.Reciever(rpcxClientStructSuffix)
		if v, ok := notImpl.Recv.List[0].Type.(*ast.StarExpr); ok {
			if vn := v.X.(*ast.Ident); ok {
				v.X = ast.NewIdent(nameFirstWordToLower(vn.Name))
			}
		}
		structName = ifc.StructName().Name
	} else {
		notImpl.Recv = nil
	}
	//生成请求参数
	notImpl.Type.Params = argListToFieldList(func(a Arg) *ast.Field {
		return a.Field(nil, false)
	}, false, m.Params...)
	result := argListToFieldList(func(a Arg) *ast.Field {
		return a.Field(nil, true)
	}, false, m.Results...)
	var returnValue []ast.Expr
	for _, v := range result.List { //response结构体中的字段拆解返回
		if len(v.Names) > 0 {
			returnValue = append(returnValue, ast.NewIdent(respParamName+"."+v.Names[0].Name))
		}
	}
	for _, v := range result.List { //清空返回字段的名称,用于函数返回
		v.Names = nil
	}
	notImpl.Type.Results = result
	var ret *ast.CallExpr //XClient.Call语句
	if ifc == nil {
		ret = &ast.CallExpr{
			Fun: ast.NewIdent("gClient.Call"),
		}
	} else {
		ret = &ast.CallExpr{
			Fun: ast.NewIdent(notImpl.Recv.List[0].Names[0].Name + ".client.Call"),
		}
	}
	notImpl.Body.List = []ast.Stmt{
		&ast.DeclStmt{
			Decl: &ast.GenDecl{
				Tok: token.VAR,
				Specs: []ast.Spec{
					&ast.ValueSpec{
						Names: []*ast.Ident{
							ast.NewIdent("err"),
						},
						Type: ast.NewIdent("error"),
					},
				},
			},
		},
	}
	var args []ast.Expr
	if ok, index := m.HasContextParam(); ok {
		args = append(args, &ast.BasicLit{
			Kind:  token.TYPE,
			Value: m.Params[index].Name.Name,
		})
	} else {
		args = append(args, &ast.BasicLit{
			Kind:  token.TYPE,
			Value: "context.Background()",
		})
	}
	// 生产XClient.Call所需的参数
	args = append(args, &ast.BasicLit{
		Kind:  token.STRING,
		Value: "\"" + m.Name.Name + "\"",
	})
	if reqParamName != "" {
		args = append(args, &ast.BasicLit{
			Kind:  token.TYPE,
			Value: reqParamName,
		})
	} else {
		args = append(args, &ast.BasicLit{Kind: token.STRING, Value: "nil"})
	}
	if respParamName != "" {
		args = append(args, &ast.BasicLit{
			Kind:  token.TYPE,
			Value: respParamName,
		})
	} else {
		args = append(args, &ast.BasicLit{Kind: token.STRING, Value: "nil"})
	}
	ret.Args = args
	// 创建请求参数
	var argr *ast.AssignStmt
	if reqParamName != "" {
		argr = &ast.AssignStmt{
			Tok: token.DEFINE,
			Lhs: []ast.Expr{ast.NewIdent(reqParamName)},
		}
		if len(m.Params) == 1 {
			if isContext(m.Params[0].Typ) {
				m.Params = nil
			} else {
				argr = nil
			}
		} else {
			var kvs []ast.Expr //请求结构体字段信息
			for _, a := range m.Params {
				if isContext(a.Typ) {
					continue
				}
				// 字段名称和参数名称进行复制 {xx:xx}
				kvs = append(kvs, &ast.KeyValueExpr{
					Key:   ast.NewIdent(nameFirstWordToUpper(a.Name.Name)),
					Value: ast.NewIdent(a.Name.Name),
				})
			}
			argr.Rhs = []ast.Expr{
				&ast.UnaryExpr{ //初始化结构体
					Op: token.AND,
					X: &ast.CompositeLit{
						Type: reqStruct,
						Elts: kvs,
					},
				},
			}
		}
	}
	if argr != nil {
		notImpl.Body.List = append(notImpl.Body.List, argr)
	}
	// 创建返回参数
	if respParamName != "" {
		reply := &ast.AssignStmt{
			Tok: token.DEFINE,
			Rhs: []ast.Expr{
				&ast.UnaryExpr{
					Op: token.AND,
					X: &ast.CompositeLit{
						Type: respStruct,
					},
				},
			},
			Lhs: []ast.Expr{ast.NewIdent(respParamName)},
		}
		notImpl.Body.List = append(notImpl.Body.List, reply)
	}
	// 定义XClient.Call返回的error变量
	r := ast.AssignStmt{
		Tok: token.ASSIGN,
		Rhs: []ast.Expr{ret},
		Lhs: []ast.Expr{ast.NewIdent("err")},
	}
	if ifc == nil { //判断gClient是不是nil,是nil的话提示生成错误
		var stmt ast.Stmt
		stmt = &ast.IfStmt{
			Cond: ast.NewIdent("gClient == nil"),
			Body: &ast.BlockStmt{
				List: []ast.Stmt{
					&ast.AssignStmt{
						Lhs: []ast.Expr{
							ast.NewIdent("err"),
						},
						Tok: token.ASSIGN,
						Rhs: []ast.Expr{
							&ast.CallExpr{
								Fun: ast.NewIdent("errors.New"),
								Args: []ast.Expr{
									ast.NewIdent("\"gClient尚未初始化\""),
								},
							},
						},
					},
				},
			},
			Else: &r,
		}
		notImpl.Body.List = append(notImpl.Body.List, stmt)
	} else {
		notImpl.Body.List = append(notImpl.Body.List, &r)
	}
	// 错误处理语句
	var errStmt []ast.Stmt
	// 错误记录日志
	errStmt = []ast.Stmt{
		&ast.IfStmt{
			Cond: ast.NewIdent("lg != nil"),
			Body: &ast.BlockStmt{
				List: []ast.Stmt{
					&ast.ExprStmt{
						X: &ast.CallExpr{
							Fun: &ast.Ident{
								Name: "lg.Errorf",
							},
							Args: []ast.Expr{
								&ast.Ident{Name: `"[` + p.fileName + " - " + structName + "." + m.Name.Name + `]RPC调用错误:%s",err.Error()`},
							},
						},
					},
				},
			},
		},
	}
	// 生成错误处理语句,如果返回结果中有error类型,将XClient.Call错误定义成appTool.RpcxError
	if ok, index := m.HasErrorResult(); ok && respParamName != "" {
		param := m.Results[index]
		errStmt = append(errStmt, &ast.AssignStmt{
			Lhs: []ast.Expr{
				&ast.SelectorExpr{
					X:   ast.NewIdent(respParamName),
					Sel: ast.NewIdent(nameFirstWordToUpper(param.Name.String())),
				},
			},
			Tok: token.ASSIGN,
			Rhs: []ast.Expr{
				ast.NewIdent("appTool.RpcxError"),
			},
		})
	}
	s := &ast.IfStmt{ //判断XClient.Call调用是否出错,出错则调用上面生成的错误语句处理错误
		Cond: &ast.BasicLit{
			Kind:  token.TYPE,
			Value: "err != nil",
		},
		Body: &ast.BlockStmt{
			List: errStmt,
		},
	}
	notImpl.Body.List = append(notImpl.Body.List, s)
	// 返回语句
	rnil := ast.ReturnStmt{
		Results: returnValue,
	}
	notImpl.Body.List = append(notImpl.Body.List, &rnil)
	p.clientFile.Decls = append(p.clientFile.Decls, notImpl)
}

// 导出客户端公共方法名
func (p *parseFile) exportClientPublicFunc(ifc methodType, rpcxService string) {
	initFunc := funcDecl()
	initFunc.Recv = ifc.Reciever(rpcxClientStructSuffix)
	if v, ok := initFunc.Recv.List[0].Type.(*ast.StarExpr); ok {
		if vn := v.X.(*ast.Ident); ok {
			v.X = ast.NewIdent(nameFirstWordToLower(vn.Name))
		}
	}
	initFunc.Name = ast.NewIdent("Init")
	requet := field(ast.NewIdent("client"), nil)
	requet.Type = &ast.BasicLit{
		Kind:  token.TYPE,
		Value: "client.XClient",
	}
	initFunc.Type.Params = fieldList(requet)
	initFunc.Type.Results = nil
	initFunc.Body.List = []ast.Stmt{
		&ast.AssignStmt{
			Lhs: []ast.Expr{
				ast.NewIdent(nameFirstWordToLower(ifc.StructName().Name[:1]) + ".client"),
			},
			Tok: token.ASSIGN,
			Rhs: []ast.Expr{
				ast.NewIdent("client"),
			},
		},
	}
	p.clientFile.Decls = append(p.clientFile.Decls, initFunc)
	servicePathFun := funcDecl()
	servicePathFun.Recv = ifc.Reciever(rpcxClientStructSuffix)
	if v, ok := servicePathFun.Recv.List[0].Type.(*ast.StarExpr); ok {
		if vn := v.X.(*ast.Ident); ok {
			v.X = ast.NewIdent(nameFirstWordToLower(vn.Name))
		}
	}
	servicePathFun.Name = ast.NewIdent("ServicePath")
	servicePathFun.Type.Params = nil
	servicePathFun.Type.Results = fieldList(&ast.Field{
		Names: nil,
		Type:  ast.NewIdent("string"),
	})
	servicePathFun.Body.List = []ast.Stmt{
		&ast.ReturnStmt{
			Results: []ast.Expr{
				ast.NewIdent("\"" + rpcxService + "\""),
			},
		},
	}
	p.clientFile.Decls = append(p.clientFile.Decls, servicePathFun)
}
