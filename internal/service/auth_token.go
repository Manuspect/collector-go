package service

import (
	"collector-go/internal/entities"
	databasesqlc "collector-go/internal/sqlc"
	dto "collector-go/internal/util"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/gofiber/fiber/v2"
	logFi "github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
)

type Tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func CreateAccessRefreshTokens(pg *databasesqlc.Queries, c *fiber.Ctx, id string) (*Tokens, error) {

	accessToken, err := CreateAccessToken(c, id, os.Getenv("JWT_KEY_ACCESS"))
	if err != nil {
		logFi.Error("CreateAccessRefreshTokens")
		return nil, c.Status(500).JSON(
			ReturnError(err.Error(), "error: can't create access token in CreateAccessRefreshTokens"),
		)
	}

	refreshToken, err := CreateRefreshToken(c, id)
	if err != nil {
		logFi.Error("CreateAccessRefreshTokens")
		return nil, c.Status(500).JSON(
			ReturnError(err.Error(), "error: can't create refresh token in CreateAccessRefreshTokens"),
		)
	}

	tokens := &Tokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	var t entities.Token
	t.Token = refreshToken
	t.User_id = uuid.MustParse(id)

	_, err = pg.IsCreatedRefreshTokenDb(context.Background(), t.User_id)
	if err != nil {
		err = nil
		_, err := pg.SaveRefreshToken(context.Background(), dto.SaveTokenToTokenDb(&t))
		if err != nil {
			logFi.Error("CreateAccessRefreshTokens")
			return nil, c.Status(500).JSON(
				ReturnError(err.Error(), "error: can't save refresh token Db in CreateAccessRefreshTokens"),
			)
		}
	} else {

		if err := pg.UpdateRefreshTokenDb(context.Background(), dto.UpdateTokenToTokenDb(&t)); err != nil {
			logFi.Error("CreateAccessRefreshTokens")
			return nil, c.Status(500).JSON(
				ReturnError(err.Error(), "error: can't update refresh token Db in CreateAccessRefreshTokens"),
			)
		}
	}

	return tokens, err

}

func CreateAccessToken(c *fiber.Ctx, id, secret string) (accessToken string, err error) {

	exp_time, _ := strconv.Atoi(os.Getenv("JWT_ACCESS_DURATION"))
	claims := jwt.MapClaims{
		"userID": id,
		"exp":    time.Now().Add(time.Duration(exp_time) * time.Second).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	t, err := token.SignedString([]byte(secret))
	if err != nil {
		return t, c.Status(500).JSON(
			ReturnError(err.Error(), "error: can't create access Token"),
		)
	}
	return t, err
}

func CreateRefreshToken(c *fiber.Ctx, id string) (refreshToken string, err error) {
	secret := os.Getenv("JWT_KEY_REFRESH")

	exp_time, _ := strconv.Atoi(os.Getenv("JWT_REFRESH_DURATION"))
	claims := jwt.MapClaims{
		"userID": id,
		"exp":    time.Now().Add(time.Duration(exp_time) * time.Second).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	t, err := token.SignedString([]byte(secret))
	if err != nil {
		return t, c.Status(500).JSON(
			ReturnError(err.Error(), "error: can't create refresh token"),
		)
	}
	return t, err
}

func GetUserIdFromHeader(c *fiber.Ctx) (string, error) {

	type auth struct {
		Authorization []string
	}

	var t auth
	err := c.ReqHeaderParser(&t)
	if err != nil {
		log.Println(err)
		return "", err
	}

	val := strings.Split(t.Authorization[0], " ")

	tok, _ := jwt.Parse(val[1], func(token *jwt.Token) (interface{}, error) {
		return []byte(""), nil
	})

	claims := tok.Claims.(jwt.MapClaims)

	userID := fmt.Sprintf("%v", claims["userID"])

	return userID, nil
}
