package assessor

import (
	"reflect"

	"github.com/stretchr/testify/mock"
)

type Matcher interface {
	Matches(argument any) bool
}

func ElementsMatch[T any](expected []T) Matcher {
	return mock.MatchedBy(func(actual []T) bool {
		if len(actual) != len(expected) {
			return false
		}

		found := make([]bool, len(actual))
		for _, actualItem := range actual {
			var foundItem bool
			for i, expectedItem := range expected {
				if found[i] {
					continue
				}

				if reflect.DeepEqual(actualItem, expectedItem) {
					found[i] = true
					foundItem = true

					break
				}
			}
			if !foundItem {
				return false
			}
		}

		return true
	})
}

func OneOf[T any](expected []T) Matcher {
	return mock.MatchedBy(func(actual T) bool {
		for _, expectedItem := range expected {
			if reflect.DeepEqual(actual, expectedItem) {
				return true
			}
		}

		return false
	})
}
