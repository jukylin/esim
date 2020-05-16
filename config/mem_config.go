package config

import (
	"time"

	"github.com/spf13/cast"
	"github.com/spf13/viper"
)

type MemConfig struct {
	data map[string]interface{}
}

func NewMemConfig() *MemConfig {
	return &MemConfig{
		make(map[string]interface{}),
	}
}

func (mc *MemConfig) Get(key string) interface{} {
	if val, ok := mc.data[key]; ok {
		return val
	}

	return nil
}

func (mc *MemConfig) GetString(key string) string {
	return cast.ToString(mc.Get(key))
}

func (mc *MemConfig) GetBool(key string) bool {
	return cast.ToBool(mc.Get(key))
}

func (mc *MemConfig) GetInt(key string) int {
	return cast.ToInt(mc.Get(key))
}

func (mc *MemConfig) GetInt32(key string) int32 {
	return cast.ToInt32(mc.Get(key))
}

func (mc *MemConfig) GetInt64(key string) int64 {
	return cast.ToInt64(mc.Get(key))
}

func (mc *MemConfig) GetUint(key string) uint {
	return cast.ToUint(mc.Get(key))
}

func (mc *MemConfig) GetUint32(key string) uint32 {
	return cast.ToUint32(mc.Get(key))
}

func (mc *MemConfig) GetUint64(key string) uint64 {
	return cast.ToUint64(mc.Get(key))
}

func (mc *MemConfig) GetFloat64(key string) float64 {
	return cast.ToFloat64(mc.Get(key))
}

func (mc *MemConfig) GetTime(key string) time.Time {
	return cast.ToTime(mc.Get(key))
}

func (mc *MemConfig) GetDuration(key string) time.Duration {
	return cast.ToDuration(mc.Get(key))
}

//func (mc *MemConfig) GetIntSlice(key string) []int { return viper.GetIntSlice(key) }

func (mc *MemConfig) GetStringSlice(key string) []string {
	return cast.ToStringSlice(mc.Get(key))
}

func (mc *MemConfig) GetStringMap(key string) map[string]interface{} {
	return cast.ToStringMap(mc.Get(key))
}

func (mc *MemConfig) GetStringMapString(key string) map[string]string {
	return cast.ToStringMapString(mc.Get(key))
}

func (mc *MemConfig) GetStringMapStringSlice(key string) map[string][]string {
	return cast.ToStringMapStringSlice(mc.Get(key))
}

//TODO 还没实现
func (mc *MemConfig) GetSizeInBytes(key string) uint {
	return 0
}

func (mc *MemConfig) UnmarshalKey(key string, rawVal interface{}, opts ...viper.DecoderConfigOption) error {
	return nil
}

func (mc *MemConfig) Unmarshal(rawVal interface{}, opts ...viper.DecoderConfigOption) error {
	return nil
}

func (mc *MemConfig) Set(key string, value interface{}) {
	mc.data[key] = value
}
