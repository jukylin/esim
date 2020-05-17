package config

import (
	"testing"
	"time"
)

func TestGet(t *testing.T) {
	memConfig := NewMemConfig()

	memConfig.Set("test", "config")

	res := memConfig.Get("test")
	if res.(string) != "config" {
		t.Errorf("结果错误 应该 config 实际 %s", res.(string))
	}
}

func TestGetString(t *testing.T) {
	memConfig := NewMemConfig()

	memConfig.Set("test", "config")

	res := memConfig.GetString("test")
	if res != "config" {
		t.Errorf("结果错误 应该 config 实际 %s", res)
	}
}

func TestGetBool(t *testing.T) {
	memConfig := NewMemConfig()

	memConfig.Set("test", true)

	res := memConfig.GetBool("test")
	if res != true {
		t.Errorf("结果错误 应该 true 实际 false")
	}
}

func TestGetInt(t *testing.T) {
	memConfig := NewMemConfig()

	memConfig.Set("test", 100)

	res := memConfig.GetInt("test")
	if res != 100 {
		t.Errorf("结果错误 应该 100 实际 %d", res)
	}
}

func TestGetInt32(t *testing.T) {
	memConfig := NewMemConfig()

	memConfig.Set("test", 100)

	res := memConfig.GetInt32("test")
	if res != 100 {
		t.Errorf("结果错误 应该 100 实际 %d", res)
	}
}

func TestGetInt64(t *testing.T) {
	memConfig := NewMemConfig()

	memConfig.Set("test", 100)

	res := memConfig.GetInt64("test")
	if res != 100 {
		t.Errorf("结果错误 应该 100 实际 %d", res)
	}
}

func TestGetUint(t *testing.T) {
	memConfig := NewMemConfig()

	memConfig.Set("test", 100)

	res := memConfig.GetUint("test")
	if res != 100 {
		t.Errorf("结果错误 应该 100 实际 %d", res)
	}
}

func TestGetUint32(t *testing.T) {
	memConfig := NewMemConfig()

	memConfig.Set("test", 100)

	res := memConfig.GetUint32("test")
	if res != 100 {
		t.Errorf("结果错误 应该 100 实际 %d", res)
	}
}

func TestGetUint64(t *testing.T) {
	memConfig := NewMemConfig()

	memConfig.Set("test", 100)

	res := memConfig.GetUint64("test")
	if res != 100 {
		t.Errorf("结果错误 应该 100 实际 %d", res)
	}
}

func TestGetFloat64(t *testing.T) {
	memConfig := NewMemConfig()

	memConfig.Set("test", 0.002)

	res := memConfig.GetFloat64("test")
	if res != 0.002 {
		t.Errorf("结果错误 应该 0.002 实际 %f", res)
	}
}

func TestGetTime(t *testing.T) {
	memConfig := NewMemConfig()

	now := time.Now()
	memConfig.Set("test", now)

	res := memConfig.GetTime("test")
	if res != now {
		t.Errorf("结果错误 应该 %s 实际 %s", now.String(), res.String())
	}
}

func TestGetDuration(t *testing.T) {
	memConfig := NewMemConfig()
	memConfig.Set("test", time.Duration(100))

	res := memConfig.GetDuration("test")
	if res != 100 {
		t.Errorf("结果错误 应该100 实际 %d", res)
	}
}

func TestGetStringSlice(t *testing.T) {
	memConfig := NewMemConfig()
	memConfig.Set("test", []string{"config", "test"})

	res := memConfig.GetStringSlice("test")
	if len(res) != 2 {
		t.Errorf("结果错误 应该 有 2 个值 实际 %d", len(res))
	}
}

func TestGetStringMap(t *testing.T) {
	memConfig := NewMemConfig()
	memConfig.Set("test", map[string]interface{}{"config": 1000})

	res := memConfig.GetStringMap("test")

	var val interface{}
	var ok bool
	if val, ok = res["config"]; !ok {
		t.Errorf("结果错误 应该 有 config 下标")
	}

	if val.(int) != 1000 {
		t.Errorf("结果错误，应该是 1000, 实际 %d", val.(int))
	}
}

func TestGetStringMapString(t *testing.T) {
	memConfig := NewMemConfig()
	memConfig.Set("test", map[string]string{"config": "test"})

	res := memConfig.GetStringMapString("test")

	var str string
	var ok bool
	if str, ok = res["config"]; !ok {
		t.Errorf("结果错误 应该 有 config 下标")
	}

	if str != "test" {
		t.Errorf("结果错误，应该是 test, 实际 %s", str)
	}
}

func TestGetStringMapStringSlice(t *testing.T) {
	memConfig := NewMemConfig()
	memConfig.Set("test", map[string][]string{"config": {"test1", "test2"}})

	res := memConfig.GetStringMapStringSlice("test")

	var slice []string
	var ok bool

	if slice, ok = res["config"]; !ok {
		t.Errorf("结果错误 应该 有 config 下标")
	}

	if len(slice) != 2 {
		t.Errorf("结果错误 应该 有 2 个值， 实际 %d", len(slice))
	}
}

func TestGetSizeInBytes(t *testing.T) {
	memConfig := NewMemConfig()
	res := memConfig.GetSizeInBytes("test")
	if res != 0 {
		t.Errorf("结果错误 应该是 0 实际 %d", res)
	}
}

func TestUnmarshalKey(t *testing.T) {
	memConfig := NewMemConfig()
	res := memConfig.UnmarshalKey("test", "test")
	if res != nil {
		t.Errorf("结果错误 应该是 nil 实际 %T", res)
	}
}

func TestUnmarshal(t *testing.T) {
	memConfig := NewMemConfig()
	res := memConfig.Unmarshal("test")
	if res != nil {
		t.Errorf("结果错误 应该是 nil 实际 %T", res)
	}
}

func TestSet(t *testing.T) {
	memConfig := NewMemConfig()
	memConfig.Set("test", "config")
	name := memConfig.GetString("test")
	if name != "config" {
		t.Errorf("结果错误 应该是 test 实际 %s", name)
	}
}
