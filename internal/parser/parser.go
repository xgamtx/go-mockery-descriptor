package parser

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/packages"
)

// loadMode загружает не только сам пакет, но и его зависимости: типы встроенных
// интерфейсов из других пакетов доступны только вместе с ними.
const loadMode = packages.NeedName | packages.NeedSyntax | packages.NeedTypes |
	packages.NeedTypesInfo | packages.NeedImports | packages.NeedDeps

type Value struct {
	Name      string
	Type      string
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
	cfg := &packages.Config{Mode: loadMode}
	pkgs, err := packages.Load(cfg, dir)
	if err != nil {
		return nil, err
	}

	if len(pkgs) != 1 {
		return nil, fmt.Errorf("expected exactly one package, got %d", len(pkgs))
	}

	pkg := pkgs[0]

	// AST нужен только чтобы сохранить порядок объявления методов: go/types отдаёт их отсортированными.
	declaredOrder, err := declaredMethodOrder(pkg.Syntax, interfaceName)
	if err != nil {
		return nil, err
	}

	iface, err := lookupInterface(pkg.Types, interfaceName)
	if err != nil {
		return nil, err
	}

	return parseInterface(interfaceName, pkg.Types, iface, declaredOrder), nil
}

// lookupInterface возвращает полный набор методов интерфейса, включая методы встроенных интерфейсов.
func lookupInterface(pkg *types.Package, name string) (*types.Interface, error) {
	obj := pkg.Scope().Lookup(name)
	if obj == nil {
		return nil, fmt.Errorf("%s is not found", name)
	}

	iface, ok := obj.Type().Underlying().(*types.Interface)
	if !ok {
		return nil, fmt.Errorf("%s is not an interface", name)
	}

	return iface, nil
}

func declaredMethodOrder(files []*ast.File, name string) ([]string, error) {
	iface, err := getInterfaceByName(files, name)
	if err != nil {
		return nil, err
	}

	var order []string
	for _, method := range iface.Methods.List {
		// Встроенные интерфейсы не имеют имён: их методы добавляются после явно объявленных.
		for _, methodName := range method.Names {
			order = append(order, methodName.Name)
		}
	}

	return order, nil
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

func parseInterface(
	interfaceName string,
	pkg *types.Package,
	iface *types.Interface,
	declaredOrder []string,
) *Interface {
	result := &Interface{
		PackageName: pkg.Name(),
		Name:        interfaceName,
		Methods:     make([]Method, 0, iface.NumMethods()),
	}

	methods := make(map[string]*types.Func, iface.NumMethods())
	for i := range iface.NumMethods() {
		method := iface.Method(i)
		methods[method.Name()] = method
	}

	// Сначала методы, объявленные в интерфейсе напрямую, — в порядке исходника.
	handled := make(map[string]struct{}, len(declaredOrder))
	for _, methodName := range declaredOrder {
		method, ok := methods[methodName]
		if !ok {
			continue
		}

		handled[methodName] = struct{}{}
		result.Methods = append(result.Methods, parseMethod(pkg, method))
	}

	// Затем методы, полученные из встроенных интерфейсов.
	for i := range iface.NumMethods() {
		method := iface.Method(i)
		if _, ok := handled[method.Name()]; ok {
			continue
		}

		result.Methods = append(result.Methods, parseMethod(pkg, method))
	}

	return result
}

func parseMethod(pkg *types.Package, method *types.Func) Method {
	desc := Method{Name: method.Name()}

	sig, ok := method.Type().(*types.Signature)
	if !ok {
		return desc
	}

	desc.Params = extractTuple(pkg, sig.Params())
	desc.Returns = extractTuple(pkg, sig.Results())

	return desc
}

func extractTuple(pkg *types.Package, tuple *types.Tuple) []Value {
	if tuple == nil || tuple.Len() == 0 {
		return nil
	}

	values := make([]Value, 0, tuple.Len())
	for i := range tuple.Len() {
		v := tuple.At(i)
		typeName, pathTypes := renderType(pkg, v.Type())
		values = append(values, Value{Name: v.Name(), Type: typeName, PathTypes: pathTypes})
	}

	return values
}

// renderType печатает тип так, как он должен выглядеть в генерируемом файле, и попутно
// собирает пути пакетов, на которые этот тип ссылается.
func renderType(pkg *types.Package, t types.Type) (string, []string) {
	var pathTypes []string

	qualifier := func(p *types.Package) string {
		if p == nil || p == pkg {
			return ""
		}

		pathTypes = append(pathTypes, p.Path())

		return p.Name()
	}

	return types.TypeString(t, qualifier), pathTypes
}
