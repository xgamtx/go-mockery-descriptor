package generator

import (
	"bytes"
	_ "embed"
	"fmt"
	"go/ast"
	"go/format"
	"strconv"
	"text/template"

	"golang.org/x/tools/imports"

	"github.com/xgamtx/go-mockery-descriptor/internal/config"
	"github.com/xgamtx/go-mockery-descriptor/internal/fieldoverwriter"
	"github.com/xgamtx/go-mockery-descriptor/internal/parser"
	"github.com/xgamtx/go-mockery-descriptor/internal/returnsrenamer"
)

const (
	anyCtxConst = "anyCtx"
	anyTxConst  = "anyTx"
)

//go:embed mock.tmpl
var tmplContent string

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

type param interface {
	GenerateField() string
	GenerateAssessor(callerName string) string
	GetPathTypes() []string
}

type stdParamView struct {
	name      string
	paramType string
	pathTypes []string
}

type ctxParamView struct{}

func (v *ctxParamView) GenerateField() string          { return "" }
func (v *ctxParamView) GenerateAssessor(string) string { return anyCtxConst }
func (v *ctxParamView) GetPathTypes() []string         { return nil }

type txParamView struct{}

func (v *txParamView) GenerateField() string          { return "" }
func (v *txParamView) GenerateAssessor(string) string { return anyTxConst }
func (v *txParamView) GetPathTypes() []string         { return nil }

type customFunctionParamView struct {
	paramName string
	paramType string
	funcName  string
	pathTypes []string
}

func newCustomFunctionParamView(v *parser.Value, fieldOverwriter fieldoverwriter.Overwriter) *customFunctionParamView {
	pathTypes := v.PathTypes
	if pathType := fieldOverwriter.GetFuncPath(); pathType != "" {
		pathTypes = append(pathTypes, pathType)
	}

	return &customFunctionParamView{
		paramName: capitalize(v.Name),
		paramType: fieldOverwriter.ModifyType(exprToString(v.Type)),
		funcName:  fieldOverwriter.GetFuncName(),
		pathTypes: pathTypes,
	}
}

func (v *customFunctionParamView) GenerateField() string {
	if v.paramType == "" {
		return ""
	}

	return v.paramName + " " + v.paramType
}

func (v *customFunctionParamView) GenerateAssessor(callerName string) string {
	if v.GenerateField() == "" {
		return v.funcName
	}

	return fmt.Sprintf("%s(%s.%s)", v.funcName, callerName, v.paramName)
}

func (v *customFunctionParamView) GetPathTypes() []string { return v.pathTypes }

func newParamView(v *parser.Value, i int, fieldOverwriter fieldoverwriter.Overwriter) param {
	if fieldOverwriter != nil {
		return newCustomFunctionParamView(v, fieldOverwriter)
	}

	t := exprToString(v.Type)
	switch t {
	case "context.Context":
		return &ctxParamView{}
	case "pgx.Tx":
		return &txParamView{}
	}
	name := v.Name
	if name == "" {
		name = "p" + strconv.Itoa(i)
	}

	return &stdParamView{name: capitalize(name), paramType: t, pathTypes: v.PathTypes}
}

func (p *stdParamView) GenerateField() string {
	return p.name + " " + p.paramType
}

func (p *stdParamView) GenerateAssessor(callerName string) string {
	return callerName + "." + p.name
}

func (p *stdParamView) GetPathTypes() []string {
	return p.pathTypes
}

type returnView struct {
	Name      string
	Type      string
	PathTypes []string
}

func newReturnView(v *parser.Value, i int, returnsRenamer *returnsrenamer.ReturnRenamer) *returnView {
	t := exprToString(v.Type)
	name := v.Name
	if name == "" && t == "error" {
		name = "err"
	}
	if name == "" {
		name = "r" + strconv.Itoa(i)
	}
	if returnsRenamer != nil && name == returnsRenamer.GetOldReturnName() {
		name = returnsRenamer.GetNewReturnName()
	}

	return &returnView{Name: "Received" + capitalize(name), Type: t, PathTypes: v.PathTypes}
}

type methodView struct {
	Name    string
	Params  []param
	Returns []returnView
}

