package config

import (
	"github.com/spf13/cast"
	"github.com/spf13/viper"
	"time"
)

type NullConfig struct {
	data map[string]interface{}
}

func NewNullConfig() *NullConfig {
	return &NullConfig{}
}

func (nc *NullConfig) Get(key string) interface{} {
	return nil
}

func (nc *NullConfig) GetString(key string) string {
	return ""
}

func (nc *NullConfig) GetBool(key string) bool {
	return false
}

func (nc *NullConfig) GetInt(key string) int {
	return 0
}

func (nc *NullConfig) GetInt32(key string) int32 {
	return 0
}

func (nc *NullConfig) GetInt64(key string) int64 {
	return 0
}

func (nc *NullConfig) GetUint(key string) uint {
	return 0
}

func (nc *NullConfig) GetUint32(key string) uint32 {
	return 0
}

func (nc *NullConfig) GetUint64(key string) uint64 {
	return 0
}

func (nc *NullConfig) GetFloat64(key string) float64 {
	return 0
}

func (nc *NullConfig) GetTime(key string) time.Time {
	return time.Time{}
}

func (nc *NullConfig) GetDuration(key string) time.Duration {
	return cast.ToDuration(0)
}

//func (nc *NullConfig) GetIntSlice(key string) []int { return viper.GetIntSlice(key) }

func (nc *NullConfig) GetStringSlice(key string) []string {
	return []string{}
}

func (nc *NullConfig) GetStringMap(key string) map[string]interface{} {
	return map[string]interface{}{}
}

func (nc *NullConfig) GetStringMapString(key string) map[string]string {
	return map[string]string{}
}

func (nc *NullConfig) GetStringMapStringSlice(key string) map[string][]string {
	return map[string][]string{}
}

//TODO 还没实现
func (nc *NullConfig) GetSizeInBytes(key string) uint {
	return 0
}

func (nc *NullConfig) UnmarshalKey(key string, rawVal interface{}, opts ...viper.DecoderConfigOption) error {
	return nil
}

func (nc *NullConfig) Unmarshal(rawVal interface{}, opts ...viper.DecoderConfigOption) error {
	return nil
}

func (nc *NullConfig) Set(key string, value interface{}) {}
