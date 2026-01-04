package parser

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

type Value struct {
	Name string
	Type ast.Expr
}

type Method struct {
	Name    string
	Params  []Value
	Returns []Value
}

type Interface struct {
	Name    string
	Methods []Method
}

func ParseInterfaceInDir(fileName, interfaceName string) (*Interface, error) {
	// Создаём набор токенов
	fset := token.NewFileSet()

	// Парсим файл
	node, err := parser.ParseFile(fset, fileName, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	iface, err := getInterfaceByName(node, interfaceName)
	if err != nil {
		return nil, err
	}

	return parseInterface(interfaceName, iface), nil
}

func getInterfaceByName(f *ast.File, name string) (*ast.InterfaceType, error) {
	// Проходим по всем декларациям в файле
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

	return nil, fmt.Errorf("%s is not found", name)
}

func parseInterface(name string, iface *ast.InterfaceType) *Interface {
	result := &Interface{
		Name:    name,
		Methods: make([]Method, 0, len(iface.Methods.List)),
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
			desc.Params = extractFields(funcType.Params.List)
		}

		// Обрабатываем возвращаемые значения
		if funcType.Results != nil {
			desc.Returns = extractFields(funcType.Results.List)
		}

		result.Methods = append(result.Methods, desc)
	}

	return result
}

func extractFields(fields []*ast.Field) []Value {
	var values []Value

	for _, field := range fields {
		if len(field.Names) == 0 {
			// Анонимный параметр (часто в возвращаемых значениях)
			values = append(values, Value{Name: "", Type: field.Type})
		} else {
			for _, name := range field.Names {
				values = append(values, Value{Name: name.Name, Type: field.Type})
			}
		}
	}

	return values
}
