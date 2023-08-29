package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/akxcix/nomadcore/pkg/handlers"
	"github.com/akxcix/nomadcore/pkg/services/auth"
	"github.com/google/uuid"
)

type Handlers struct {
	Service *auth.Service
}

func New(s *auth.Service) *Handlers {
	h := Handlers{
		Service: s,
	}

	return &h
}

// func (h *Handlers) RegisterUser(w http.ResponseWriter, r *http.Request) {
// 	var req UserAuthReq
// 	if err := handlers.FromRequest(r, &req); err != nil {
// 		handlers.RespondWithError(w, r, err, http.StatusBadRequest)
// 		return
// 	}

// 	if req.Email == "" {
// 		handlers.RespondWithError(w, r, errors.New("email is empty"), http.StatusBadRequest)
// 		return
// 	}

// 	if req.Password == "" {
// 		handlers.RespondWithError(w, r, errors.New("password is empty"), http.StatusBadRequest)
// 		return
// 	}

// 	msg, err := h.Service.RegisterUser(req.Email, req.Password)
// 	if err != nil {
// 		handlers.RespondWithError(w, r, err, http.StatusInternalServerError)
// 		return
// 	}

// 	handlers.RespondWithData(w, r, msg)
// }

// func (h *Handlers) LoginUser(w http.ResponseWriter, r *http.Request) {
// 	var req UserAuthReq
// 	if err := handlers.FromRequest(r, &req); err != nil {
// 		handlers.RespondWithError(w, r, err, http.StatusBadRequest)
// 		return
// 	}

// 	if req.Email == "" {
// 		handlers.RespondWithError(w, r, errors.New("email is empty"), http.StatusBadRequest)
// 		return
// 	}

// 	if req.Password == "" {
// 		handlers.RespondWithError(w, r, errors.New("password is empty"), http.StatusBadRequest)
// 		return
// 	}

// 	msg, err := h.Service.LoginUser(req.Email, req.Password)
// 	if err != nil {
// 		handlers.RespondWithError(w, r, err, http.StatusInternalServerError)
// 		return
// 	}

// 	handlers.RespondWithData(w, r, msg)
// }

// func (h *Handlers) ValidateJwt(w http.ResponseWriter, r *http.Request) {
// 	var req JwtVerifyRequest
// 	if err := handlers.FromRequest(r, &req); err != nil {
// 		handlers.RespondWithError(w, r, err, http.StatusBadRequest)
// 		return
// 	}

// 	_, isValid := h.Service.ValidateJwt(req.Jwt)
// 	if !isValid {
// 		handlers.RespondWithError(w, r, ErrInvalidJwt, http.StatusUnauthorized)
// 		return
// 	}

// 	handlers.RespondWithData(w, r, true)
// }

func (h *Handlers) CreateCalendar(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(handlers.UserIdContextKey).(uuid.UUID)
	if !ok {
		handlers.RespondWithError(w, r, ErrInvalidJwt, http.StatusBadRequest)
		return
	}
	var req CreateCalendarReq
	if err := handlers.FromRequest(r, &req); err != nil {
		handlers.RespondWithError(w, r, err, http.StatusBadRequest)
		return
	}

	err := h.Service.AuthRepo.CreateCalendar(userID, req.Name, req.Visibility)
	if err != nil {
		handlers.RespondWithError(w, r, err, http.StatusInternalServerError)
		return
	}

	handlers.RespondWithData(w, r, "update successful")
}

func (h *Handlers) GetPublicCalendars(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(handlers.UserIdContextKey).(uuid.UUID)
	if !ok {
		handlers.RespondWithError(w, r, ErrInvalidJwt, http.StatusBadRequest)
		return
	}

	calendars, err := h.Service.AuthRepo.GetPublicCalendars(userID)
	if err != nil {
		handlers.RespondWithError(w, r, err, http.StatusInternalServerError)
		return
	}

	calendarDtos := make([]CalendarDTO, 0)
	for _, cal := range calendars {
		calDto := CalendarDTO{
			ID:         cal.ID,
			Name:       cal.Name,
			Visibility: cal.Visibility,
		}

		calendarDtos = append(calendarDtos, calDto)
	}

	res := GetPublicCalendarsRes{
		Calendars: calendarDtos,
	}

	handlers.RespondWithData(w, r, res)
}

func (h *Handlers) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bearerToken := r.Header.Get("Authorization")

		// Check if token exists
		if bearerToken == "" {
			handlers.RespondWithError(w, r, ErrInvalidJwt, http.StatusUnauthorized)
			return
		}

		// Extract token from header
		tokenParts := strings.Split(bearerToken, " ")
		if len(tokenParts) != 2 {
			handlers.RespondWithError(w, r, ErrInvalidJwt, http.StatusUnauthorized)
			return
		}
		tokenString := tokenParts[1]

		// Validate token
		claims, isValid := h.Service.ValidateJwt(tokenString)
		if claims == nil || !isValid {
			handlers.RespondWithError(w, r, ErrInvalidJwt, http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), handlers.UserIdContextKey, claims.ID)

		// If token is valid, forward to the actual handler
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}