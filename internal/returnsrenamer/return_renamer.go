package returnsrenamer

import (
	"errors"
	"regexp"
	"strings"
)

var errInvalidReturnRenamerParams = errors.New("invalid return renamer params")

type ReturnRenamer struct {
	nameAliases map[string]string
}

func (r *ReturnRenamer) Append(oldName, newName string) {
	if r.nameAliases == nil {
		r.nameAliases = make(map[string]string)
	}
	r.nameAliases[oldName] = newName
}

func (r *ReturnRenamer) GetNewReturnName(oldName string) *string {
	if r == nil {
		return nil
	}
	newName, ok := r.nameAliases[oldName]
	if !ok {
		return nil
	}

	return &newName
}

type Storage struct {
	renamer map[string]ReturnRenamer
}

func NewStorage(params map[string]string) (*Storage, error) {
	s := Storage{renamer: make(map[string]ReturnRenamer)}
	for k, v := range params {
		methodName, oldVal, newVal, err := s.parse(k, v)
		if err != nil {
			return nil, err
		}

		key := s.getKey(methodName)
		renamer := s.renamer[key]
		renamer.Append(oldVal, newVal)
		s.renamer[key] = renamer
	}

	return &s, nil
}

func (s *Storage) GetReturnRenamer(methodName string) *ReturnRenamer {
	if r, ok := s.renamer[strings.ToLower(methodName)]; ok {
		return &r
	}

	return nil
}

func (s *Storage) getKey(methodName string) string { return strings.ToLower(methodName) }

func (s *Storage) parse(k, v string) (methodName, oldVal, newVal string, err error) { //nolint:nonamedreturns
	paramsParser := regexp.MustCompile(`^([a-zA-Z0-9]+)\.([a-zA-Z0-9]+)$`)
	match := paramsParser.FindStringSubmatch(k)
	if len(match) == 0 {
		return "", "", "", errInvalidReturnRenamerParams
	}

	return match[1], match[2], v, nil
}
