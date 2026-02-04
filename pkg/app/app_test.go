package app_test

import (
	_ "embed"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/xgamtx/go-mockery-descriptor/pkg/app"
)

//go:embed some.gen_test.go
var expectedRes string

//go:generate mockery --name=Some --inpackage --with-expecter=true --structname=mockSome
func TestRun(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string

		dir                   string
		interfaceName         string
		fieldOverwriterParams []string
		fullPackagePath       string

		want       string
		wantErrMsg string
	}{
		{
			name: "success",

			dir:                   ".",
			interfaceName:         "Some",
			fieldOverwriterParams: []string{"Slice.rows=elementsMatch", "SetX.x=oneOf", "Anything.v=any"},
			fullPackagePath:       "github.com/xgamtx/go-mockery-descriptor",

			want: expectedRes,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := app.Run(tt.dir, tt.interfaceName, tt.fieldOverwriterParams, tt.fullPackagePath)
			assert.Equal(t, tt.want, got)
			if tt.wantErrMsg != "" {
				assert.Error(t, err, tt.wantErrMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
