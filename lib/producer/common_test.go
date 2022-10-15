package producer

import (
	"github.com/rs/zerolog"
	"testing"
)

func BenchmarkProducer_AddEvent(b *testing.B) {
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	producer := New(WithType(Beast), WithFiles([]string{"testdata/beast-smallish.sample"}))
	go func() {
		for range producer.out {
		}
	}()

	for n := 0; n < b.N; n++ {
		producer.run()
	}

}
