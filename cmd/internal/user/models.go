package user

import "time"

// User representa um usuário do sistema.
type User struct {
	ID        int64     `json:"id"`
	Subject   string    `json:"subject"` // identificador externo (ex: OIDC sub)
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Active    bool      `json:"active"`
	Picture   string    `json:"picture"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// ------------------------------------------- REQUESTS ------------------------------------------- //

type GrantPermissionsRequest struct {
	Email       string   `json:"email"`
	Permissions []string `json:"permissions"`
}

// ------------------------------------------- RESPONSES ------------------------------------------- //

type GrantPermissionsResponseItem struct {
	Permission string
	Granted    bool
}

type GrantPermissionsResponse struct {
	Total     int32                          `json:"total"`
	Successes int32                          `json:"successes"`
	Failures  int32                          `json:"failures"`
	Results   []GrantPermissionsResponseItem `json:"results"`
}
