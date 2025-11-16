package httphandlers

import (
	"encoding/json"
	"net/http"
	"time"

	"pr-reviewer-assigment-service/internal/api/dto"
	"pr-reviewer-assigment-service/internal/application/service"
	"pr-reviewer-assigment-service/internal/domain"
)

// PullRequestHandlers содержит хендлеры для /pullRequest/*
type PullRequestHandlers struct {
	prService *service.PullRequestService
}

func NewPullRequestHandlers(prService *service.PullRequestService) *PullRequestHandlers {
	return &PullRequestHandlers{prService: prService}
}

func (h *PullRequestHandlers) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w)
		return
	}

	var req dto.PullRequestCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadRequest(w, "invalid JSON body")
		return
	}

	if req.PullRequestID == "" {
		writeBadRequest(w, "pull_request_id is required")
		return
	}
	if req.PullRequestName == "" {
		writeBadRequest(w, "pull_request_name is required")
		return
	}
	if req.AuthorID == "" {
		writeBadRequest(w, "author_id is required")
		return
	}

	pr, err := h.prService.Create(r.Context(), req.PullRequestID, req.PullRequestName, req.AuthorID)
	if err != nil {
		writeDomainError(w, err)
		return
	}

	resp := dto.PullRequestCreateResponse{
		PR: toPullRequestDto(pr),
	}

	writeJSON(w, http.StatusCreated, resp)
}

func (h *PullRequestHandlers) Merge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w)
		return
	}

	var req dto.PullRequestMergeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadRequest(w, "invalid JSON body")
		return
	}

	if req.PullRequestID == "" {
		writeBadRequest(w, "pull_request_id is required")
		return
	}

	pr, err := h.prService.Merge(r.Context(), req.PullRequestID)
	if err != nil {
		writeDomainError(w, err)
		return
	}

	resp := dto.PullRequestMergeResponse{
		PR: toPullRequestDto(pr),
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *PullRequestHandlers) Reassign(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w)
		return
	}

	var req dto.PullRequestReassignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadRequest(w, "invalid JSON body")
		return
	}

	if req.PullRequestID == "" {
		writeBadRequest(w, "pull_request_id is required")
		return
	}
	if req.OldUserID == "" {
		writeBadRequest(w, "old_user_id is required")
		return
	}

	pr, replacedBy, err := h.prService.Reassign(r.Context(), req.PullRequestID, req.OldUserID)
	if err != nil {
		writeDomainError(w, err)
		return
	}

	resp := dto.PullRequestReassignResponse{
		PR:         toPullRequestDto(pr),
		ReplacedBy: replacedBy,
	}

	writeJSON(w, http.StatusOK, resp)
}

func toPullRequestDto(pr *domain.PullRequest) dto.PullRequestDto {
	var createdAtStr *string
	if pr.CreatedAt != nil {
		s := pr.CreatedAt.UTC().Format(time.RFC3339)
		createdAtStr = &s
	}

	var mergedAtStr *string
	if pr.MergedAt != nil {
		s := pr.MergedAt.UTC().Format(time.RFC3339)
		mergedAtStr = &s
	}

	return dto.PullRequestDto{
		PullRequestID:     pr.PullRequestID,
		PullRequestName:   pr.PullRequestName,
		AuthorID:          pr.AuthorID,
		Status:            pr.Status,
		AssignedReviewers: append([]string(nil), pr.AssignedReviewers...),
		CreatedAt:         createdAtStr,
		MergedAt:          mergedAtStr,
	}
}
