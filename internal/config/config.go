package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Config struct {
	Dir             string `mapstructure:"dir"`
	Output          string `mapstructure:"output"`
	ConstructorName string `mapstructure:"constructor-name"`
	PackageName     string `mapstructure:"package-name"`
	Interfaces      []InterfaceConfig
}

type InterfaceConfig struct {
	Dir             string `mapstructure:"dir"`
	Output          string `mapstructure:"output"`
	ConstructorName string `mapstructure:"constructor-name"`
	PackageName     string `mapstructure:"package-name"`

	Name                  string            `mapstructure:"name"`
	FieldOverwriterParams []string          `mapstructure:"field-overwriter-param"`
	RenameReturns         map[string]string `mapstructure:"rename-returns"`
}

func (cfg *Config) Init() {
	for i := range cfg.Interfaces {
		if cfg.Interfaces[i].Dir == "" {
			cfg.Interfaces[i].Dir = cfg.Dir
		}
		if cfg.Interfaces[i].Output == "" {
			cfg.Interfaces[i].Output = cfg.Output
		}
		if cfg.Interfaces[i].ConstructorName == "" {
			cfg.Interfaces[i].ConstructorName = cfg.ConstructorName
		}
		if cfg.Interfaces[i].PackageName == "" {
			cfg.Interfaces[i].PackageName = cfg.PackageName
		}
	}
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
	viper.SetDefault("output", "{{ . }}.mockery-helper_test.go")
	viper.SetDefault("package-name", "{{ . }}_test")
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

	cfg.Init()

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
