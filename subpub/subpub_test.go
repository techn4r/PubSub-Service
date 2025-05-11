package subpub_test

import (
	"VKIntern/subpub"
	"context"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestPublishSubscribe(t *testing.T) {
	sb := subpub.NewSubPub()
	defer sb.Close(context.Background())

	ch := make(chan interface{}, 1)
	_, err := sb.Subscribe("topic", func(msg interface{}) {
		ch <- msg
	})
	require.NoError(t, err)

	err = sb.Publish("topic", "hello")
	require.NoError(t, err)

	select {
	case got := <-ch:
		require.Equal(t, "hello", got)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for message")
	}
}

func TestOrdering(t *testing.T) {
	sb := subpub.NewSubPub()
	defer sb.Close(context.Background())

	const n = 3
	in := []string{"one", "two", "three"}
	out := make([]string, 0, n)
	done := make(chan struct{})

	_, err := sb.Subscribe("k", func(msg interface{}) {
		out = append(out, msg.(string))
		if len(out) == n {
			close(done)
		}
	})
	require.NoError(t, err)

	for _, v := range in {
		require.NoError(t, sb.Publish("k", v))
	}

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for all messages")
	}

	require.Equal(t, in, out)
}

func TestUnsubscribe(t *testing.T) {
	sb := subpub.NewSubPub()
	defer sb.Close(context.Background())

	var got []int
	sub, err := sb.Subscribe("foo", func(msg interface{}) {
		got = append(got, msg.(int))
	})
	require.NoError(t, err)

	require.NoError(t, sb.Publish("foo", 1))
	time.Sleep(20 * time.Millisecond)

	sub.Unsubscribe()
	require.NoError(t, sb.Publish("foo", 2))
	time.Sleep(20 * time.Millisecond)

	require.Equal(t, []int{1}, got)
}

func TestCloseCancels(t *testing.T) {
	sb := subpub.NewSubPub()

	_, err := sb.Subscribe("a", func(msg interface{}) {
		time.Sleep(200 * time.Millisecond)
	})
	require.NoError(t, err)

	require.NoError(t, sb.Publish("a", "x"))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	err = sb.Close(ctx)
	require.Error(t, err, "ожидаем ошибку по ctx")
}

func TestAfterClose(t *testing.T) {
	sb := subpub.NewSubPub()
	require.NoError(t, sb.Close(context.Background()))

	_, err := sb.Subscribe("x", func(_ interface{}) {})
	require.Equal(t, subpub.ErrClosed, err)

	err = sb.Publish("x", nil)
	require.Equal(t, subpub.ErrClosed, err)
}
