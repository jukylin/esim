package config

import (
	"time"

	"github.com/spf13/viper"
)

type Config interface {
	Get(key string) interface{}

	GetString(key string) string

	GetBool(key string) bool

	GetInt(key string) int

	GetInt32(key string) int32

	GetInt64(key string) int64

	GetUint(key string) uint

	GetUint32(key string) uint32

	GetUint64(key string) uint64

	GetFloat64(key string) float64

	GetTime(key string) time.Time

	GetDuration(key string) time.Duration

	// GetIntSlice(key string) []int { return viper.GetIntSlice(key) }

	GetStringSlice(key string) []string

	GetStringMap(key string) map[string]interface{}

	GetStringMapString(key string) map[string]string

	GetStringMapStringSlice(key string) map[string][]string

	GetSizeInBytes(key string) uint

	UnmarshalKey(key string, rawVal interface{}, opts ...viper.DecoderConfigOption) error

	Unmarshal(rawVal interface{}, opts ...viper.DecoderConfigOption) error

	Set(key string, value interface{})
}
