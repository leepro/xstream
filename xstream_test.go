package xstream

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestCreateStreamGroup(t *testing.T) {
	sg := New(context.TODO())
	if sg == nil {
		t.Error("StreamGroup is not created")
	}
}

func TestRegisterProducer(t *testing.T) {
	ctx := context.TODO()
	sg := New(ctx)
	sm := sg.RegisterProducer(ctx)
	if sm == nil {
		t.Error("Stream producer is not registered")
	}
	t.Log("Stream producer is registered", sm.ID)

	clients := 3
	var cnt int32
	var wg sync.WaitGroup

	for i := 0; i < clients; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			ctx, cancel := context.WithCancel(context.TODO())
			defer cancel()

			s, err := sg.Subscribe(ctx, sm.ID)
			if err != nil {
				t.Errorf("Subscriber %d is not registered. err=%s", i, err)
				return
			}

			t.Logf("Subscriber %d is registered with ID %s", i, sm.ID)
			go func() {
				time.Sleep(2 * time.Second)
				cancel()
			}()

			for d := range s.C {
				t.Logf("Subscriber %d received data [%v]", i, d.(string))
				atomic.AddInt32(&cnt, 1)
			}

			t.Logf("Subscriber %d is done", i)
		}(i)
	}

	sm.WriteC <- "Hello"
	sm.WriteC <- "Hello"
	sm.WriteC <- "Hello"

	time.Sleep(3 * time.Second)

	for _, sid := range sg.AllSessions() {
		sm, err := sg.LookupStream(sid)
		if err != nil {
			t.Errorf("Failed to lookup. err=%s\n", err)
			return
		}
		sm.CloseSubscribers()
	}

	// Wait for all goroutines to finish
	wg.Wait()

	if int(cnt) != 3*clients {
		t.Error("Subscriber did not receive all data", cnt)
	}
}
