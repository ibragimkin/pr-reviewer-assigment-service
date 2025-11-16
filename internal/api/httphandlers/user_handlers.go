package httphandlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"pr-reviewer-assigment-service/internal/api/dto"
	"pr-reviewer-assigment-service/internal/application/service"
	"pr-reviewer-assigment-service/internal/domain"
)

// UserHandlers содержит хендлеры для /users/*
type UserHandlers struct {
	userService *service.UserService
}

func NewUserHandlers(userService *service.UserService) *UserHandlers {
	return &UserHandlers{userService: userService}
}

func (h *UserHandlers) SetIsActive(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w)
		return
	}

	var req dto.UserSetIsActiveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadRequest(w, "invalid JSON body")
		return
	}

	if req.UserID == "" {
		writeBadRequest(w, "user_id is required")
		return
	}

	user, err := h.userService.SetIsActive(r.Context(), req.UserID, req.IsActive)
	if err != nil {
		writeDomainError(w, err)
		return
	}

	resp := dto.UserResponse{
		User: dto.UserDto{
			UserID:   user.UserID,
			Username: user.Username,
			TeamName: user.TeamName,
			IsActive: user.IsActive,
		},
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *UserHandlers) GetReview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}

	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		writeBadRequest(w, "user_id is required")
		return
	}

	id, prs, err := h.userService.GetReview(r.Context(), userID)
	if err != nil {
		writeDomainError(w, err)
		return
	}

	prDtos := make([]dto.PullRequestShortDto, 0, len(prs))
	for _, pr := range prs {
		prDtos = append(prDtos, dto.PullRequestShortDto{
			PullRequestID:   pr.PullRequestID,
			PullRequestName: pr.PullRequestName,
			AuthorID:        pr.AuthorID,
			Status:          pr.Status,
		})
	}

	resp := dto.UserGetReviewResponse{
		UserID:       id,
		PullRequests: prDtos,
	}

	writeJSON(w, http.StatusOK, resp)
}

type errorResponse struct {
	Error errorBody `json:"error"`
}

type errorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeBadRequest(w http.ResponseWriter, message string) {
	writeJSON(w, http.StatusBadRequest, errorResponse{
		Error: errorBody{
			Code:    string(domain.ErrorNotFound),
			Message: message,
		},
	})
}

func writeMethodNotAllowed(w http.ResponseWriter) {
	w.Header().Set("Allow", "GET, POST")
	writeJSON(w, http.StatusMethodNotAllowed, errorResponse{
		Error: errorBody{
			Code:    "METHOD_NOT_ALLOWED",
			Message: "method not allowed",
		},
	})
}

func writeDomainError(w http.ResponseWriter, err error) {
	var dErr *domain.Error
	if errors.As(err, &dErr) {
		switch dErr.Code {
		case domain.ErrorNotFound:
			writeJSON(w, http.StatusNotFound, errorResponse{
				Error: errorBody{
					Code:    string(dErr.Code),
					Message: dErr.Message,
				},
			})
			return
		case domain.ErrorTeamExists,
			domain.ErrorPRExists,
			domain.ErrorNotAssigned,
			domain.ErrorNoCandidate,
			domain.ErrorPRMerged:
			writeJSON(w, http.StatusConflict, errorResponse{
				Error: errorBody{
					Code:    string(dErr.Code),
					Message: dErr.Message,
				},
			})
			return
		default:
			writeJSON(w, http.StatusBadRequest, errorResponse{
				Error: errorBody{
					Code:    string(dErr.Code),
					Message: dErr.Message,
				},
			})
			return
		}
	}

	writeJSON(w, http.StatusInternalServerError, errorResponse{
		Error: errorBody{
			Code:    "INTERNAL_ERROR",
			Message: "internal server error",
		},
	})
}
