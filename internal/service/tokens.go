package service

import (
	"collector-go/internal/entities"
	databasesqlc "collector-go/internal/sqlc"
	dto "collector-go/internal/util"
	"context"
	"fmt"
	"os"
	"strconv"

	logFi "github.com/gofiber/fiber/v2/log"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// @Summary		Check refresh token
// @Description	Check refresh token, if ok return access and refresh tokens and update refresh token
// @Tags		Tokens
// @Accept		json
// @Produce		json
// @Param		refresh_token	body		Tokens			true	"{refresh_token: ...}"
// @Success		200				{object}	Tokens
// @Failure		400				{object}	entities.ServerError
// @Failure		500				{object}	entities.ServerError
// @Router		/refresh		[post]
func CheckRefresh(pg *databasesqlc.Queries) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {

		var t *Tokens
		if err := c.BodyParser(&t); err != nil {
			logFi.Error("CheckRefresh")
			return c.Status(400).JSON(
				ReturnError(err.Error(), "error: can't parse body with refresh token"),
			)
		}

		token, err := jwt.Parse(t.RefreshToken, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("error: parse refresh token")
			}
			return []byte(os.Getenv("JWT_KEY_REFRESH")), nil
		})

		if err != nil {
			logFi.Error("CheckRefresh")
			return c.Status(500).JSON(
				ReturnError(err.Error(), "error: refresh token has expired or token signature is invalid, return to login"),
			)
		}

		claims, ok := token.Claims.(jwt.MapClaims)

		if !ok && !token.Valid {
			return fmt.Errorf("error get user claims from refresh token")
		}

		userID := claims["userID"].(string)
		id := uuid.MustParse(userID)

		tokenBd, err := pg.IsCreatedRefreshTokenDb(context.Background(), id)
		if err != nil {
			logFi.Error("CheckRefresh")
			return c.Status(500).JSON(
				ReturnError(err.Error(), "error: can't find refresh token from DB"),
			)
		}

		if tokenBd != t.RefreshToken {
			return c.Status(500).JSON(
				ReturnError("error: the refresh token is fake", "error: the refresh token is fake"),
			)
		}

		accessToken, err := CreateAccessToken(c, userID, os.Getenv("JWT_KEY_ACCESS"))
		if err != nil {
			logFi.Error("CheckRefresh")
			return c.Status(500).JSON(
				ReturnError(err.Error(), "error: can't CreateAccessToken"),
			)
		}

		refreshToken, err := CreateRefreshToken(c, userID)
		if err != nil {
			logFi.Error("CheckRefresh")
			return c.Status(500).JSON(
				ReturnError(err.Error(), "error: can't CreateRefreshToken"),
			)
		}

		tok_ref := &entities.Token{
			Token:   refreshToken,
			User_id: id,
		}

		pg.UpdateRefreshTokenDb(context.Background(), dto.UpdateTokenToTokenDb(tok_ref))

		return c.Status(200).JSON(Tokens{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		})
	}
}

// @Summary		Get all exists tokens
// @Description	Get all exists tokens
// @Tags		Tokens
// @Accept		json
// @Produce		json
// @Success		200		{array}		entities.Token
// @Failure		500		{object}	entities.ServerError
// @Router		/tokens	[get]
func GetTokens(pg *databasesqlc.Queries) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {

		users_db, err := pg.GetRefreshTokensDb(context.Background())
		if err != nil {
			logFi.Error("GetTokens")
			return c.Status(500).JSON(
				ReturnError(err.Error(), "error: can't get all refresh Tokens"),
			)
		}

		return c.Status(200).JSON(users_db)
	}
}

// @Summary		Get exist token
// @Description	Get exist token
// @Tags		Tokens
// @Accept		json
// @Produce		json
// @Security	JWT
// @Param		id			path		string		true	"user_id:uuid"
// @Success		200			{object}	entities.TokenSwagger
// @Failure		500			{object}	entities.ServerError
// @Router		/token/{id}	[get]
func GetTokenByUserId(pg *databasesqlc.Queries) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		if c.Params("id") == "" {
			logFi.Error("GetTokenByUserId")
			return c.Status(500).JSON(
				ReturnError("error: id can't be empty in GetTokenByUserId", "error: id can't be empty in GetTokenByUserId"),
			)
		}

		user_id, err := uuid.Parse(c.Params("id"))
		if err != nil {
			logFi.Error("GetTokenByUserId")
			return c.Status(500).JSON(
				ReturnError(err.Error(), "error: can't parse uuid in GetTokenByUserId"),
			)
		}

		db_token, err := pg.IsCreatedRefreshTokenDb(context.Background(), user_id)
		if err != nil {
			logFi.Error("GetTokenByUserId")
			return c.Status(500).JSON(
				ReturnError(err.Error(), "error: can't find created refresh Token in GetTokenByUserId"),
			)
		}

		return c.Status(200).JSON(entities.ServerOk{
			Message: fmt.Sprintf("Token: %s", db_token),
		})
	}
}

// @Summary		Delete token by id
// @Description	Delete token by id
// @Tags		Tokens
// @Accept		json
// @Produce		json
// @Security	JWT
// @Param		id				path		int		true	"id: 2"
// @Success		200				{object}	entities.ServerOk
// @Failure		500				{object}	entities.ServerError
// @Router		/token/{id}		[delete]
func DeleteTokenById(pg *databasesqlc.Queries) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {

		if c.Params("id") == "" {
			logFi.Error("DeleteTokenById")
			return c.Status(500).JSON(
				ReturnError("error: id can't be empty in GetTokenByUserId", "error:id can't be empty in DeleteTokenById"),
			)
		}

		id, err := strconv.Atoi(c.Params("id"))
		if err != nil {
			logFi.Error("DeleteTokenById")
			return c.Status(500).JSON(
				ReturnError(err.Error(), "error: can't convert string to int in DeleteTokenById"),
			)
		}

		err = pg.DeleteRefreshTokenDb(context.Background(), int32(id))
		if err != nil {
			logFi.Error("DeleteTokenById")
			return c.Status(500).JSON(
				ReturnError(err.Error(), "error: can't delete refresh Token in DeleteTokenById"),
			)
		}

		return c.Status(200).JSON(entities.ServerOk{
			Message: fmt.Sprintf("token successful deleted id: %d", id),
		})
	}
}
