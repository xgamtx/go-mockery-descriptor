// The test lives in package app: the generated constructor and call structs are unexported.
package app //nolint:testpackage

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMakeHandlerMock checks that the constructor of a function type returns a value that is
// callable as the function itself rather than the mock struct.
func TestMakeHandlerMock(t *testing.T) {
	t.Parallel()

	handler := makeHandlerMock(t, []handlerCall{
		{Id: "id-1", ReceivedResult: "first"},
		{Id: "id-2", ReceivedResult: "second"},
	})

	got, err := handler(context.Background(), "id-1")
	require.NoError(t, err)
	assert.Equal(t, "first", got)

	got, err = handler(context.Background(), "id-2")
	require.NoError(t, err)
	assert.Equal(t, "second", got)
}

// TestMakeTickerMock covers a function type without parameters and return values.
func TestMakeTickerMock(t *testing.T) {
	t.Parallel()

	tick := makeTickerMock(t, []tickerCall{{}, {}})

	tick()
	tick()
}

// TestMakeFilterMock checks that param matchers replaced through field-overwriter-param are
// actually applied when the function type is called.
func TestMakeFilterMock(t *testing.T) {
	t.Parallel()

	errFilter := errors.New("filter failed")

	filter := makeFilterMock(t, []filterCall{
		{Rows: []string{"a", "b"}, Mode: []int{1, 2}, ReceivedErr: errFilter},
	})

	// elementsMatch ignores the order of elements, oneOf accepts any of the listed values and
	// any skips the comparison of the parameter entirely.
	require.ErrorIs(t, filter([]string{"b", "a"}, 2, true), errFilter)
}
