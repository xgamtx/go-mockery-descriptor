package fieldoverwriter

import (
	"errors"
	"regexp"
	"strings"
)

var errInvalidFieldOverwriterParams = errors.New("invalid field overwriter params")

type Overwriter interface {
	GetFuncPath() string
	GetFuncName() string
}

func getAliasFromPath(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		return path
	}

	re := regexp.MustCompile("v[0-9]+")
	if re.MatchString(parts[len(parts)-1]) {
		return parts[len(parts)-2]
	}

	return parts[len(parts)-1]
}

type FieldOverwriter struct {
	methodName string
	fieldName  string
	funcPath   string
	funcName   string
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

	return &FieldOverwriter{
		methodName: match[1],
		fieldName:  match[2],
		funcPath:   funcPath,
		funcName:   funcName,
	}, nil
}

func (f *FieldOverwriter) GetMethodName() string {
	return f.methodName
}

func (f *FieldOverwriter) GetFieldName() string {
	return f.fieldName
}

func (f *FieldOverwriter) GetFuncPath() string {
	return f.funcPath
}

func (f *FieldOverwriter) GetFuncName() string {
	return f.funcName
}

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

func (s *Storage) Get(methodName, paramName string, _ int) Overwriter {
	for _, overwriter := range s.overwriters {
		if methodName == overwriter.GetMethodName() && paramName == overwriter.GetFieldName() {
			return &overwriter
		}
	}

	return nil
}

// TODO support field index
// TODO support several overwriters
// TODO add alias to function name
