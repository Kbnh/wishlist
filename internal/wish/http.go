package wish

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"
)

// DTO
type createWishRequest struct {
	OwnerEmail  string `json:"owner_email"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
}

type updateWishRequest struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
}

type WishResponse struct {
	ID          int        `json:"id"`
	OwnerEmail  string     `json:"owner_email"`
	Title       string     `json:"title"`
	Description string     `json:"description,omitempty"`
	IsBought    bool       `json:"is_bought"`
	BoughtAt    *time.Time `json:"bought_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

// Conv
func toResponse(w *Wish) WishResponse {
	return WishResponse{
		ID:          w.ID,
		OwnerEmail:  w.OwnerEmail,
		Title:       w.Title,
		Description: w.Description,
		IsBought:    w.IsBought,
		BoughtAt:    w.BoughtAt,
		CreatedAt:   w.CreatedAt,
	}
}

// Handlers
func CreateWishHandler(svc WishService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req createWishRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondWithError(w, http.StatusBadRequest, "invalid JSON")
			return
		}

		wish, err := svc.CreateWish(r.Context(), req.OwnerEmail, req.Title, req.Description)
		if err != nil {
			if errors.Is(err, ErrOwnerEmailRequired) || errors.Is(err, ErrTitleRequired) {
				respondWithError(w, http.StatusBadRequest, err.Error())
			} else {
				respondWithError(w, http.StatusInternalServerError, "failed to create wish")
			}
			return
		}

		respondWithJSON(w, http.StatusCreated, toResponse(wish))
	}
}

func GetWishHandler(svc WishService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "invalid wish ID")
			return
		}

		wish, err := svc.GetWish(r.Context(), id)
		if err != nil {
			if errors.Is(err, ErrWishNotFound) {
				respondWithError(w, http.StatusNotFound, ErrWishNotFound.Error())
			} else {
				respondWithError(w, http.StatusInternalServerError, "failed to fetch wish")
			}
			return
		}

		respondWithJSON(w, http.StatusOK, toResponse(wish))
	}
}

func ListWishHandler(svc WishService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ownerEmail := r.URL.Query().Get("owner_email")
		if ownerEmail == "" {
			respondWithError(w, http.StatusBadRequest, ErrOwnerEmailRequired.Error())
			return
		}

		boughtStr := r.URL.Query().Get("bought")
		var bought *bool
		if boughtStr != "" {
			b, err := strconv.ParseBool(boughtStr)
			if err != nil {
				respondWithError(w, http.StatusBadRequest, "invalid bought parameter")
				return
			}
			bought = &b
		}

		wishes, err := svc.ListWishes(r.Context(), ownerEmail, bought)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "failed to list wishes")
			return
		}

		responses := make([]WishResponse, len(wishes))
		for i, wish := range wishes {
			responses[i] = toResponse(wish)
		}

		respondWithJSON(w, http.StatusOK, responses)
	}
}

func UpdateWishHandler(svc WishService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "invalid wish ID")
			return
		}

		var req updateWishRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondWithError(w, http.StatusBadRequest, "invalid JSON")
			return
		}

		wish, err := svc.UpdateWish(r.Context(), id, req.Title, req.Description)
		if err != nil {
			if errors.Is(err, ErrWishNotFound) {
				respondWithError(w, http.StatusNotFound, ErrWishNotFound.Error())
			} else {
				respondWithError(w, http.StatusInternalServerError, "failed to update wish")
			}
			return
		}

		respondWithJSON(w, http.StatusOK, toResponse(wish))
	}
}

func DeleteWishHandler(svc WishService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "invalid wish ID")
			return
		}

		if err := svc.DeleteWish(r.Context(), id); err != nil {
			if errors.Is(err, ErrWishNotFound) {
				respondWithError(w, http.StatusNotFound, ErrWishNotFound.Error())
				return
			} else {
				respondWithError(w, http.StatusInternalServerError, "couldnt delete wish")
				return
			}
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func BuyWishHandler(svc WishService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "invalid wish ID")
			return
		}

		if err := svc.BuyWish(r.Context(), id); err != nil {
			if errors.Is(err, ErrWishAlreadyBought) {
				respondWithError(w, http.StatusBadRequest, ErrWishAlreadyBought.Error())
			} else if errors.Is(err, ErrWishNotFound) {
				respondWithError(w, http.StatusNotFound, ErrWishNotFound.Error())
			} else {
				respondWithError(w, http.StatusInternalServerError, "failed to buy wish")
			}
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func StatsHandler(svc WishService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ownerEmail := r.URL.Query().Get("owner_email")
		if ownerEmail == "" {
			respondWithError(w, http.StatusBadRequest, ErrOwnerEmailRequired.Error())
			return
		}

		stats, err := svc.GetStats(r.Context(), ownerEmail)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "failed to get stats")
			return
		}

		respondWithJSON(w, http.StatusOK, stats)
	}
}

// Respond functions
func respondWithJSON(w http.ResponseWriter, code int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
