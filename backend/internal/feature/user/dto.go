package user

import "github.com/google/uuid"

type UpdateUserRequest struct {
	Email    string `json:"email" validate:"omitempty,email,max=254"`
	Name     string `json:"name" validate:"omitempty,max=254"`
	Password string `json:"password" validate:"omitempty,min=6,max=72"`
}

type UserResponse struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Email string    `json:"email"`
}
