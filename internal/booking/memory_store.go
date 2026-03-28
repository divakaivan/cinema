package booking

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type MemoryStore struct {
	bookings map[string]Booking
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		bookings: map[string]Booking{},
	}
}

func (s *MemoryStore) Book(b Booking) (Booking, error) {
	if _, exists := s.bookings[b.SeatID]; exists {
		return Booking{}, ErrSeatAlreadyBooked
	}

	b.ID = uuid.New().String()
	b.Status = "held"
	b.ExpiresAt = time.Now().Add(2 * time.Minute)

	s.bookings[b.SeatID] = b

	return b, nil
}

func (s *MemoryStore) ListBookings(movieID string) []Booking {
	var res []Booking
	for _, b := range s.bookings {
		if b.MovieID == movieID {
			res = append(res, b)
		}
	}
	return res
}

func (s *MemoryStore) Confirm(ctx context.Context, sessionID string, userID string) (Booking, error) {
	for seatID, b := range s.bookings {
		if b.ID == sessionID && b.UserID == userID {
			b.Status = "confirmed"
			s.bookings[seatID] = b
			return b, nil
		}
	}
	return Booking{}, ErrSeatAlreadyBooked
}

func (s *MemoryStore) Release(ctx context.Context, sessionID string, userID string) error {
	for seatID, b := range s.bookings {
		if b.ID == sessionID && b.UserID == userID {
			delete(s.bookings, seatID)
			return nil
		}
	}
	return ErrSeatAlreadyBooked
}
