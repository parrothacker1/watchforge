package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Root  string `mapstructure:"root"`
	Build struct {
		Command string `mapstructure:"command"`
	} `mapstructure:"build"`
	Run struct {
		Command string `mapstructure:"command"`
	} `mapstructure:"run"`
	Watcher struct {
		Debounce  int      `mapstructure:"debounce"`
		IgnoreGit bool     `mapstructure:"ignore-git"`
		Ignore    []string `mapstructure:"ignore"`
		Exclude   []string `mapstructure:"exclude"`
	} `mapstructure:"watcher"`
	Actions struct {
		Build   []string `mapstructure:"build"`
		Restart []string `mapstructure:"restart"`
	} `mapstructure:"actions"`
	Log struct {
		Debug bool `mapstructure:"debug"`
	} `mapstructure:"log"`
	Runner struct {
		CrashWindow  int `mapstructure:"crash-window"`
		MaxCrashes   int `mapstructure:"max-crashes"`
		RestartDelay int `mapstructure:"restart-delay"`
	} `mapstructure:"runner"`
}

func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigName("watchforge")
	v.SetConfigType("toml")
	v.AddConfigPath(path)

	v.SetDefault("root", ".")
	v.SetDefault("watcher.debounce", 200)
	v.SetDefault("watcher.ignore-git", true)
	v.SetDefault("watcher.ignore", []string{})
	v.SetDefault("watcher.exclude", []string{})
	v.SetDefault("actions.build", []string{".go"})
	v.SetDefault("actions.restart", []string{".env", ".yaml", ".yml", ".json", ".md"})
	v.SetDefault("log.debug", false)
	v.SetDefault("runner.crash-window", 2000)
	v.SetDefault("runner.max-crashes", 3)
	v.SetDefault("runner.restart-delay", 2000)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func Default() *Config {
	cfg := &Config{}
	cfg.Root = "."
	cfg.Build.Command = "go build -o ./bin/server ./cmd/server"
	cfg.Run.Command = "./bin/server"
	cfg.Watcher.Debounce = 200
	cfg.Watcher.IgnoreGit = true
	cfg.Watcher.Ignore = []string{
		"node_modules",
		"bin",
	}
	cfg.Watcher.Exclude = []string{}
	cfg.Log.Debug = false
	cfg.Actions.Build = []string{".go"}

	cfg.Actions.Restart = []string{
		".env",
		".yaml",
		".yml",
		".json",
		".md",
	}
	return cfg
}

func Write(path string, cfg *Config) error {
	v := viper.New()
	v.Set("root", cfg.Root)
	v.Set("build.command", cfg.Build.Command)
	v.Set("run.command", cfg.Run.Command)
	v.Set("watcher.debounce", cfg.Watcher.Debounce)
	v.Set("watcher.ignore-git", cfg.Watcher.IgnoreGit)
	v.Set("watcher.ignore", cfg.Watcher.Ignore)
	v.Set("watcher.exclude", cfg.Watcher.Exclude)
	v.Set("log.debug", cfg.Log.Debug)
	v.Set("actions.build", cfg.Actions.Build)
	v.Set("actions.restart", cfg.Actions.Restart)
	v.SetDefault("runner.crash-window", 2000)
	v.SetDefault("runner.max-crashes", 3)
	v.SetDefault("runner.restart-delay", 2000)
	v.SetConfigType("toml")

	return v.WriteConfigAs(path)
}
