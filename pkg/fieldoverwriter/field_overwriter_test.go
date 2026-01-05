package fieldoverwriter //nolint:testpackage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NewFieldOverwriter(t *testing.T) {
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
				fieldName:  "X",
				funcPath:   "context",
				funcName:   "context.Context",
			},
		},
		{
			name:   "OK, external function",
			params: "SetX.X=github.com/xgamtx/go-mockery-descriptor/pkg/assessor.OneOf",

			want: &FieldOverwriter{
				methodName: "SetX",
				fieldName:  "X",
				funcPath:   "github.com/xgamtx/go-mockery-descriptor/pkg/assessor",
				funcName:   "assessor.OneOf",
			},
		},
		{
			name:   "OK, external function with version",
			params: "SetX.X=github.com/jackc/pgx/v5.Tx",

			want: &FieldOverwriter{
				methodName: "SetX",
				fieldName:  "X",
				funcPath:   "github.com/jackc/pgx/v5",
				funcName:   "pgx.Tx",
			},
		},
		{
			name:   "OK, internal function",
			params: "SetX.X=OneOf",

			want: &FieldOverwriter{
				methodName: "SetX",
				fieldName:  "X",
				funcPath:   "",
				funcName:   "OneOf",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := newFieldOverwriter(tt.params)
			assert.Equal(t, tt.want, got)
			if tt.wantErrMsg != "" {
				assert.Error(t, err, tt.wantErrMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
