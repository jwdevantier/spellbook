package utils

import (
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

const configName = ".spellbook"

func PrettyPrintConfig() (string, error) {
	bs, err := yaml.Marshal(viper.AllSettings())
	if err != nil {
		return "", err
	}
	return string(bs), nil
}

type Command struct {
	Cmd string
	Desc string
}

type Config struct {
	Commands []Command
}

func (c *Config) merge(other *Config) error {
	c.Commands = append(c.Commands, other.Commands...)
	return nil
}

func readRawConfig(confDir string) *viper.Viper {
	conf := viper.New()
	conf.AddConfigPath(confDir)
	conf.SetConfigName(configName)
	if err := conf.ReadInConfig(); err == nil {
		return conf
	}
	return nil
}

func readRawConfigs(configDirs []string) []*viper.Viper {
	// Accept that not every supplied dir has a config to read
	// Ergo len(configDirs) >= len(res)
	res := make([]*viper.Viper, 0, len(configDirs))
	for _, configDir := range configDirs {
		curr := readRawConfig(configDir)
		if curr != nil {
			res = append(res, curr)
		}
	}
	return res
}

type ConfigError struct {
	ConfigFile string
	Message string
	Cause error
}

func (c *ConfigError) Error() string {
	return fmt.Sprintf("%s: %s", c.ConfigFile, c.Message)
}

func unmarshalRawConfigs(rawConfigs []*viper.Viper) ([]*Config, error) {
	// expect len(rawConfigs) == len(res)
	res := make([]*Config, 0, len(rawConfigs))
	for _, rawConfig := range rawConfigs {
		var conf Config
		// unmarshal
		err := rawConfig.Unmarshal(&conf)
		if err != nil {
			return nil, &ConfigError{
				rawConfig.ConfigFileUsed(),
				"cannot unmarshal",
				err}
		}
		res = append(res, &conf)
	}
	return res, nil
}

type Mergeable interface {
	merge(val Mergeable) interface{}
}

func ReadConfig(configDirs []string) (*Config, error) {
	// read in all available configs
	rawConfigs := readRawConfigs(configDirs)
	if len(rawConfigs) == 0 {
		return nil, errors.New("no configs to read")
	}

	// unmarshal each config into its own Config instance
	configs, err := unmarshalRawConfigs(rawConfigs)
	if err != nil {
		return nil, err
	}
	if len(configs) == 1 {
		// single config, no merging needed
		return configs[0], nil
	}

	// merge all configs into one
	conf := Config{}
	for _, otherConf := range configs {
		err = conf.merge(otherConf)
		if err != nil {
			// TODO: should probably detail which configs we failed to merge
			return nil, err
		}
	}
	return &conf, nil
}
