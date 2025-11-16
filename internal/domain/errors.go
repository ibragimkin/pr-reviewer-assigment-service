package domain

// ErrorCode —  код ошибки.
type ErrorCode string

const (
	ErrorTeamExists  ErrorCode = "TEAM_EXISTS"
	ErrorPRExists    ErrorCode = "PR_EXISTS"
	ErrorPRMerged    ErrorCode = "PR_MERGED"
	ErrorNotAssigned ErrorCode = "NOT_ASSIGNED"
	ErrorNoCandidate ErrorCode = "NO_CANDIDATE"
	ErrorNotFound    ErrorCode = "NOT_FOUND"
)

// Error структура для проброса ошибок из домена.
type Error struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
}

func (e *Error) Error() string {
	return string(e.Code) + ": " + e.Message
}

// NewError - помогает создавать ошибки удобнее.
func NewError(code ErrorCode, message string) *Error {
	return &Error{Code: code, Message: message}
}