func newMethodView(
	method *parser.Method,
	fieldOverwriterStorage *fieldoverwriter.Storage,
	returnsRenamerStorage *returnsrenamer.Storage,
) *methodView {
	res := &methodView{
		Name:    method.Name,
		Params:  make([]param, 0, len(method.Params)),
		Returns: make([]returnView, 0, len(method.Returns)),
	}
	for i, param := range method.Params {
		fieldOverwriter := fieldOverwriterStorage.Get(method.Name, param.Name, i)
		res.Params = append(res.Params, newParamView(&param, i, fieldOverwriter))
	}
	returnRenamer := returnsRenamerStorage.GetReturnRenamer(method.Name)
	for i, r := range method.Returns {
		res.Returns = append(res.Returns, *newReturnView(&r, i, returnRenamer))
	}

	return res
}

func (m *methodView) IsAnyField() bool {
	if len(m.Returns) > 0 {
		return true
	}
	for _, param := range m.Params {
		if param.GenerateField() != "" {
			return true
		}
	}

	return false
}

func (m *methodView) GetStructureName() string {
	return unCapitalize(m.Name) + "Call"
}

func (m *methodView) GetStructureFieldName() string {
	return capitalize(m.Name)
}

type interfaceView struct {
	PackageName string
	Name        string
	Methods     []methodView
}

func newInterfaceView(
	iface *parser.Interface,
	fieldOverwriterStorage *fieldoverwriter.Storage,
	returnsRenamerStorage *returnsrenamer.Storage,
) *interfaceView {
	res := &interfaceView{
		PackageName: iface.PackageName,
		Name:        iface.Name,
		Methods:     make([]methodView, 0, len(iface.Methods)),
	}
	for _, method := range iface.Methods {
		res.Methods = append(res.Methods, *newMethodView(&method, fieldOverwriterStorage, returnsRenamerStorage))
	}

	return res
}

func (iv *interfaceView) GetStructureName() string {
	return unCapitalize(iv.Name) + "Calls"
}

func (iv *interfaceView) GetConstructureName() string {
	return "make" + capitalize(iv.Name) + "Mock"
}

func (iv *interfaceView) AdditionalVars() []string {
	res := make([]string, 0, 2) //nolint:mnd
	if iv.isCtxRequired() {
		res = append(res, anyCtxConst+" := mock.Anything")
	}
	if iv.isTxRequired() {
		res = append(res, anyTxConst+" := mock.Anything")
	}

	return res
}

func (iv *interfaceView) GetImports() []string {
	res := make([]string, 0, 2)
	res = append(res, "testing", "github.com/stretchr/testify/mock")
	for _, m := range iv.Methods {
		for _, param := range m.Params {
			res = append(res, param.GetPathTypes()...)
		}
		for _, ret := range m.Returns {
			res = append(res, ret.PathTypes...)
		}
	}

	return unique(res)
}

func (iv *interfaceView) isCtxRequired() bool {
	for _, m := range iv.Methods {
		for _, param := range m.Params {
			if param.GenerateAssessor("") == anyCtxConst {
				return true
			}
		}
	}

	return false
}

func (iv *interfaceView) isTxRequired() bool {
	for _, m := range iv.Methods {
		for _, param := range m.Params {
			if param.GenerateAssessor("") == anyTxConst {
				return true
			}
		}
	}

	return false
}

func Generate(
	cfg *config.Config,
	iface *parser.Interface,
	fieldOverwriterStorage *fieldoverwriter.Storage,
	returnsRenamerStorage *returnsrenamer.Storage,
) (string, error) {
	view := newInterfaceView(iface, fieldOverwriterStorage, returnsRenamerStorage)
	tmpl := template.New("mock.tmpl")
	tmpl, err := tmpl.Parse(`{{ define "constructor" }}` + cfg.ConstructorName + `{{ end }} ` + tmplContent)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err = tmpl.Execute(&buf, view); err != nil {
		return "", err
	}

	formatted := buf.Bytes()
	formatted1, err := format.Source(formatted)
	if err == nil {
		formatted = formatted1
	}

	formatted1, err = formatImports(formatted)
	if err == nil {
		formatted = formatted1
	}

	return string(formatted), nil
}

func formatImports(content []byte) ([]byte, error) {
	return imports.Process("", content, nil)
}

func unique(vals []string) []string {
	res := make([]string, 0, len(vals))
	m := make(map[string]struct{})
	for _, val := range vals {
		if _, ok := m[val]; !ok {
			m[val] = struct{}{}
			res = append(res, val)
		}
	}

	return res
}

// TODO add sub interface support.
// TODO support function instead of interfaces
// TODO support package name override
// TODO add interface_name prefix option
