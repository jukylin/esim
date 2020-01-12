package config

import (
	"github.com/spf13/viper"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"
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

func (this *viperConf) Get(key string) interface{} { return this.Viper.Get(key) }

func (this *viperConf) GetString(key string) string { return this.Viper.GetString(key) }

func (this *viperConf) GetBool(key string) bool { return this.Viper.GetBool(key) }

func (this *viperConf) GetInt(key string) int { return this.Viper.GetInt(key) }

func (this *viperConf) GetInt32(key string) int32 { return this.Viper.GetInt32(key) }

func (this *viperConf) GetInt64(key string) int64 { return this.Viper.GetInt64(key) }

func (this *viperConf) GetUint(key string) uint { return this.Viper.GetUint(key) }

func (this *viperConf) GetUint32(key string) uint32 { return this.Viper.GetUint32(key) }

func (this *viperConf) GetUint64(key string) uint64 { return this.Viper.GetUint64(key) }

func (this *viperConf) GetFloat64(key string) float64 { return this.Viper.GetFloat64(key) }

func (this *viperConf) GetTime(key string) time.Time { return this.Viper.GetTime(key) }

func (this *viperConf) GetDuration(key string) time.Duration { return this.Viper.GetDuration(key) }

//func GetIntSlice(key string) []int { return config.GetIntSlice(key) }

func (this *viperConf) GetStringSlice(key string) []string { return this.Viper.GetStringSlice(key) }

func (this *viperConf) GetStringMap(key string) map[string]interface{} {
	return this.Viper.GetStringMap(key)
}

func (this *viperConf) GetStringMapString(key string) map[string]string {
	return this.Viper.GetStringMapString(key)
}

func (this *viperConf) GetStringMapStringSlice(key string) map[string][]string {
	return this.Viper.GetStringMapStringSlice(key)
}

func (this *viperConf) GetSizeInBytes(key string) uint { return this.Viper.GetSizeInBytes(key) }

func (this *viperConf) UnmarshalKey(key string, rawVal interface{}, opts ...viper.DecoderConfigOption) error {
	return this.Viper.UnmarshalKey(key, rawVal, opts...)
}

func (this *viperConf) Unmarshal(rawVal interface{}, opts ...viper.DecoderConfigOption) error {
	return this.Viper.Unmarshal(rawVal, opts...)
}

func (this *viperConf) Set(key string, value interface{}) { this.Viper.Set(key, value) }
