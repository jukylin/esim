package config

import (
	"testing"
	"time"
)

const (
	testKey = "test"

	testStrVal = "config"

	testIntVal = 100

	testMapVal = 1000

	testFloatVal = 0.002
)

func TestGet(t *testing.T) {
	memConfig := NewMemConfig()

	memConfig.Set(testKey, testStrVal)

	res := memConfig.Get(testKey)
	if res.(string) != testStrVal {
		t.Errorf("结果错误 应该 config 实际 %s", res.(string))
	}
}

func TestGetString(t *testing.T) {
	memConfig := NewMemConfig()

	memConfig.Set(testKey, testStrVal)

	res := memConfig.GetString(testKey)
	if res != testStrVal {
		t.Errorf("结果错误 应该 config 实际 %s", res)
	}
}

func TestGetBool(t *testing.T) {
	memConfig := NewMemConfig()

	memConfig.Set(testKey, true)

	res := memConfig.GetBool(testKey)
	if res != true {
		t.Errorf("结果错误 应该 true 实际 false")
	}
}

func TestGetInt(t *testing.T) {
	memConfig := NewMemConfig()

	memConfig.Set(testKey, testIntVal)

	res := memConfig.GetInt(testKey)
	if res != testIntVal {
		t.Errorf("结果错误 应该 100 实际 %d", res)
	}
}

func TestGetInt32(t *testing.T) {
	memConfig := NewMemConfig()

	memConfig.Set(testKey, testIntVal)

	res := memConfig.GetInt32(testKey)
	if res != testIntVal {
		t.Errorf("结果错误 应该 100 实际 %d", res)
	}
}

func TestGetInt64(t *testing.T) {
	memConfig := NewMemConfig()

	memConfig.Set(testKey, testIntVal)

	res := memConfig.GetInt64(testKey)
	if res != testIntVal {
		t.Errorf("结果错误 应该 100 实际 %d", res)
	}
}

func TestGetUint(t *testing.T) {
	memConfig := NewMemConfig()

	memConfig.Set(testKey, testIntVal)

	res := memConfig.GetUint(testKey)
	if res != testIntVal {
		t.Errorf("结果错误 应该 100 实际 %d", res)
	}
}

func TestGetUint32(t *testing.T) {
	memConfig := NewMemConfig()

	memConfig.Set(testKey, testIntVal)

	res := memConfig.GetUint32(testKey)
	if res != testIntVal {
		t.Errorf("结果错误 应该 100 实际 %d", res)
	}
}

func TestGetUint64(t *testing.T) {
	memConfig := NewMemConfig()

	memConfig.Set(testKey, testIntVal)

	res := memConfig.GetUint64(testKey)
	if res != testIntVal {
		t.Errorf("结果错误 应该 100 实际 %d", res)
	}
}

func TestGetFloat64(t *testing.T) {
	memConfig := NewMemConfig()

	memConfig.Set(testKey, testFloatVal)

	res := memConfig.GetFloat64(testKey)
	if res != testFloatVal {
		t.Errorf("结果错误 应该 0.002 实际 %f", res)
	}
}

func TestGetTime(t *testing.T) {
	memConfig := NewMemConfig()

	now := time.Now()
	memConfig.Set(testKey, now)

	res := memConfig.GetTime(testKey)
	if res != now {
		t.Errorf("结果错误 应该 %s 实际 %s", now.String(), res.String())
	}
}

func TestGetDuration(t *testing.T) {
	memConfig := NewMemConfig()
	memConfig.Set(testKey, time.Duration(testIntVal))

	res := memConfig.GetDuration(testKey)
	if res != testIntVal {
		t.Errorf("结果错误 应该100 实际 %d", res)
	}
}

func TestGetStringSlice(t *testing.T) {
	memConfig := NewMemConfig()
	memConfig.Set(testKey, []string{testStrVal, testKey})

	res := memConfig.GetStringSlice(testKey)
	if len(res) != 2 {
		t.Errorf("结果错误 应该 有 2 个值 实际 %d", len(res))
	}
}

func TestGetStringMap(t *testing.T) {
	memConfig := NewMemConfig()
	memConfig.Set(testKey, map[string]interface{}{testStrVal: testMapVal})

	res := memConfig.GetStringMap(testKey)

	var val interface{}
	var ok bool
	if val, ok = res[testStrVal]; !ok {
		t.Errorf("结果错误 应该 有 config 下标")
	}

	if val.(int) != testMapVal {
		t.Errorf("结果错误，应该是 1000, 实际 %d", val.(int))
	}
}

func TestGetStringMapString(t *testing.T) {
	memConfig := NewMemConfig()
	memConfig.Set(testKey, map[string]string{testStrVal: testKey})

	res := memConfig.GetStringMapString(testKey)

	var str string
	var ok bool
	if str, ok = res[testStrVal]; !ok {
		t.Errorf("结果错误 应该 有 config 下标")
	}

	if str != testKey {
		t.Errorf("结果错误，应该是 test, 实际 %s", str)
	}
}

func TestGetStringMapStringSlice(t *testing.T) {
	memConfig := NewMemConfig()
	memConfig.Set(testKey, map[string][]string{testStrVal: {"test1", "test2"}})

	res := memConfig.GetStringMapStringSlice(testKey)

	var slice []string
	var ok bool

	if slice, ok = res[testStrVal]; !ok {
		t.Errorf("结果错误 应该 有 config 下标")
	}

	if len(slice) != 2 {
		t.Errorf("结果错误 应该 有 2 个值， 实际 %d", len(slice))
	}
}

func TestGetSizeInBytes(t *testing.T) {
	memConfig := NewMemConfig()
	res := memConfig.GetSizeInBytes(testKey)
	if res != 0 {
		t.Errorf("结果错误 应该是 0 实际 %d", res)
	}
}

func TestUnmarshalKey(t *testing.T) {
	memConfig := NewMemConfig()
	res := memConfig.UnmarshalKey(testKey, testKey)
	if res != nil {
		t.Errorf("结果错误 应该是 nil 实际 %T", res)
	}
}

func TestUnmarshal(t *testing.T) {
	memConfig := NewMemConfig()
	res := memConfig.Unmarshal(testKey)
	if res != nil {
		t.Errorf("结果错误 应该是 nil 实际 %T", res)
	}
}

func TestSet(t *testing.T) {
	memConfig := NewMemConfig()
	memConfig.Set(testKey, testStrVal)
	name := memConfig.GetString(testKey)
	if name != "config" {
		t.Errorf("结果错误 应该是 test 实际 %s", name)
	}
}
