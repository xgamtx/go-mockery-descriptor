package parser_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/xgamtx/go-mockery-descriptor/internal/parser"
)

const embedsDir = "./testdata/embeds"

func methodNames(iface *parser.Interface) []string {
	names := make([]string, 0, len(iface.Methods))
	for _, m := range iface.Methods {
		names = append(names, m.Name)
	}

	return names
}

func TestParseInterfaceInDir_Embedded(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string

		interfaceName string

		wantMethods []string
	}{
		{
			name: "own methods keep source order and embedded ones are appended",

			interfaceName: "Transitive",

			// Own объявлен явно, остальные приходят из Middle -> Base и io.ReaderFrom.
			wantMethods: []string{"Own", "Mid", "Ping", "ReadFrom"},
		},
		{
			name: "transitive embedding",

			interfaceName: "Middle",

			wantMethods: []string{"Mid", "Ping"},
		},
		{
			name: "overlapping method sets are deduplicated",

			interfaceName: "Overlapping",

			wantMethods: []string{"Mid", "Ping"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := parser.ParseInterfaceInDir(embedsDir, tt.interfaceName)
			require.NoError(t, err)

			assert.Equal(t, tt.interfaceName, got.Name)
			assert.Equal(t, "embeds", got.PackageName)
			assert.Equal(t, tt.wantMethods, methodNames(got))
		})
	}
}

func TestParseInterfaceInDir_EmbeddedMethodSignatures(t *testing.T) {
	t.Parallel()

	iface, err := parser.ParseInterfaceInDir(embedsDir, "Transitive")
	require.NoError(t, err)

	byName := make(map[string]parser.Method, len(iface.Methods))
	for _, m := range iface.Methods {
		byName[m.Name] = m
	}

	// Тип из другого пакета должен остаться квалифицированным, а его импорт — попасть в PathTypes.
	readFrom := byName["ReadFrom"]
	require.Len(t, readFrom.Params, 1)
	assert.Equal(t, "r", readFrom.Params[0].Name)
	assert.Equal(t, "io.Reader", readFrom.Params[0].Type)
	assert.Equal(t, []string{"io"}, readFrom.Params[0].PathTypes)

	require.Len(t, readFrom.Returns, 2)
	assert.Equal(t, "n", readFrom.Returns[0].Name)
	assert.Equal(t, "int64", readFrom.Returns[0].Type)

	ping := byName["Ping"]
	require.Len(t, ping.Params, 1)
	assert.Equal(t, "context.Context", ping.Params[0].Type)
	assert.Equal(t, []string{"context"}, ping.Params[0].PathTypes)

	mid := byName["Mid"]
	require.Len(t, mid.Params, 1)
	assert.Equal(t, "[]string", mid.Params[0].Type)
	assert.Empty(t, mid.Params[0].PathTypes)

	require.Len(t, mid.Returns, 2)
	assert.Equal(t, "map[string]int", mid.Returns[0].Type)
	assert.Equal(t, "error", mid.Returns[1].Type)
}
