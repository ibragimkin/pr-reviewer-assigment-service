package domain

import "time"

// PullRequestStatus описывает статус PR: OPEN или MERGED.
type PullRequestStatus string

const (
	StatusOpen   PullRequestStatus = "OPEN"
	StatusMerged PullRequestStatus = "MERGED"
)

// PullRequest описывает полный PR.
type PullRequest struct {
	PullRequestID     string     `json:"pull_request_id"`    // ID PR
	PullRequestName   string     `json:"pull_request_name"`  // Название PR
	AuthorID          string     `json:"author_id"`          // ID автора
	Status            string     `json:"status"`             // Статус PR (OPEN/MERGED)
	AssignedReviewers []string   `json:"assigned_reviewers"` // Список user_id назначенных ревьюверов (0-2)
	CreatedAt         *time.Time `json:"created_at"`         // Время создания PR
	MergedAt          *time.Time `json:"merged_at"`          // Время слияния PR
}

// PullRequestShort - сокращённая версия PR
type PullRequestShort struct {
	PullRequestID   string `json:"pull_request_id"`   // ID PR
	PullRequestName string `json:"pull_request_name"` // Название PR
	AuthorID        string `json:"author_id"`         // ID автора
	Status          string `json:"status"`            // OPEN/MERGED
}
