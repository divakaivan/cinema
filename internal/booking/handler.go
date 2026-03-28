package booking

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/divakaivan/cinema/internal/utils"
)

type handler struct {
	svc *Service
}

func NewHandler(svc *Service) *handler {
	return &handler{svc: svc}
}

type errorResponse struct {
	Error string `json:"error"`
}

func (h *handler) ListSeats(w http.ResponseWriter, r *http.Request) {
	movieID := r.PathValue("movieID")
	if movieID == "" {
		utils.WriteJSON(w, http.StatusBadRequest, errorResponse{Error: "movieID is required"})
		return
	}

	bookings := h.svc.store.ListBookings(movieID)
	seats := make([]seatInfo, 0, len(bookings))
	for _, b := range bookings {
		seats = append(seats, seatInfo{
			SeatID:    b.SeatID,
			UserID:    b.UserID,
			Booked:    true,
			Confirmed: b.Status == "confirmed",
		})
	}

	utils.WriteJSON(w, http.StatusOK, seats)
}

func (h *handler) HoldSeat(w http.ResponseWriter, r *http.Request) {
	movieID := r.PathValue("movieID")
	seatID := r.PathValue("seatID")

	if movieID == "" || seatID == "" {
		utils.WriteJSON(w, http.StatusBadRequest, errorResponse{Error: "movieID and seatID are required"})
		return
	}

	var req holdRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})
		return
	}

	if req.UserID == "" {
		utils.WriteJSON(w, http.StatusBadRequest, errorResponse{Error: "user_id is required"})
		return
	}

	data := Booking{
		UserID:  req.UserID,
		SeatID:  seatID,
		MovieID: movieID,
	}

	session, err := h.svc.Book(data)
	if err != nil {
		log.Printf("failed to book seat: %v", err)
		status := http.StatusInternalServerError
		if err == ErrSeatAlreadyBooked {
			status = http.StatusConflict
		}
		utils.WriteJSON(w, status, errorResponse{Error: err.Error()})
		return
	}

	type holdResponse struct {
		SessionID string `json:"session_id"`
		MovieID   string `json:"movieID"`
		SeatID    string `json:"seat_id"`
		ExpiresAt string `json:"expires_at"`
	}
	utils.WriteJSON(w, http.StatusCreated, holdResponse{
		SeatID:    seatID,
		MovieID:   session.MovieID,
		SessionID: session.ID,
		ExpiresAt: session.ExpiresAt.Format(time.RFC3339),
	})
}

func (h *handler) ConfirmSession(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("sessionID")

	if sessionID == "" {
		utils.WriteJSON(w, http.StatusBadRequest, errorResponse{Error: "sessionID is required"})
		return
	}

	var req holdRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})
		return
	}

	if req.UserID == "" {
		utils.WriteJSON(w, http.StatusBadRequest, errorResponse{Error: "user_id is required"})
		return
	}

	session, err := h.svc.ConfirmSeat(r.Context(), sessionID, req.UserID)
	if err != nil {
		log.Printf("failed to confirm seat: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, errorResponse{Error: err.Error()})
		return
	}

	utils.WriteJSON(w, http.StatusOK, sessionResponse{
		SessionID: session.ID,
		MovieID:   session.MovieID,
		SeatID:    session.SeatID,
		UserID:    req.UserID,
		Status:    session.Status,
	})
}

func (h *handler) ReleaseSession(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("sessionID")

	if sessionID == "" {
		utils.WriteJSON(w, http.StatusBadRequest, errorResponse{Error: "sessionID is required"})
		return
	}

	var req holdRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})
		return
	}

	if req.UserID == "" {
		utils.WriteJSON(w, http.StatusBadRequest, errorResponse{Error: "user_id is required"})
		return
	}

	err := h.svc.ReleaseSeat(r.Context(), sessionID, req.UserID)
	if err != nil {
		log.Printf("failed to release seat: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, errorResponse{Error: err.Error()})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type sessionResponse struct {
	SessionID string `json:"session_id"`
	MovieID   string `json:"movie_id"`
	SeatID    string `json:"seat_id"`
	UserID    string `json:"user_id"`
	Status    string `json:"status"`
	ExpiresAt string `json:"expires_at,omitempty"`
}

type holdRequest struct {
	UserID string `json:"user_id"`
}

type seatInfo struct {
	SeatID    string `json:"seat_id"`
	UserID    string `json:"user_id"`
	Booked    bool   `json:"booked"`
	Confirmed bool   `json:"confirmed"`
}
