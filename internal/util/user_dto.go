package dto

import (
	"collector-go/internal/entities"
	databasesqlc "collector-go/internal/sqlc"

	"github.com/jackc/pgx/v5/pgtype"
)

func UserToUserBd(u *entities.User) databasesqlc.CreateUserDbParams {
	return databasesqlc.CreateUserDbParams{
		FirstName: u.First_name,
		LastName:  pgtype.Text{String: u.Last_name, Valid: len(u.Last_name) > 0},
		FullName:  pgtype.Text{String: u.Full_name, Valid: len(u.Full_name) > 0},
		JobTitle:  pgtype.Text{String: u.Job_title, Valid: len(u.Job_title) > 0},
		Email:     u.Email,
		Password:  u.Password,
	}
}

func UpdateUserToUserBd(u *entities.User) databasesqlc.UpdateUserParams {
	return databasesqlc.UpdateUserParams{
		FirstName: u.First_name,
		LastName:  pgtype.Text{String: u.Last_name, Valid: len(u.Last_name) > 0},
		FullName:  pgtype.Text{String: u.Full_name, Valid: len(u.Full_name) > 0},
		JobTitle:  pgtype.Text{String: u.Job_title, Valid: len(u.Job_title) > 0},
		Email:     u.Email,
		ID:        u.Id,
	}
}

func EditPasswordToUserBd(u *entities.UserPasswordID) databasesqlc.EditPasswordParams {
	return databasesqlc.EditPasswordParams{
		Password: u.Password,
		ID:       u.Id,
	}
}
