package config

import (
	"time"
	"github.com/spf13/viper"
	"github.com/spf13/cast"
)

type MemConfig struct{
	data map[string]interface{}
}


func NewMemConfig() *MemConfig {
	return &MemConfig{
		make(map[string]interface{}),
		}
}


func (this *MemConfig)Get(key string) interface{} {
	if val, ok := this.data[key]; ok {
		return val
	}else{
		return nil
	}
}


func (this *MemConfig) GetString(key string) string {
	return cast.ToString(this.Get(key))
}


func (this *MemConfig) GetBool(key string) bool {
	return cast.ToBool(this.Get(key))
}


func (this *MemConfig) GetInt(key string) int {
	return cast.ToInt(this.Get(key))
}


func (this *MemConfig) GetInt32(key string) int32 {
	return cast.ToInt32(this.Get(key))
}


func (this *MemConfig) GetInt64(key string) int64 {
	return cast.ToInt64(this.Get(key))
}


func (this *MemConfig) GetUint(key string) uint {
	return cast.ToUint(this.Get(key))
}


func (this *MemConfig) GetUint32(key string) uint32 {
	return cast.ToUint32(this.Get(key))
}


func (this *MemConfig) GetUint64(key string) uint64 {
	return cast.ToUint64(this.Get(key))
}


func (this *MemConfig) GetFloat64(key string) float64 {
	return cast.ToFloat64(this.Get(key))
}


func (this *MemConfig) GetTime(key string) time.Time {
	return cast.ToTime(this.Get(key))
}


func (this *MemConfig) GetDuration(key string) time.Duration {
	return cast.ToDuration(this.Get(key))
}


//func (this *MemConfig) GetIntSlice(key string) []int { return viper.GetIntSlice(key) }


func (this *MemConfig) GetStringSlice(key string) []string {
	return cast.ToStringSlice(this.Get(key))
}


func (this *MemConfig) GetStringMap(key string) map[string]interface{} {
	return cast.ToStringMap(this.Get(key))
}


func (this *MemConfig) GetStringMapString(key string) map[string]string {
	return cast.ToStringMapString(this.Get(key))
}


func (this *MemConfig) GetStringMapStringSlice(key string) map[string][]string {
	return cast.ToStringMapStringSlice(this.Get(key))
}

//TODO 还没实现
func (this *MemConfig) GetSizeInBytes(key string) uint {
	return 0
}


func (this *MemConfig) UnmarshalKey(key string, rawVal interface{}, opts ...viper.DecoderConfigOption) error {
	return nil
}


func (this *MemConfig) Unmarshal(rawVal interface{}, opts ...viper.DecoderConfigOption) error {
	return nil
}

func (this *MemConfig) Set(key string, value interface{}){
	this.data[key] = value
}
