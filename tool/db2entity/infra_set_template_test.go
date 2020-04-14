package db2entity

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestInfraSetTemplate(t *testing.T)  {
	infraSetArgs := infraSetArgs{}
	infraSetArgs.Args = append(infraSetArgs.Args, "provideRedis")
	infraSetArgs.Args = append(infraSetArgs.Args, "provideDb")

	result := infraSetArgs.String()
	assert.NotEmpty(t, result)
}





