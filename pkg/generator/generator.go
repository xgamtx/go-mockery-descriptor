package generator

import (
	"go/ast"
	"strconv"
	"strings"

	"github.com/xgamtx/go-mockery-descriptor/pkg/parser"
)

type paramKind int

const (
	kindUnknown paramKind = iota
	kindCtx
	kindTx
)

func exprToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + exprToString(t.X)
	case *ast.SelectorExpr:
		return exprToString(t.X) + "." + t.Sel.Name
	case *ast.ArrayType:
		return "[]" + exprToString(t.Elt)
	case *ast.MapType:
		return "map[" + exprToString(t.Key) + "]" + exprToString(t.Value)
	case *ast.InterfaceType:
		if t.Methods == nil || len(t.Methods.List) == 0 {
			return "any"
		}

		return "interface{...}"
	case *ast.FuncType:
		return "func(/*...*/)"
	default:
		return "unknown"
	}
}

type paramView struct {
	Kind paramKind
	Name string
	Type string
}

func newParamView(v *parser.Value, i int) *paramView {
	t := exprToString(v.Type)
	switch t {
	case "context.Context":
		return &paramView{Kind: kindCtx}
	case "pgx.Tx":
		return &paramView{Kind: kindTx}
	}
	name := v.Name
	if name == "" {
		name = "p" + strconv.Itoa(i)
	}

	return &paramView{Kind: kindUnknown, Name: capitalize(name), Type: t}
}

type returnView struct {
	Name string
	Type string
}

func newReturnView(v *parser.Value, i int) *returnView {
	t := exprToString(v.Type)
	name := v.Name
	if v.Name == "" {
		name = "r" + strconv.Itoa(i)
	}

	return &returnView{Name: "Received" + capitalize(name), Type: t}
}

type methodView struct {
	Name    string
	Params  []paramView
	Returns []returnView
}

func newMethodView(method *parser.Method) *methodView {
	res := &methodView{
		Name:    method.Name,
		Params:  make([]paramView, 0, len(method.Params)),
		Returns: make([]returnView, 0, len(method.Returns)),
	}
	for i, param := range method.Params {
		res.Params = append(res.Params, *newParamView(&param, i))
	}
	for i, r := range method.Returns {
		res.Returns = append(res.Returns, *newReturnView(&r, i))
	}

	return res
}

func (m *methodView) IsAnyField() bool {
	if len(m.Returns) > 0 {
		return true
	}
	for _, param := range m.Params {
		if param.Kind == kindUnknown {
			return true
		}
	}

	return false
}

func (m *methodView) getStructureName() string {
	return unCapitalize(m.Name) + "Call"
}

func (m *methodView) getStructureFieldName() string {
	return capitalize(m.Name)
}

func (m *methodView) generateStructure() string {
	if !m.IsAnyField() {
		return "type " + m.getStructureName() + " struct {}"
	}

	lines := []string{
		"type " + m.getStructureName() + " struct {",
	}
	var paramsCount int
	for _, param := range m.Params {
		if param.Kind == kindUnknown {
			lines = append(lines, "\t"+param.Name+" "+param.Type)
			paramsCount++
		}
	}

	if paramsCount > 0 && len(m.Returns) > 0 {
		lines = append(lines, "")
	}
	for _, r := range m.Returns {
		lines = append(lines, "\t"+r.Name+" "+r.Type)
	}

	lines = append(lines, "}")

	return strings.Join(lines, "\n")
}

func (m *methodView) generateField() string {
	return m.getStructureFieldName() + " []" + m.getStructureName()
}

