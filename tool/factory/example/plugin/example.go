package main

import (
	"github.com/jukylin/esim/config"
	"github.com/jukylin/esim/log"
)

type Test struct {
	b int64

	c int8

	i bool

	f float32

	a int32

	h []int

	m map[string]interface{}

	e string

	g byte

	u [3]string

	d int16

	logger log.Logger

	conf config.Config
}

type empty struct{}
