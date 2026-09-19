package outbox

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
)

type fakeStore struct {
	mu        sync.Mutex
	event     Event
	available time.Time
	claimed   bool
	published bool
	dead      bool
	failures  int
}

func (s *fakeStore) Enqueue(context.Context, string, string, *uuid.UUID, any) (*Event, error) {
	return nil, errors.New("not implemented")
}

func (s *fakeStore) Claim(_ context.Context, _ int, leaseUntil time.Time) ([]Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.published || s.dead || s.claimed || time.Now().UTC().Before(s.available) {
		return nil, nil
	}
	s.claimed = true
	s.available = leaseUntil
	s.event.AttemptCount = int32(s.failures)
	return []Event{s.event}, nil
}

func (s *fakeStore) MarkPublished(_ context.Context, _ uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.published = true
	s.claimed = false
	return nil
}

func (s *fakeStore) RecordFailure(_ context.Context, _ uuid.UUID, _ string, availableAt time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.failures++
	s.claimed = false
	s.available = availableAt
	return nil
}

func (s *fakeStore) DeadLetter(_ context.Context, _ uuid.UUID, _ string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.failures++
	s.claimed = false
	s.dead = true
	return nil
}

func TestDispatcherRetriesFailedEventThenPublishes(t *testing.T) {
	store := &fakeStore{
		event:     Event{ID: uuid.New(), EventType: "test.requested"},
		available: time.Now().Add(-time.Second),
	}
	dispatcher := NewDispatcher(store, nil, DispatcherConfig{
		BatchSize:     1,
		LeaseDuration: time.Second,
		MaxBackoff:    time.Nanosecond,
	})

	attempts := 0
	if err := dispatcher.Register("test.requested", func(context.Context, Event) error {
		attempts++
		if attempts == 1 {
			return errors.New("temporary failure")
		}
		return nil
	}); err != nil {
		t.Fatalf("register handler: %v", err)
	}

	processed, err := dispatcher.RunOnce(context.Background())
	if err == nil || processed != 0 {
		t.Fatalf("expected first dispatch to fail, processed=%d err=%v", processed, err)
	}
	time.Sleep(time.Millisecond)
	processed, err = dispatcher.RunOnce(context.Background())
	if err != nil || processed != 1 {
		t.Fatalf("expected retry to publish event, processed=%d err=%v", processed, err)
	}
	if !store.published || store.failures != 1 || attempts != 2 {
		t.Fatalf("unexpected retry state: published=%v failures=%d attempts=%d", store.published, store.failures, attempts)
	}
}

func TestDispatcherLeavesUnknownEventsForRetry(t *testing.T) {
	store := &fakeStore{
		event:     Event{ID: uuid.New(), EventType: "unknown.event"},
		available: time.Now().Add(-time.Second),
	}
	dispatcher := NewDispatcher(store, nil, DispatcherConfig{MaxBackoff: time.Nanosecond})

	processed, err := dispatcher.RunOnce(context.Background())
	if err == nil || processed != 0 {
		t.Fatalf("expected missing handler error, processed=%d err=%v", processed, err)
	}
	if store.published || store.failures != 1 {
		t.Fatalf("unknown event must remain unpublished with one recorded failure")
	}
}

func TestDispatcherDeadLettersPoisonEvent(t *testing.T) {
	store := &fakeStore{
		event:     Event{ID: uuid.New(), EventType: "poison.event", AttemptCount: 2},
		available: time.Now().Add(-time.Second),
		failures:  2,
	}
	dispatcher := NewDispatcher(store, nil, DispatcherConfig{MaxAttempts: 3})
	if err := dispatcher.Register("poison.event", func(context.Context, Event) error {
		return errors.New("permanent failure")
	}); err != nil {
		t.Fatalf("register handler: %v", err)
	}

	if _, err := dispatcher.RunOnce(context.Background()); err == nil {
		t.Fatal("expected poison event dispatch to fail")
	}
	if !store.dead || store.published || store.failures != 3 {
		t.Fatalf("expected event to be dead-lettered on third attempt: %+v", store)
	}
}
