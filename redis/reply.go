// Copyright 2012 Gary Burd
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package redis

import (
	"encoding/json"

	"github.com/gomodule/redigo/redis"
)

//refer redis.Int
func Int(reply interface{}, err error) (int, error) {
	rep, err := redis.Int(reply, err)
	return rep, decorError(err)
}

//refer redis.Int64
func Int64(reply interface{}, err error) (int64, error) {
	rep, err := redis.Int64(reply, err)
	return rep, decorError(err)
}

//refer redis.Uint64
func Uint64(reply interface{}, err error) (uint64, error) {
	rep, err := redis.Uint64(reply, err)
	return rep, decorError(err)
}

//refer redis.Float64
func Float64(reply interface{}, err error) (float64, error) {
	rep, err := redis.Float64(reply, err)
	return rep, decorError(err)
}

//refer redis.String
func String(reply interface{}, err error) (string, error) {
	rep, err := redis.String(reply, err)
	return rep, decorError(err)
}

//refer redis.Bytes
func Bytes(reply interface{}, err error) ([]byte, error) {
	rep, err := redis.Bytes(reply, err)
	return rep, decorError(err)
}

//refer redis.Bool
func Bool(reply interface{}, err error) (bool, error) {
	rep, err := redis.Bool(reply, err)
	return rep, decorError(err)
}

//refer redis.Values
func Values(reply interface{}, err error) ([]interface{}, error) {
	rep, err := redis.Values(reply, err)
	return rep, decorError(err)
}

//refer redis.Float64s
func Float64s(reply interface{}, err error) ([]float64, error) {
	return redis.Float64s(reply, err)
}

//refer redis.Strings
func Strings(reply interface{}, err error) ([]string, error) {
	return redis.Strings(reply, err)
}

//refer redis.ByteSlices
func ByteSlices(reply interface{}, err error) ([][]byte, error) {
	return redis.ByteSlices(reply, err)
}

//refer redis.Int64s
func Int64s(reply interface{}, err error) ([]int64, error) {
	return redis.Int64s(reply, err)
}

//refer redis.Ints
func Ints(reply interface{}, err error) ([]int, error) {
	return redis.Ints(reply, err)
}

//refer redis.StringMap
func StringMap(result interface{}, err error) (map[string]string, error) {
	return redis.StringMap(result, err)
}

//refer redis.IntMap
func IntMap(result interface{}, err error) (map[string]int, error) {
	return redis.IntMap(result, err)
}

//refer redis.IntMap
func Int64Map(result interface{}, err error) (map[string]int64, error) {
	return redis.Int64Map(result, err)
}

//refer redis.Positions
func Positions(result interface{}, err error) ([]*[2]float64, error) {
	return redis.Positions(result, err)
}

func Struct(result interface{}, err error, v ...interface{}) error {
	bytes, err := redis.Bytes(result, err)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bytes, &v)

	return err
}

//decorError implement decoration redis.ErrNil
func decorError(err error) error {
	if err == redis.ErrNil {
		err = nil
	}

	return err
}
