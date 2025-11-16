package dto

// /users/setIsActive

type UserSetIsActiveRequest struct {
	UserID   string `json:"user_id" validate:"required"`
	IsActive bool   `json:"is_active"`
}

type UserResponse struct {
	User UserDto `json:"user"`
}

type UserDto struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	TeamName string `json:"team_name"`
	IsActive bool   `json:"is_active"`
}

//  /users/getReview

type UserGetReviewResponse struct {
	UserID       string                `json:"user_id"`
	PullRequests []PullRequestShortDto `json:"pull_requests"`
}

type PullRequestShortDto struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID        string `json:"author_id"`
	Status          string `json:"status"`
}
