package user

import "github.com/google/uuid"

type CreateUserRequest struct {
	Email    string `json:"email" validate:"required,email,size:254"`
	Name     string `json:"name" validate:"required,size:254"`
	Password string `json:"password" validate:"required,min:6,max:254"`
}

type UserResponse struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Email string    `json:"email"`
}
