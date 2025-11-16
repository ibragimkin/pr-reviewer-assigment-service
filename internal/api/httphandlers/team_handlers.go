package httphandlers

import (
	"encoding/json"
	"net/http"

	"pr-reviewer-assigment-service/internal/api/dto"
	"pr-reviewer-assigment-service/internal/application/service"
	"pr-reviewer-assigment-service/internal/domain"
)

// TeamHandlers содержит хендлеры для /team/*
type TeamHandlers struct {
	teamService *service.TeamService
}

func NewTeamHandlers(teamService *service.TeamService) *TeamHandlers {
	return &TeamHandlers{teamService: teamService}
}

func (h *TeamHandlers) Add(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w)
		return
	}

	var req dto.TeamAddRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadRequest(w, "invalid JSON body")
		return
	}

	if req.TeamName == "" {
		writeBadRequest(w, "team_name is required")
		return
	}

	members := make([]domain.TeamMember, 0, len(req.Members))
	for _, m := range req.Members {
		members = append(members, domain.TeamMember{
			UserID:   m.UserID,
			Username: m.Username,
			IsActive: m.IsActive,
		})
	}

	team, err := h.teamService.Add(r.Context(), req.TeamName, members)
	if err != nil {
		writeDomainError(w, err)
		return
	}

	resp := dto.TeamAddResponse{
		Team: dto.TeamDto{
			TeamName: team.TeamName,
			Members:  make([]dto.TeamMemberDto, 0, len(team.Members)),
		},
	}

	for _, m := range team.Members {
		resp.Team.Members = append(resp.Team.Members, dto.TeamMemberDto{
			UserID:   m.UserID,
			Username: m.Username,
			IsActive: m.IsActive,
		})
	}

	writeJSON(w, http.StatusCreated, resp)
}

func (h *TeamHandlers) Get(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}

	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		writeBadRequest(w, "team_name is required")
		return
	}

	team, err := h.teamService.Get(r.Context(), teamName)
	if err != nil {
		writeDomainError(w, err)
		return
	}

	resp := dto.TeamGetResponse{
		TeamName: team.TeamName,
		Members:  make([]dto.TeamMemberDto, 0, len(team.Members)),
	}

	for _, m := range team.Members {
		resp.Members = append(resp.Members, dto.TeamMemberDto{
			UserID:   m.UserID,
			Username: m.Username,
			IsActive: m.IsActive,
		})
	}

	writeJSON(w, http.StatusOK, resp)
}
