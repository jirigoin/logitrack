package repository

import (
	"fmt"
	"sync"

	"github.com/logitrack/core/internal/model"
)

// EventStore is the append-only log of domain events.
type EventStore interface {
	Append(event model.DomainEvent) error
	LoadStream(trackingID string) ([]model.DomainEvent, error)
	LoadAll() ([]model.DomainEvent, error)
}

type inMemoryEventStore struct {
	mu          sync.RWMutex
	streams     map[string][]model.DomainEvent // keyed by tracking ID
	allEvents   []model.DomainEvent            // global append-only log
	draftToReal map[string]string              // draft ID → real tracking ID
}

func NewInMemoryEventStore() EventStore {
	return &inMemoryEventStore{
		streams:     make(map[string][]model.DomainEvent),
		allEvents:   []model.DomainEvent{},
		draftToReal: make(map[string]string),
	}
}

func (s *inMemoryEventStore) Append(event model.DomainEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if event.EventType == model.EventDraftConfirmed {
		return s.applyDraftConfirmed(event)
	}

	// Resolve draft ID to real tracking ID if already confirmed
	trackingID := event.TrackingID
	if realID, ok := s.draftToReal[trackingID]; ok {
		trackingID = realID
		event.TrackingID = trackingID
	}

	event.Version = len(s.streams[trackingID]) + 1
	s.streams[trackingID] = append(s.streams[trackingID], event)
	s.allEvents = append(s.allEvents, event)
	return nil
}

// applyDraftConfirmed moves the entire draft stream to the new tracking ID.
// Called with the lock already held.
func (s *inMemoryEventStore) applyDraftConfirmed(event model.DomainEvent) error {
	payload, ok := event.Payload.(model.DraftConfirmedPayload)
	if !ok {
		return fmt.Errorf("invalid payload for draft_confirmed event")
	}
	oldID := payload.OldTrackingID
	newID := payload.NewTrackingID

	// Retag all prior draft events with the new tracking ID
	existing := s.streams[oldID]
	for i := range existing {
		existing[i].TrackingID = newID
	}

	event.Version = len(existing) + 1
	newStream := append(existing, event)
	s.streams[newID] = newStream
	delete(s.streams, oldID)
	s.draftToReal[oldID] = newID

	// Patch allEvents to reflect the new tracking ID
	for i := range s.allEvents {
		if s.allEvents[i].TrackingID == oldID {
			s.allEvents[i].TrackingID = newID
		}
	}
	s.allEvents = append(s.allEvents, event)
	return nil
}

func (s *inMemoryEventStore) LoadStream(trackingID string) ([]model.DomainEvent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	events, ok := s.streams[trackingID]
	if !ok {
		return nil, fmt.Errorf("stream not found: %s", trackingID)
	}
	result := make([]model.DomainEvent, len(events))
	copy(result, events)
	return result, nil
}

func (s *inMemoryEventStore) LoadAll() ([]model.DomainEvent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]model.DomainEvent, len(s.allEvents))
	copy(result, s.allEvents)
	return result, nil
}
