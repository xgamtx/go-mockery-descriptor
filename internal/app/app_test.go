package app_test

import (
	_ "embed"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/xgamtx/go-mockery-descriptor/internal/app"
	"github.com/xgamtx/go-mockery-descriptor/internal/config"
)

//go:embed some.gen_test.go
var expectedRes string

//go:generate mockery --name=Some --inpackage --with-expecter=true --structname=mockSome
func TestRun(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string

		cfg *config.Config

		want       string
		wantErrMsg string
	}{
		{
			name: "success",

			cfg: &config.Config{
				Interface:             "Some",
				FieldOverwriterParams: []string{"Slice.rows=elementsMatch", "SetX.x=oneOf", "Anything.v=any"},
				RenameReturns:         map[string]string{"GetX.r0": "X"},
			},

			want: expectedRes,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := app.Run(tt.cfg)
			assert.Equal(t, tt.want, got)
			if tt.wantErrMsg != "" {
				assert.Error(t, err, tt.wantErrMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
