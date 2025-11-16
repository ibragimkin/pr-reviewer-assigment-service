package httphandlers

import (
	"net/http"

	"pr-reviewer-assigment-service/internal/api/dto"
	"pr-reviewer-assigment-service/internal/application/service"
)

type StatsHandlers struct {
	statsService *service.StatsService
}

func NewStatsHandlers(statsService *service.StatsService) *StatsHandlers {
	return &StatsHandlers{statsService: statsService}
}

func (h *StatsHandlers) GetReviewerStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}

	stats, err := h.statsService.GetReviewerStats(r.Context())
	if err != nil {
		writeDomainError(w, err)
		return
	}

	items := make([]dto.ReviewerStatDto, 0, len(stats))
	for _, s := range stats {
		items = append(items, dto.ReviewerStatDto{
			UserID:      s.UserID,
			ReviewCount: s.ReviewCount,
		})
	}

	resp := dto.ReviewerStatsResponse{Items: items}
	writeJSON(w, http.StatusOK, resp)
}
