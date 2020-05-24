package filedir

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrWrite_Write(t *testing.T) {
	write := NewErrWrite(0)
	err := write.Write("", "")
	assert.Error(t, err)

	write = NewErrWrite(2)
	err = write.Write("", "")
	assert.Nil(t, err)

	err = write.Write("", "")
	assert.Nil(t, err)

	err = write.Write("", "")
	assert.Error(t, err)
}
