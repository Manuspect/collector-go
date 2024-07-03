// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0

package databasesqlc

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type Token struct {
	ID     int32     `json:"id"`
	UserID uuid.UUID `json:"user_id"`
	Token  string    `json:"token"`
}

type User struct {
	ID         uuid.UUID          `json:"id"`
	FirstName  string             `json:"first_name"`
	LastName   pgtype.Text        `json:"last_name"`
	FullName   pgtype.Text        `json:"full_name"`
	Email      string             `json:"email"`
	Password   string             `json:"password"`
	JobTitle   pgtype.Text        `json:"job_title"`
	IsDeleted  pgtype.Bool        `json:"is_deleted"`
	CreateDate pgtype.Timestamptz `json:"create_date"`
	UpdateDate pgtype.Timestamptz `json:"update_date"`
}
