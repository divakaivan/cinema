package booking

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
)

type ConcurrentStore struct {
	bookings map[string]Booking
	sync.RWMutex
}

func NewConcurrentStore() *ConcurrentStore {
	return &ConcurrentStore{
		bookings: map[string]Booking{},
	}
}

func (s *ConcurrentStore) Book(b Booking) (Booking, error) {
	s.Lock()
	defer s.Unlock()
	if _, exists := s.bookings[b.SeatID]; exists {
		return Booking{}, ErrSeatAlreadyBooked
	}

	b.ID = uuid.New().String()
	b.Status = "held"
	b.ExpiresAt = time.Now().Add(2 * time.Minute)

	s.bookings[b.SeatID] = b

	return b, nil
}

func (s *ConcurrentStore) ListBookings(movieID string) []Booking {
	s.RLock()
	defer s.RUnlock()
	var res []Booking
	for _, b := range s.bookings {
		if b.MovieID == movieID {
			res = append(res, b)
		}
	}
	return res
}

func (s *ConcurrentStore) Confirm(ctx context.Context, sessionID string, userID string) (Booking, error) {
	s.Lock()
	defer s.Unlock()
	for seatID, b := range s.bookings {
		if b.ID == sessionID && b.UserID == userID {
			b.Status = "confirmed"
			s.bookings[seatID] = b
			return b, nil
		}
	}
	return Booking{}, ErrSeatAlreadyBooked
}

func (s *ConcurrentStore) Release(ctx context.Context, sessionID string, userID string) error {
	s.Lock()
	defer s.Unlock()
	for seatID, b := range s.bookings {
		if b.ID == sessionID && b.UserID == userID {
			delete(s.bookings, seatID)
			return nil
		}
	}
	return ErrSeatAlreadyBooked
}
