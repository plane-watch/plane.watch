package producer

import (
	"github.com/rs/zerolog"
	"sync"
	"testing"
)

func BenchmarkProducer_AddEvent(b *testing.B) {
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	producer := New(WithType(Beast), WithFiles([]string{"testdata/beast-smallish.sample"}))
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		for range producer.out {
		}
		wg.Done()
	}()

	for n := 0; n < b.N; n++ {
		producer.run()
	}
	producer.Stop()
	producer.Cleanup()
	wg.Wait()
}
