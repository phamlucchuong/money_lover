package auth

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type RegisterRequest struct {
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type AuthResponse struct {
	Authenticated bool   `json:"authenticated"`
	Token         string `json:"token"`
	RefreshToken  string `json:"refresh_token"`
}

type RefreshTokenRequest struct {
	Token        string `json:"token" validate:"required"`
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type UserContext struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}
