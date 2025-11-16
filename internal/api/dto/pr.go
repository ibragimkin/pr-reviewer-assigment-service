package dto

//  /pullRequest/create

type PullRequestCreateRequest struct {
	PullRequestID   string `json:"pull_request_id" validate:"required"`
	PullRequestName string `json:"pull_request_name" validate:"required"`
	AuthorID        string `json:"author_id" validate:"required"`
}

type PullRequestCreateResponse struct {
	PR PullRequestDto `json:"pr"`
}

// /pullRequest/merge

type PullRequestMergeRequest struct {
	PullRequestID string `json:"pull_request_id" validate:"required"`
}

type PullRequestMergeResponse struct {
	PR PullRequestDto `json:"pr"`
}

//  /pullRequest/reassign

type PullRequestReassignRequest struct {
	PullRequestID string `json:"pull_request_id" validate:"required"`
	OldUserID     string `json:"old_user_id" validate:"required"`
}

type PullRequestReassignResponse struct {
	PR         PullRequestDto `json:"pr"`
	ReplacedBy string         `json:"replaced_by"`
}

//

type PullRequestDto struct {
	PullRequestID     string   `json:"pull_request_id"`
	PullRequestName   string   `json:"pull_request_name"`
	AuthorID          string   `json:"author_id"`
	Status            string   `json:"status"`
	AssignedReviewers []string `json:"assigned_reviewers"`
	CreatedAt         *string  `json:"createdAt,omitempty"`
	MergedAt          *string  `json:"mergedAt,omitempty"`
}

type PullRequestShort struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID        string `json:"author_id"`
	Status          string `json:"status"`
}
