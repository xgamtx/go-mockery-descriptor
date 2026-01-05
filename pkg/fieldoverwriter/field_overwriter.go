package fieldoverwriter

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
)

const (
	stdFuncOneOf             = "oneOf"
	stdFunctionElementsMatch = "elementsMatch"
)

type stdFuncDescription struct {
	Name         string
	Path         string
	TypeModifier func(originalType string) string
}

var stdFunctions = map[string]stdFuncDescription{ //nolint:gochecknoglobals
	stdFuncOneOf: {
		Name:         "assessor.OneOf",
		Path:         "github.com/xgamtx/go-mockery-descriptor/pkg/assessor",
		TypeModifier: func(originalType string) string { return "[]" + originalType },
	},
	stdFunctionElementsMatch: {
		Name:         "assessor.ElementsMatch",
		Path:         "github.com/xgamtx/go-mockery-descriptor/pkg/assessor",
		TypeModifier: func(originalType string) string { return originalType },
	},
}

var errInvalidFieldOverwriterParams = errors.New("invalid field overwriter params")

type Overwriter interface {
	GetFuncPath() string
	GetFuncName() string
	ModifyType(original string) string
}

func getAliasFromPath(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) < 2 { //nolint:mnd
		return path
	}

	re := regexp.MustCompile("v[0-9]+")
	if re.MatchString(parts[len(parts)-1]) {
		return parts[len(parts)-2]
	}

	return parts[len(parts)-1]
}

func tryParseUnsignedInt(s string) int {
	val, err := strconv.Atoi(s)
	if err != nil || val < 0 {
		return -1
	}

	return val
}

type FieldOverwriter struct {
	methodName   string
	fieldName    *string
	fieldIndex   *int
	funcPath     string
	funcName     string
	typeModifier func(originalType string) string // currently supported on std functions
}

func newFieldOverwriter(params string) (*FieldOverwriter, error) {
	paramsParser := regexp.MustCompile(`^([a-zA-Z0-9]+)\.([a-zA-Z0-9]+)=((.+)\.)?([a-zA-Z0-9]+)$`)
	match := paramsParser.FindStringSubmatch(params)
	if len(match) == 0 {
		return nil, errInvalidFieldOverwriterParams
	}

	funcPath := match[4]
	funcName := match[5]
	if alias := getAliasFromPath(funcPath); alias != "" {
		funcName = alias + "." + funcName
	}

	typeModifier := func(originalType string) string { return originalType }
	if stdFunc := getStdFunction(funcPath, funcName); stdFunc != nil {
		funcPath = stdFunc.Path
		funcName = stdFunc.Name
		typeModifier = stdFunc.TypeModifier
	}
	var fieldIndex *int
	fieldName := &match[2]
	if parsedFieldName := tryParseUnsignedInt(*fieldName); parsedFieldName >= 0 {
		fieldIndex = &parsedFieldName
		fieldName = nil
	}

	return &FieldOverwriter{
		methodName:   match[1],
		fieldName:    fieldName,
		fieldIndex:   fieldIndex,
		funcPath:     funcPath,
		funcName:     funcName,
		typeModifier: typeModifier,
	}, nil
}

func getStdFunction(funcPath, funcName string) *stdFuncDescription {
	if funcPath != "" {
		return nil
	}

	stdFunc, ok := stdFunctions[funcName]
	if !ok {
		return nil
	}

	return &stdFunc
}

func (f *FieldOverwriter) GetMethodName() string             { return f.methodName }
func (f *FieldOverwriter) GetFieldName() *string             { return f.fieldName }
func (f *FieldOverwriter) GetFieldIndex() *int               { return f.fieldIndex }
func (f *FieldOverwriter) GetFuncPath() string               { return f.funcPath }
func (f *FieldOverwriter) GetFuncName() string               { return f.funcName }
func (f *FieldOverwriter) ModifyType(original string) string { return f.typeModifier(original) }

type Storage struct {
	overwriters []FieldOverwriter
}

func NewStorage(overwritersParams []string) (*Storage, error) {
	overwriters := make([]FieldOverwriter, 0, len(overwritersParams))
	for _, param := range overwritersParams {
		overwriter, err := newFieldOverwriter(param)
		if err != nil {
			return nil, err
		}

		overwriters = append(overwriters, *overwriter)
	}

	return &Storage{overwriters: overwriters}, nil
}

func (s *Storage) Get(methodName, paramName string, index int) Overwriter {
	for _, overwriter := range s.overwriters {
		if methodName == overwriter.GetMethodName() && ((overwriter.GetFieldName() != nil && paramName == *overwriter.GetFieldName()) ||
			(overwriter.GetFieldIndex() != nil && index == *overwriter.GetFieldIndex())) {
			return &overwriter
		}
	}

	return nil
}
