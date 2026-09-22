package hub

import (
	"sync"
	"testing"
	"time"
)

func TestPublishToNoSubscribersDoesNotPanic(t *testing.T) {
	h := New()
	h.Publish("nobody-listening", []byte("x"))
}

func TestSubscribeReceivesPublishedValue(t *testing.T) {
	h := New()
	ch, unsubscribe := h.Subscribe("geometry")
	defer unsubscribe()

	h.Publish("geometry", []byte(`{"a":1}`))

	select {
	case got := <-ch:
		if string(got) != `{"a":1}` {
			t.Errorf("got %q", got)
		}
	case <-time.After(time.Second):
		t.Fatal("did not receive published value")
	}
}

func TestPublishReachesEverySubscriber(t *testing.T) {
	h := New()
	ch1, unsub1 := h.Subscribe("geometry")
	defer unsub1()
	ch2, unsub2 := h.Subscribe("geometry")
	defer unsub2()

	h.Publish("geometry", []byte("v"))

	for _, ch := range []<-chan []byte{ch1, ch2} {
		select {
		case <-ch:
		case <-time.After(time.Second):
			t.Fatal("a subscriber did not receive the value")
		}
	}
}

func TestPublishIsIsolatedByTopic(t *testing.T) {
	h := New()
	ch, unsubscribe := h.Subscribe("a")
	defer unsubscribe()

	h.Publish("b", []byte("for b"))

	select {
	case got := <-ch:
		t.Fatalf("subscriber to a received %q, a publish to b", got)
	case <-time.After(50 * time.Millisecond):
	}
}

// unsubscribe closes the channel, giving a forwarding goroutine reading it a
// clean stop signal instead of one that just goes quiet.
func TestUnsubscribeClosesTheChannel(t *testing.T) {
	h := New()
	ch, unsubscribe := h.Subscribe("geometry")
	unsubscribe()

	select {
	case got, ok := <-ch:
		if ok {
			t.Fatalf("channel not closed, received %q", got)
		}
	case <-time.After(time.Second):
		t.Fatal("read from closed channel should not block")
	}
}

// A slow subscriber sees only the latest value, matching the pattern used
// elsewhere in the host (Geometry.Encoded, the Python bus this replaces).
func TestSlowSubscriberSeesOnlyLatestValue(t *testing.T) {
	h := New()
	ch, unsubscribe := h.Subscribe("geometry")
	defer unsubscribe()

	h.Publish("geometry", []byte("first"))
	h.Publish("geometry", []byte("second"))

	select {
	case got := <-ch:
		if string(got) != "second" {
			t.Fatalf("got %q, want the latest value only", got)
		}
	case <-time.After(time.Second):
		t.Fatal("did not receive a value")
	}

	select {
	case got := <-ch:
		t.Fatalf("received a second value %q, want the dropped first one gone for good", got)
	case <-time.After(50 * time.Millisecond):
	}
}

func TestConcurrentSubscribePublishUnsubscribe(t *testing.T) {
	h := New()

	var wg sync.WaitGroup

	for range 8 {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for range 50 {
				ch, unsubscribe := h.Subscribe("geometry")
				h.Publish("geometry", []byte("v"))

				select {
				case <-ch:
				default:
				}

				unsubscribe()
			}
		}()
	}

	for range 4 {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for range 50 {
				h.Publish("geometry", []byte("v"))
			}
		}()
	}

	wg.Wait()
}
