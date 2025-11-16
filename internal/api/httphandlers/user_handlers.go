package httphandlers

import (
	"encoding/json"
	"net/http"
	"pr-reviewer-assigment-service/internal/api/dto"
	"pr-reviewer-assigment-service/internal/application/service"
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
