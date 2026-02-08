package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Config struct {
	Dir                   string            `mapstructure:"dir"`
	Interface             string            `mapstructure:"interface"`
	Output                string            `mapstructure:"output"`
	FieldOverwriterParams []string          `mapstructure:"field-overwriter-param"`
	RenameReturns         map[string]string `mapstructure:"rename-returns"`
	ConstructorName       string            `mapstructure:"constructor-name"`
}

func initFlags() {
	pflag.String("dir", "", "output directory")
	pflag.String("interface", "", "interface name")
	pflag.String("output", "", "output file")
	pflag.StringSlice("field-overwriter-param", nil, "field overwriter param, can be used more than once")

	pflag.Parse()
}

func initDefaults() {
	viper.SetDefault("dir", ".")
	viper.SetDefault("constructor-name", "newMock{{ . }}")
}

func New() (*Config, error) {
	initFlags()
	initDefaults()
	viper.SetOptions(viper.KeyDelimiter("::"))
	viper.SetConfigType("yaml")

	dirs, err := getConfigPaths()
	if err != nil {
		return nil, err
	}

	for i := len(dirs) - 1; i >= 0; i-- {
		dir := dirs[i]
		candidate := filepath.Join(dir, ".mockery-descriptor.yaml")
		if _, err = os.Stat(candidate); os.IsNotExist(err) {
			continue
		}

		viper.SetConfigFile(candidate)
		if err = viper.MergeInConfig(); err != nil {
			return nil, err
		}
	}

	if err = viper.BindPFlags(pflag.CommandLine); err != nil {
		return nil, err
	}

	var cfg Config
	if err = viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func getConfigPaths() ([]string, error) {
	var paths []string

	current, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	paths = append(paths, current)
	for {
		modPath := filepath.Join(current, "go.mod")
		if _, err = os.Stat(modPath); err == nil {
			break
		}

		parent := filepath.Dir(current)
		if parent == current {
			break
		}

		paths = append(paths, parent)
		current = parent
	}

	return paths, nil
}
