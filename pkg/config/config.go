package config

import (
	"strings"

	"github.com/spf13/viper"
)

type Config interface {
	GetString(key string) string
	GetInt(key string) int
	GetBool(key string) bool
	GetStringSlice(key string) []string
}

type config struct {
	cfg *viper.Viper
}

func New() Config {
	c := new(config)
	c.mustLoad()

	return c
}

func (c *config) mustLoad() {
	cfg := viper.New()
	cfg.SetConfigName("config")
	cfg.SetConfigType("yaml")
	cfg.AddConfigPath("./config")

	err := cfg.ReadInConfig()
	if err != nil {
		panic(err)
	}

	cfg.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	cfg.AutomaticEnv()
	cfg.WatchConfig()

	c.cfg = cfg
}

func (c *config) GetString(key string) string {
	return c.cfg.GetString(key)
}

func (c *config) GetInt(key string) int {
	return c.cfg.GetInt(key)
}

func (c *config) GetBool(key string) bool {
	return c.cfg.GetBool(key)
}

func (c *config) GetStringSlice(key string) []string {
	return c.cfg.GetStringSlice(key)
}
