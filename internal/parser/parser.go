package parser

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/packages"
)

type Value struct {
	Name      string
	Type      ast.Expr
	PathTypes []string
}

type Method struct {
	Name    string
	Params  []Value
	Returns []Value
}

type Interface struct {
	PackageName string
	Name        string
	Methods     []Method
}

func ParseInterfaceInDir(dir, interfaceName string) (*Interface, error) {
	cfg := &packages.Config{Mode: packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo}
	pkgs, err := packages.Load(cfg, dir)
	if err != nil {
		return nil, err
	}

	if len(pkgs) != 1 {
		return nil, fmt.Errorf("expected exactly one package, got %d", len(pkgs))
	}

	iface, err := getInterfaceByName(pkgs[0].Syntax, interfaceName)
	if err != nil {
		return nil, err
	}

	return parseInterface(interfaceName, pkgs[0].Types.Name(), iface, pkgs[0].TypesInfo), nil
}

func getInterfaceByName(files []*ast.File, name string) (*ast.InterfaceType, error) {
	for _, f := range files {
		for _, decl := range f.Decls {
			// Ищем декларацию типа
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.TYPE {
				continue
			}

			for _, spec := range genDecl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}

				// Проверяем, что это интерфейс с нужным именем
				if typeSpec.Name.Name != name {
					continue
				}

				iface, ok := typeSpec.Type.(*ast.InterfaceType)
				if !ok {
					return nil, fmt.Errorf("%s is not an interface", name)
				}

				return iface, nil
			}
		}
	}

	return nil, fmt.Errorf("%s is not found", name)
}

func getImportsForExpr(expr ast.Expr, typesInfo *types.Info) []string {
	switch t := expr.(type) {
	case *ast.Ident:
		if obj := typesInfo.ObjectOf(t); obj != nil {
			if pkgName, ok := obj.(*types.PkgName); ok {
				return []string{pkgName.Imported().Path()}
			}
		}

		return nil // internal or core

	case *ast.SelectorExpr:
		// Type with package
		return getImportsForExpr(t.X, typesInfo)

	case *ast.StarExpr:
		// Pointer: *Type
		return getImportsForExpr(t.X, typesInfo)

	case *ast.ArrayType:
		// Slice: []Type
		return getImportsForExpr(t.Elt, typesInfo)

	case *ast.MapType:
		// Map: map[Key]Value
		key := getImportsForExpr(t.Key, typesInfo)
		value := getImportsForExpr(t.Value, typesInfo)
		if key == nil && value == nil {
			return nil
		}

		return append(key, value...)

	default:
		return nil
	}
}

func parseInterface(interfaceName, packageName string, iface *ast.InterfaceType, typesInfo *types.Info) *Interface {
	result := &Interface{
		PackageName: packageName,
		Name:        interfaceName,
		Methods:     make([]Method, 0, len(iface.Methods.List)),
	}

	for _, method := range iface.Methods.List {
		// Пропускаем встроенные интерфейсы (embedding)
		if len(method.Names) == 0 {
			continue
		}

		methodName := method.Names[0].Name

		funcType, ok := method.Type.(*ast.FuncType)
		if !ok {
			continue
		}

		desc := Method{Name: methodName}

		// Обрабатываем параметры
		if funcType.Params != nil {
			desc.Params = extractFields(funcType.Params.List, typesInfo)
		}

		// Обрабатываем возвращаемые значения
		if funcType.Results != nil {
			desc.Returns = extractFields(funcType.Results.List, typesInfo)
		}

		result.Methods = append(result.Methods, desc)
	}

	return result
}

func extractFields(fields []*ast.Field, typesInfo *types.Info) []Value {
	var values []Value

	for _, field := range fields {
		imports := getImportsForExpr(field.Type, typesInfo)
		if len(field.Names) == 0 {
			// Анонимный параметр (часто в возвращаемых значениях)
			values = append(values, Value{Name: "", Type: field.Type, PathTypes: imports})
		} else {
			for _, name := range field.Names {
				values = append(values, Value{Name: name.Name, Type: field.Type, PathTypes: imports})
			}
		}
	}

	return values
}
