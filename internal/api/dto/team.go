package dto

// /team/add

type TeamAddRequest struct {
	TeamName string          `json:"team_name" validate:"required"`
	Members  []TeamMemberDto `json:"members"   validate:"required,dive"`
}

type TeamMemberDto struct {
	UserID   string `json:"user_id" validate:"required"`
	Username string `json:"username" validate:"required"`
	IsActive bool   `json:"is_active"`
}

type TeamAddResponse struct {
	Team TeamDto `json:"team"`
}

// /team/get

type TeamGetResponse struct {
	TeamName string          `json:"team_name"`
	Members  []TeamMemberDto `json:"members"`
}

type TeamDto struct {
	TeamName string          `json:"team_name"`
	Members  []TeamMemberDto `json:"members"`
}
