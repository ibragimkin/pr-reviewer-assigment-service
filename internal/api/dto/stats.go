package dto

type ReviewerStatDto struct {
	UserID      string `json:"user_id"`
	ReviewCount int    `json:"review_count"`
}

type ReviewerStatsResponse struct {
	Items []ReviewerStatDto `json:"items"`
}
