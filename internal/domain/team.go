package domain

// TeamMember описывает участника команды в контексте Team.
type TeamMember struct {
	UserID   string `json:"user_id"`   // Идентификатор пользователя
	Username string `json:"username"`  // Имя пользователя
	IsActive bool   `json:"is_active"` // Флаг активности
}

// Team описывает команду и её участников.
type Team struct {
	TeamName string       `json:"team_name"` // Уникальное имя команды
	Members  []TeamMember `json:"members"`   // Участники команды
}
