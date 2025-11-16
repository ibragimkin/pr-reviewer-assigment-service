package domain

type User struct {
	UserID   string `json:"user_id"`   // Идентификатор пользователя
	Username string `json:"username"`  // Имя пользователя
	TeamName string `json:"team_name"` // Название команды
	IsActive bool   `json:"is_active"` // Статус активности пользователя
}
