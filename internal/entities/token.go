package entities

import "github.com/google/uuid"

type Token struct {
	Id      int       `json:"id"`
	Token   string    `json:"token"`
	User_id uuid.UUID `json:"user_id"`
}

type TokenSwagger struct {
	Token string `json:"token"`
}
