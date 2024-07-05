package dto

import (
	"collector-go/internal/entities"
	databasesqlc "collector-go/internal/sqlc"
)

func UpdateTokenToTokenDb(t *entities.Token) databasesqlc.UpdateRefreshTokenDbParams {
	return databasesqlc.UpdateRefreshTokenDbParams{
		UserID: t.User_id,
		Token:  t.Token,
	}
}

func SaveTokenToTokenDb(t *entities.Token) databasesqlc.SaveRefreshTokenParams {
	return databasesqlc.SaveRefreshTokenParams{
		UserID: t.User_id,
		Token:  t.Token,
	}
}
