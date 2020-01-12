package config

import (
	"time"
	"github.com/spf13/viper"
	"github.com/spf13/cast"
)

type NullConfig struct{
	data map[string]interface{}
}


func NewNullConfig() *NullConfig {
	return &NullConfig{}
}


func (this *NullConfig) Get(key string) interface{} {
	return nil
}


func (this *NullConfig) GetString(key string) string {
	return ""
}


func (this *NullConfig) GetBool(key string) bool {
	return false
}


func (this *NullConfig) GetInt(key string) int {
	return 0
}


func (this *NullConfig) GetInt32(key string) int32 {
	return 0
}


func (this *NullConfig) GetInt64(key string) int64 {
	return 0
}


func (this *NullConfig) GetUint(key string) uint {
	return 0
}


func (this *NullConfig) GetUint32(key string) uint32 {
	return 0
}


func (this *NullConfig) GetUint64(key string) uint64 {
	return 0
}


func (this *NullConfig) GetFloat64(key string) float64 {
	return 0
}


func (this *NullConfig) GetTime(key string) time.Time {
	return time.Time{}
}


func (this *NullConfig) GetDuration(key string) time.Duration {
	return cast.ToDuration(0)
}


//func (this *NullConfig) GetIntSlice(key string) []int { return viper.GetIntSlice(key) }


func (this *NullConfig) GetStringSlice(key string) []string {
	return []string{}
}


func (this *NullConfig) GetStringMap(key string) map[string]interface{} {
	return map[string]interface{}{}
}


func (this *NullConfig) GetStringMapString(key string) map[string]string {
	return map[string]string{}
}


func (this *NullConfig) GetStringMapStringSlice(key string) map[string][]string {
	return map[string][]string{}
}

//TODO 还没实现
func (this *NullConfig) GetSizeInBytes(key string) uint {
	return 0
}


func (this *NullConfig) UnmarshalKey(key string, rawVal interface{}, opts ...viper.DecoderConfigOption) error {
	return nil
}


func (this *NullConfig) Unmarshal(rawVal interface{}, opts ...viper.DecoderConfigOption) error {
	return nil
}

func (this *NullConfig) Set(key string, value interface{}){}
