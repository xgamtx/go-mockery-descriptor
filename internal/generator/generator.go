package generator

import (
	"fmt"
	"go/ast"
	"go/format"
	"strconv"
	"strings"

	"golang.org/x/tools/imports"

	"github.com/xgamtx/go-mockery-descriptor/internal/fieldoverwriter"
	"github.com/xgamtx/go-mockery-descriptor/internal/parser"
)

const (
	anyCtxConst = "anyCtx"
	anyTxConst  = "anyTx"
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

func newReturnView(v *parser.Value, i int) *returnView {
	t := exprToString(v.Type)
	name := v.Name
	if name == "" && t == "error" {
		name = "err"
	}
	if name == "" {
		name = "r" + strconv.Itoa(i)
	}

	return &returnView{Name: "Received" + capitalize(name), Type: t, PathTypes: v.PathTypes}
}

type methodView struct {
	Name    string
	Params  []param
	Returns []returnView
}

func newMethodView(method *parser.Method, fieldOverwriterStorage *fieldoverwriter.Storage) *methodView {
	res := &methodView{
		Name:    method.Name,
		Params:  make([]param, 0, len(method.Params)),
		Returns: make([]returnView, 0, len(method.Returns)),
	}
	for i, param := range method.Params {
		fieldOverwriter := fieldOverwriterStorage.Get(method.Name, param.Name, i)
		res.Params = append(res.Params, newParamView(&param, i, fieldOverwriter))
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
		if param.GenerateField() != "" {
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
		if view := param.GenerateField(); view != "" {
			lines = append(lines, view)
			paramsCount++
		}
	}

	if paramsCount > 0 && len(m.Returns) > 0 {
		lines = append(lines, "")
	}
	for _, r := range m.Returns {
		lines = append(lines, r.Name+" "+r.Type)
	}

	lines = append(lines, "}")

	return strings.Join(lines, "\n")
}

func (m *methodView) generateField() string {
	return m.getStructureFieldName() + " []" + m.getStructureName()
}

func (m *methodView) generateCall() string {
	lines := make([]string, 0, 3)
	if m.IsAnyField() {
		lines = append(lines, "for _, call := range calls."+m.getStructureFieldName()+" {")
	} else {
		lines = append(lines, "for range calls."+m.getStructureFieldName()+" {")
	}
	line := "m.EXPECT()." + m.Name + "("
	for i, param := range m.Params {
		if i > 0 {
			line += ", "
		}
		line += param.GenerateAssessor("call")
	}
	line += ").Return("
	for i, r := range m.Returns {
		if i > 0 {
			line += ", "
		}
		line += "call." + r.Name
	}
	line += ").Once()"

	lines = append(lines, line, "}")

	return strings.Join(lines, "\n")
}

type interfaceView struct {
	PackageName string
	Name        string
	Methods     []methodView
}

func newInterfaceView(iface *parser.Interface, fieldOverwriterStorage *fieldoverwriter.Storage) *interfaceView {
	res := &interfaceView{
		PackageName: iface.PackageName,
		Name:        iface.Name,
		Methods:     make([]methodView, 0, len(iface.Methods)),
	}
	for _, method := range iface.Methods {
		res.Methods = append(res.Methods, *newMethodView(&method, fieldOverwriterStorage))
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
	lines := make([]string, 0, 2+len(iv.Methods))
	lines = append(lines, "type "+iv.getStructureName()+" struct {")
	for _, m := range iv.Methods {
		lines = append(lines, m.generateField())
	}

	lines = append(lines, "}")

	return strings.Join(lines, "\n")
}

func (iv *interfaceView) generateConstructor() string {
	lines := []string{
		"func " + iv.getConstructureName() + "(t *testing.T, calls *" + iv.getStructureName() + ") " + iv.Name + " {",
		"t.Helper()",
		"m := newMock" + capitalize(iv.Name) + "(t)",
	}

	if additionalVars := iv.generateAdditionalVars(); additionalVars != "" {
		lines = append(lines, additionalVars)
	}

	for _, method := range iv.Methods {
		lines = append(lines, method.generateCall())
	}

	lines = append(lines, "return m")
	lines = append(lines, "}")

	return strings.Join(lines, "\n")
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

func (iv *interfaceView) generateAdditionalVars() string {
	res := make([]string, 0, 2) //nolint:mnd
	if iv.isCtxRequired() {
		res = append(res, anyCtxConst+" := mock.Anything")
	}
	if iv.isTxRequired() {
		res = append(res, anyTxConst+" := mock.Anything")
	}

	return strings.Join(res, "\n")
}

func (iv *interfaceView) generatePackageLine() string {
	return "package " + iv.PackageName
}

func (iv *interfaceView) getImports() []string {
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

func (iv *interfaceView) generateImports() string {
	imports := iv.getImports()
	switch len(imports) {
	case 0:
		return ""
	case 1:
		return "import " + imports[0]
	}
	lines := []string{"import ("}
	for _, imp := range imports {
		lines = append(lines, `"`+imp+`"`)
	}

	lines = append(lines, ")")

	return strings.Join(lines, "\n")
}

func Generate(iface *parser.Interface, fieldOverwriterStorage *fieldoverwriter.Storage) (string, error) {
	view := newInterfaceView(iface, fieldOverwriterStorage)

	lines := []string{
		"// Code generated by go-mockery-descriptor v1.0.0. DO NOT EDIT.",
		view.generatePackageLine(),
	}
	if imports := view.generateImports(); imports != "" {
		lines = append(lines, imports)
	}
	for _, method := range view.Methods {
		if methodStructure := method.generateStructure(); methodStructure != "" {
			lines = append(lines, method.generateStructure())
		}
	}
	lines = append(lines, view.generateStructure())
	lines = append(lines, view.generateConstructor())

	formattedOutput, err := format.Source([]byte(strings.Join(lines, "\n\n")))
	if err != nil {
		return "", err
	}

	formattedOutput, err = formatImports(formattedOutput)
	if err != nil {
		return "", err
	}

	return string(formattedOutput), nil
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
// TODO parametrize via config?
// TODO support function instead of interfaces
// TODO support package name override
// TODO add interface_name prefix option
