package returnsrenamer

import (
	"errors"
	"regexp"
	"strings"
)

var errInvalidReturnRenamerParams = errors.New("invalid return renamer params")

type ReturnRenamer struct {
	methodName string
	returnName string

	resultName string
}

func newReturnRenamer(k, val string) (ReturnRenamer, error) {
	paramsParser := regexp.MustCompile(`^([a-zA-Z0-9]+)\.([a-zA-Z0-9]+)$`)
	match := paramsParser.FindStringSubmatch(k)
	if len(match) == 0 {
		return ReturnRenamer{}, errInvalidReturnRenamerParams
	}

	return ReturnRenamer{
		methodName: match[1],
		returnName: match[2],
		resultName: val,
	}, nil
}

func (r ReturnRenamer) GetMethodName() string    { return r.methodName }
func (r ReturnRenamer) GetOldReturnName() string { return r.returnName }
func (r ReturnRenamer) GetNewReturnName() string { return r.resultName }

type Storage struct {
	renamer map[string]ReturnRenamer
}

func NewStorage(params map[string]string) (*Storage, error) {
	s := Storage{renamer: make(map[string]ReturnRenamer, len(params))}
	for k, v := range params {
		r, err := newReturnRenamer(k, v)
		if err != nil {
			return nil, err
		}
		s.renamer[strings.ToLower(r.GetMethodName())] = r
	}

	return &s, nil
}

func (s *Storage) GetReturnRenamer(methodName string) *ReturnRenamer {
	if r, ok := s.renamer[strings.ToLower(methodName)]; ok {
		return &r
	}

	return nil
}
