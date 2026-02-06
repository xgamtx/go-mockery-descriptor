package fieldoverwriter //nolint:testpackage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_NewFieldOverwriter(t *testing.T) { //nolint:funlen
	t.Parallel()

	tests := []struct {
		name string

		params string

		want       *FieldOverwriter
		wantErrMsg string
	}{
		{
			name: "empty params",

			wantErrMsg: errInvalidFieldOverwriterParams.Error(),
		},
		{
			name:   "OK, external standard function",
			params: "SetX.X=context.Context",

			want: &FieldOverwriter{
				methodName: "SetX",
				fieldName:  Link("X"),
				funcPath:   "context",
				funcName:   "context.Context",
			},
		},
		{
			name:   "OK, external function",
			params: "SetX.X=github.com/xgamtx/go-mockery-descriptor/pkg/assessor.OneOf",

			want: &FieldOverwriter{
				methodName: "SetX",
				fieldName:  Link("X"),
				funcPath:   "github.com/xgamtx/go-mockery-descriptor/pkg/assessor",
				funcName:   "assessor.OneOf",
			},
		},
		{
			name:   "OK, external function with version",
			params: "SetX.X=github.com/jackc/pgx/v5.Tx",

			want: &FieldOverwriter{
				methodName: "SetX",
				fieldName:  Link("X"),
				funcPath:   "github.com/jackc/pgx/v5",
				funcName:   "pgx.Tx",
			},
		},
		{
			name:   "OK, internal function",
			params: "SetX.X=OneOf",

			want: &FieldOverwriter{
				methodName: "SetX",
				fieldName:  Link("X"),
				funcPath:   "",
				funcName:   "OneOf",
			},
		},
		{
			name:   "OK, standard function",
			params: "SetX.X=oneOf",

			want: &FieldOverwriter{
				methodName: "SetX",
				fieldName:  Link("X"),
				funcPath:   "github.com/xgamtx/go-mockery-descriptor/pkg/assessor",
				funcName:   "assessor.OneOf",
			},
		},
		{
			name:   "OK, with param index",
			params: "SetX.0=oneOf",

			want: &FieldOverwriter{
				methodName: "SetX",
				fieldIndex: Link(0),
				funcPath:   "github.com/xgamtx/go-mockery-descriptor/pkg/assessor",
				funcName:   "assessor.OneOf",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := newFieldOverwriter(tt.params)
			if got != nil {
				got.typeModifier = nil // validate field in another test
			}
			assert.Equal(t, tt.want, got)
			if tt.wantErrMsg != "" {
				assert.Error(t, err, tt.wantErrMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func Test_NewFieldOverwriter_typeModifier(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string

		params       string
		originalType string

		wantType string
	}{
		{
			name: "OK, external function",

			params:       "SetX.X=github.com/xgamtx/go-mockery-descriptor/pkg/assessor.OneOf",
			originalType: "bool",

			wantType: "bool",
		},
		{
			name: "OK, oneOf function",

			params:       "SetX.X=oneOf",
			originalType: "bool",

			wantType: "[]bool",
		},
		{
			name: "OK, elementsMatch function",

			params:       "SetX.X=elementsMatch",
			originalType: "[]bool",

			wantType: "[]bool",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := newFieldOverwriter(tt.params)
			require.NotNil(t, got)
			require.NoError(t, err)

			assert.Equal(t, tt.wantType, got.typeModifier(tt.originalType))
		})
	}
}

func Link[T any](val T) *T { return &val }
