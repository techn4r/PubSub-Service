package main

import (
	"context"
	"sync"
	"testing"
)

func TestSubPub(t *testing.T) {
	sp := NewSubPub()
	var wg sync.WaitGroup
	received := ""

	wg.Add(1)
	sp.Subscribe("topic", func(msg interface{}) {
		if str, ok := msg.(string); ok {
			received = str
			wg.Done()
		}
	})

	err := sp.Publish("topic", "Hello, World!")
	if err != nil {
		t.Errorf("Publish failed: %v", err)
	}

	wg.Wait()

	if received != "Hello, World!" {
		t.Errorf("Expected 'Hello, World!', got '%s'", received)
	}

	sp.Close(context.Background())
}
