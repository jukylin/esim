package config

import (
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type viperConf struct {
	*viper.Viper

	configType string

	configFile []string
}

type ViperConfOptions struct{}

type Option func(c *viperConf)

func NewViperConfig(options ...Option) Config {
	viperConf := &viperConf{}

	for _, option := range options {
		option(viperConf)
	}

	if viperConf.configType == "" {
		viperConf.configType = "yaml"
	}

	v := viper.New()
	v.SetConfigType(viperConf.configType)

	for k, configFile := range viperConf.configFile {

		if k == 0 {
			content, err := ioutil.ReadFile(configFile)
			if err != nil {
				log.Panicf("Fatal error config file: %s \n", err.Error())
			}

			err = v.ReadConfig(strings.NewReader(os.ExpandEnv(string(content))))
			if err != nil { // Handle errors reading the config file
				log.Panicf("Fatal error config file: %s \n", err.Error())
			}
		}

		if k > 0 {
			content, err := ioutil.ReadFile(configFile)
			if err != nil {
				log.Panicf("Fatal error config file: %s \n", err.Error())
			}

			err = v.MergeConfig(strings.NewReader(os.ExpandEnv(string(content))))
			if err != nil { // Handle errors reading the config file
				log.Panicf("Fatal error config file: %s \n", err.Error())
			}
		}
	}
	viperConf.Viper = v

	return viperConf
}

func (ViperConfOptions) WithConfigType(configType string) Option {
	return func(v *viperConf) {
		v.configType = configType
	}
}

func (ViperConfOptions) WithConfFile(configFile []string) Option {
	return func(l *viperConf) {
		l.configFile = configFile
	}
}

func (vc *viperConf) Get(key string) interface{} { return vc.Viper.Get(key) }

func (vc *viperConf) GetString(key string) string { return vc.Viper.GetString(key) }

func (vc *viperConf) GetBool(key string) bool { return vc.Viper.GetBool(key) }

func (vc *viperConf) GetInt(key string) int { return vc.Viper.GetInt(key) }

func (vc *viperConf) GetInt32(key string) int32 { return vc.Viper.GetInt32(key) }

func (vc *viperConf) GetInt64(key string) int64 { return vc.Viper.GetInt64(key) }

func (vc *viperConf) GetUint(key string) uint { return vc.Viper.GetUint(key) }

func (vc *viperConf) GetUint32(key string) uint32 { return vc.Viper.GetUint32(key) }

func (vc *viperConf) GetUint64(key string) uint64 { return vc.Viper.GetUint64(key) }

func (vc *viperConf) GetFloat64(key string) float64 { return vc.Viper.GetFloat64(key) }

func (vc *viperConf) GetTime(key string) time.Time { return vc.Viper.GetTime(key) }

func (vc *viperConf) GetDuration(key string) time.Duration { return vc.Viper.GetDuration(key) }

// func GetIntSlice(key string) []int { return config.GetIntSlice(key) }

func (vc *viperConf) GetStringSlice(key string) []string { return vc.Viper.GetStringSlice(key) }

func (vc *viperConf) GetStringMap(key string) map[string]interface{} {
	return vc.Viper.GetStringMap(key)
}

func (vc *viperConf) GetStringMapString(key string) map[string]string {
	return vc.Viper.GetStringMapString(key)
}

func (vc *viperConf) GetStringMapStringSlice(key string) map[string][]string {
	return vc.Viper.GetStringMapStringSlice(key)
}

func (vc *viperConf) GetSizeInBytes(key string) uint { return vc.Viper.GetSizeInBytes(key) }

func (vc *viperConf) UnmarshalKey(key string, rawVal interface{},
	opts ...viper.DecoderConfigOption) error {
	return vc.Viper.UnmarshalKey(key, rawVal, opts...)
}

func (vc *viperConf) Unmarshal(rawVal interface{}, opts ...viper.DecoderConfigOption) error {
	return vc.Viper.Unmarshal(rawVal, opts...)
}

func (vc *viperConf) Set(key string, value interface{}) { vc.Viper.Set(key, value) }
