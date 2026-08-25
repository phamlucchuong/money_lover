package user

import "github.com/google/uuid"

type CreateUserRequest struct {
	Email    string `json:"email" validate:"required,email,size:254"`
	Name     string `json:"name" validate:"required,size:254"`
	Password string `json:"password" validate:"required,min:6,max:72"`
}

type UpdateUserRequest struct {
	Email    string `json:"email" validate:"omitempty,email,size:254"`
	Name     string `json:"name" validate:"omitempty,size:254"`
	Password string `json:"password" validate:"omitempty,min:6,max:72"`
}

type UserResponse struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Email string    `json:"email"`
}