func (m *methodView) generateCall() string {
	var lines []string
	if m.IsAnyField() {
		lines = []string{"\tfor _, call := range calls." + m.getStructureFieldName() + " {"}
	} else {
		lines = []string{"\tfor range calls." + m.getStructureFieldName() + " {"}
	}
	line := "\t\tm.EXPECT()." + m.Name + "("
	for i, param := range m.Params {
		if i > 0 {
			line += ", "
		}
		switch param.Kind {
		case kindUnknown:
			line += "call." + param.Name
		case kindCtx:
			line += "anyCtx"
		case kindTx:
			line += "anyTx"
		}
	}
	line += ").Return("
	for i, r := range m.Returns {
		if i > 0 {
			line += ", "
		}
		line += "call." + r.Name
	}
	line += ").Once()"

	lines = append(lines, line, "\t}")

	return strings.Join(lines, "\n")
}

type interfaceView struct {
	Name    string
	Methods []methodView
}

func newInterfaceView(iface *parser.Interface) *interfaceView {
	res := &interfaceView{
		Name:    iface.Name,
		Methods: make([]methodView, 0, len(iface.Methods)),
	}
	for _, method := range iface.Methods {
		res.Methods = append(res.Methods, *newMethodView(&method))
	}

	return res
}

func (iv *interfaceView) getStructureName() string {
	return unCapitalize(iv.Name) + "Calls"
}

func (iv *interfaceView) getConstructureName() string {
	return "make" + capitalize(iv.Name) + "Mock"
}

func (iv *interfaceView) generateStructure() string {
	lines := []string{
		"type " + iv.getStructureName() + " struct {",
	}
	for _, m := range iv.Methods {
		lines = append(lines, "\t"+m.generateField())
	}

	lines = append(lines, "}")

	return strings.Join(lines, "\n")
}

func (iv *interfaceView) generateConstructor() string {
	lines := []string{
		"func " + iv.getConstructureName() + "(t *testing.T, calls *" + iv.getStructureName() + ") " + iv.Name + " {",
		"\tt.Helper()",
		"\tm := NewMock" + capitalize(iv.Name) + "(t)",
	}
	for _, method := range iv.Methods {
		lines = append(lines, method.generateCall())
	}

	lines = append(lines, "\treturn m")
	lines = append(lines, "}")

	return strings.Join(lines, "\n")
}

func (iv *interfaceView) isCtxRequired() bool {
	for _, m := range iv.Methods {
		for _, param := range m.Params {
			if param.Kind == kindCtx {
				return true
			}
		}
	}

	return false
}

func (iv *interfaceView) isTxRequired() bool {
	for _, m := range iv.Methods {
		for _, param := range m.Params {
			if param.Kind == kindTx {
				return true
			}
		}
	}

	return false
}

func generateAdditionalVars(iface *interfaceView) string {
	ctxRequired := iface.isCtxRequired()
	txRequired := iface.isTxRequired()
	switch {
	case !ctxRequired && !txRequired:
		return ""
	case ctxRequired && txRequired:
		lines := []string{
			"var (",
			"\tanyCtx = mock.Anything",
			"\tanyTx = mock.Anything",
			")",
		}

		return strings.Join(lines, "\n")
	case ctxRequired:
		return "var anyCtx = mock.Anything"
	default:
		return "var anyTx = mock.Anything"
	}
}

func Generate(iface *parser.Interface) string {
	view := newInterfaceView(iface)

	lines := []string{
		"// Code generated by mock-galls-generator v1.0.0. DO NOT EDIT.",
		"package example\n\nimport (\n\t\"testing\"\n\n\t\"github.com/stretchr/testify/mock\"\n)",
	}
	if additionalVars := generateAdditionalVars(view); additionalVars != "" {
		lines = append(lines, additionalVars)
	}
	for _, method := range view.Methods {
		if methodStructure := method.generateStructure(); methodStructure != "" {
			lines = append(lines, method.generateStructure())
		}
	}
	lines = append(lines, view.generateStructure())
	lines = append(lines, view.generateConstructor())

	return strings.Join(lines, "\n\n")
}

// TODO add additional imports support.
// TODO add check if anyCtx/anyTx exists.
// TODO add custom functions.
// TODO add sub interface support.
