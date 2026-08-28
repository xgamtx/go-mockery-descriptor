package app_test

import (
	_ "embed"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/xgamtx/go-mockery-descriptor/internal/app"
	"github.com/xgamtx/go-mockery-descriptor/internal/config"
)

//go:embed some.gen_test.go
var expectedRes string

//go:embed embedded.gen_test.go
var expectedEmbeddedRes string

//go:embed handler.gen_test.go
var expectedHandlerRes string

//go:embed ticker.gen_test.go
var expectedTickerRes string

//go:embed filter.gen_test.go
var expectedFilterRes string

//go:generate mockery --name=Some --inpackage --with-expecter=true --structname=mockSome
//go:generate mockery --name=Embedded --inpackage --with-expecter=true --structname=mockEmbedded
func TestRun(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string

		cfg *config.InterfaceConfig

		want       string
		wantErrMsg string
	}{
		{
			name: "success",

			cfg: &config.InterfaceConfig{
				Name:                  "Some",
				ConstructorName:       "newMock{{ . }}",
				PackageName:           "{{ . }}",
				FieldOverwriterParams: []string{"Slice.rows=elementsMatch", "SetX.x=oneOf", "Anything.v=any"},
				RenameReturns: map[string]string{
					"GetX.r0":  "X",
					"Multi.r0": "X",
					"Multi.r1": "Y",
				},
			},

			want: expectedRes,
		},
		{
			name: "embedded interfaces",

			cfg: &config.InterfaceConfig{
				Name:            "Embedded",
				ConstructorName: "newMock{{ . }}",
				PackageName:     "{{ . }}",
			},

			want: expectedEmbeddedRes,
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

//go:generate mockery --name=Handler --inpackage --with-expecter=true --structname=mockHandler
//go:generate mockery --name=Ticker --inpackage --with-expecter=true --structname=mockTicker
//go:generate mockery --name=Filter --inpackage --with-expecter=true --structname=mockFilter
func TestRunFuncTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string

		cfg *config.InterfaceConfig

		want string
	}{
		{
			name: "function type",

			cfg: &config.InterfaceConfig{
				Name:            "Handler",
				ConstructorName: "newMock{{ . }}",
				PackageName:     "{{ . }}",
				RenameReturns:   map[string]string{"Handler.r0": "Result"},
			},

			want: expectedHandlerRes,
		},
		{
			name: "function type without params and returns",

			cfg: &config.InterfaceConfig{
				Name:            "Ticker",
				ConstructorName: "newMock{{ . }}",
				PackageName:     "{{ . }}",
			},

			want: expectedTickerRes,
		},
		{
			name: "function type with overwritten param matchers",

			cfg: &config.InterfaceConfig{
				Name:            "Filter",
				ConstructorName: "newMock{{ . }}",
				PackageName:     "{{ . }}",
				FieldOverwriterParams: []string{
					"Filter.rows=elementsMatch",
					"Filter.mode=oneOf",
					"Filter.ignored=any",
				},
			},

			want: expectedFilterRes,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := app.Run(tt.cfg)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
