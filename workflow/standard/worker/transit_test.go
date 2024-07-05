package worker

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewTransit(t *testing.T) {
	t.Run("nil worker & nil pool", func(t *testing.T) {
		transit, err := NewTransit("test", []string{}, []string{}, nil, nil)
		assert.Nil(t, transit)
		assert.Error(t, err, "worker is nil")
	})
	t.Run("nil pool", func(t *testing.T) {
		transit, err := NewTransit("test", []string{}, []string{}, func(context.Context, ...any) (any, error) { return nil, nil }, nil)
		assert.Nil(t, transit)
		assert.Error(t, err, "pool is nil")
	})
}

func TestSimpleTransit(t *testing.T) {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	worker := func(ctx context.Context, a ...any) (any, error) {
		log.Println(a, "begins...")
		defer log.Println(a, "ends...")
		time.Sleep(time.Millisecond * 100)
		return nil, nil
	}
	t.Run("parallel", func(t *testing.T) {
		pool := NewPool(WithMaxWorkers(1))
		transit, err := NewTransit("test", []string{"input"}, []string{"output"}, worker, pool)
		assert.NoError(t, err)
		transitWorker := transit.GetWorker()
		for i := 0; i < 3; i++ {
			transitWorker(context.Background(), i)
		}
	})
	t.Run("AddListeningAndSendingChannels", func(t *testing.T) {
		pool := NewPool(WithMaxWorkers(1))
		transit, err := NewTransit("test", []string{"input"}, []string{"output"}, worker, pool)
		assert.NoError(t, err)
		transit.AppendListeningChannels("transit1", "listening1", "listening2", "listening3")
		transit.AppendSendingChannels("transit1", "sending1", "sending2", "sending3")
	})
}
