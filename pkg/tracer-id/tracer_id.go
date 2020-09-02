package tracerid

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/uber/jaeger-client-go/utils"
)

type contextKey struct{}

var ActiveEsimKey = contextKey{}

func TracerID() func() string {
	seedGenerator := utils.NewRand(time.Now().UnixNano())
	pool := sync.Pool{
		New: func() interface{} {
			return rand.NewSource(seedGenerator.Int63())
		},
	}

	randomNumber := func() string {
		generator := pool.Get().(rand.Source)
		number := uint64(generator.Int63())
		pool.Put(generator)
		return fmt.Sprintf("%x", number)
	}

	return randomNumber
}
