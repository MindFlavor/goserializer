package serializer

import (
	"testing"

	log "github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestInstantiate(t *testing.T) {
	pa := New()
	assert.NotNil(t, pa)
	pa.Close()
}

func TestConcurrentAdd(t *testing.T) {
	iterations := 100

	pa := New()
	assert.NotNil(t, pa)

	log.SetLevel(log.DebugLevel)

	var slice []int

	cCount := make(chan int, iterations)

	for i := 0; i < iterations; i++ {
		go func(i int) {
			pa.Serialize(func() interface{} {
				slice = append(slice, i)
				cCount <- i
				return nil
			})
		}(i)
	}

	for i := 0; i < iterations; i++ {
		log.Infof("%d", i)
		<-cCount
	}

	log.Debug("test sending Close()")
	pa.Close()

	assert.Equal(t, iterations, len(slice))
}

func TestClose(t *testing.T) {
	pa := New()

	assert.NotNil(t, pa)

	pa.Close()

	assert.Panics(t, func() { pa.Serialize(func() interface{} { return nil }) })
	assert.Panics(t, func() { pa.Close() })
}
