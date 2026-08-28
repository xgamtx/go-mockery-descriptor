package parser

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/packages"
)

// loadMode loads the package along with its dependencies: types of interfaces embedded from
// other packages are only available together with them.
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

// Interface describes a generation target: either an interface or a function type.
// For a function type the single "method" is named after the type itself.
type Interface struct {
	PackageName string
	Name        string
	IsFunc      bool
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

	obj := pkg.Types.Scope().Lookup(interfaceName)
	if obj == nil {
		return nil, fmt.Errorf("%s is not found", interfaceName)
	}

	switch underlying := obj.Type().Underlying().(type) {
	case *types.Interface:
		// The AST is only needed to keep the declaration order: go/types returns methods sorted.
		declaredOrder, err := declaredMethodOrder(pkg.Syntax, interfaceName)
		if err != nil {
			return nil, err
		}

		return parseInterface(interfaceName, pkg.Types, underlying, declaredOrder), nil

	case *types.Signature:
		return parseFunc(interfaceName, pkg.Types, underlying), nil

	default:
		return nil, fmt.Errorf("%s is neither an interface nor a function type", interfaceName)
	}
}

// parseFunc describes a function type as a single call: mockery generates a mock with a single
// Execute method for such a type.
func parseFunc(funcName string, pkg *types.Package, sig *types.Signature) *Interface {
	return &Interface{
		PackageName: pkg.Name(),
		Name:        funcName,
		IsFunc:      true,
		Methods: []Method{{
			Name:    funcName,
			Params:  extractTuple(pkg, sig.Params()),
			Returns: extractTuple(pkg, sig.Results()),
		}},
	}
}

func declaredMethodOrder(files []*ast.File, name string) ([]string, error) {
	iface, err := getInterfaceByName(files, name)
	if err != nil {
		return nil, err
	}

	var order []string
	for _, method := range iface.Methods.List {
		// Embedded interfaces have no names: their methods are appended after the declared ones.
		for _, methodName := range method.Names {
			order = append(order, methodName.Name)
		}
	}

	return order, nil
}

func getInterfaceByName(files []*ast.File, name string) (*ast.InterfaceType, error) {
	for _, f := range files {
		for _, decl := range f.Decls {
			// Looking for a type declaration
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.TYPE {
				continue
			}

			for _, spec := range genDecl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}

				// Make sure this is an interface with the required name
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

	// First go the methods declared in the interface itself, in source order.
	handled := make(map[string]struct{}, len(declaredOrder))
	for _, methodName := range declaredOrder {
		method, ok := methods[methodName]
		if !ok {
			continue
		}

		handled[methodName] = struct{}{}
		result.Methods = append(result.Methods, parseMethod(pkg, method))
	}

	// Then go the methods promoted from embedded interfaces.
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

// renderType prints the type the way it must appear in the generated file and collects the
// paths of the packages this type refers to.
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
