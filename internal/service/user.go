package service

import (
	"bytes"
	"collector-go/internal/entities"
	databasesqlc "collector-go/internal/sqlc"
	dto "collector-go/internal/util"
	"context"
	"html/template"
	"net/smtp"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	logFi "github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
	"github.com/oklog/ulid/v2"
)

// @Summary		Create user
// @Description	Create user
// @Tags		Users
// @Accept		json
// @Produce		json
// @Param		user			body		entities.UserRegistrationSwagger	true	"{first_name: Сергей,last_name: Николаевич, email: name@gmail.com, password: a-zA-Z0-9_ ,'~!@#$%^&*()?-+={}<>|:;[]} (doble quotes also included)"
// @Success		200				{object}	entities.User
// @Failure		400				{object}	entities.ServerError
// @Failure		500				{object}	entities.ServerError
// @Router		/registration	[post]
func CreateUser(pg *databasesqlc.Queries) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var u entities.User
		if err := c.BodyParser(&u); err != nil {
			logFi.Error("CreateUser")
			return c.Status(400).JSON(
				ReturnError(err.Error(), "error: can't parse request body"),
			)
		}

		if err := u.ValidateCreateUser(); err != nil {
			logFi.Error("CreateUser")
			return c.Status(400).JSON(
				ReturnError(err.Error(), err.Error()),
			)
		}
		u.Email = strings.ToLower(u.Email)
		u.Password = GetHashPassword(u.Password)
		u_db, err := pg.CreateUserDb(context.Background(), dto.UserToUserBd(&u))
		if err != nil {
			logFi.Error("CreateUser")
			return c.Status(500).JSON(
				ReturnError(err.Error(), "error: can't create user"),
			)
		}

		return c.Status(200).JSON(u_db)
	}
}

// @Summary		Login user by email and queries password, if ok return access and refresh tokens
// @Description	Check email and password user.
// @Tags		Users
// @Accept		json
// @Produce		json
// @Param		user	body		entities.UserLoginSwagger	true	"{email: example@gmail.com, password: a-zA-Z0-9_ ,'~!@#$%^&*()?-+={}<>|:;[]} (doble quotes also included)"
// @Success		200		{object}	Tokens
// @Failure		400		{object}	entities.ServerError
// @Failure		500		{object}	entities.ServerError
// @Router		/login	[post]
func CheckLogin(pg *databasesqlc.Queries) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {

		var u entities.User
		if err := c.BodyParser(&u); err != nil {
			logFi.Error("CheckLogin")
			return c.Status(400).JSON(
				ReturnError(err.Error(), "error: can't parse body"),
			)
		}

		if err := u.ValidateLogin(); err != nil {
			logFi.Error("CheckLogin")
			return c.Status(400).JSON(
				ReturnError(err.Error(), "error: can't ValidateLogin"),
			)
		}

		u.Email = strings.ToLower(u.Email)

		udb, err := pg.GetUserByEmail(context.Background(), u.Email)
		if err != nil {
			logFi.Error("CheckLogin")
			return c.Status(500).JSON(
				ReturnError(err.Error(), "error: can't GetUserByEmail"),
			)
		}

		if udb.IsDeleted.Bool {
			logFi.Error("CheckLogin")
			return c.Status(500).JSON(
				ReturnError("error: can't deleted user in CheckLogin", "error: can't deleted user"),
			)
		}

		if !ComparePassword(u.Password, udb.Password) {
			logFi.Error("CheckLogin")
			return c.Status(500).JSON(
				ReturnError("error: can't there is no such password in CheckLogin", "error: can't there is no such password"),
			)
		}

		id := udb.ID.String()
		tokens, err := CreateAccessRefreshTokens(pg, c, id)
		if err != nil {
			logFi.Error("CheckLogin")
			return c.Status(500).JSON(
				ReturnError(err.Error(), "error: can't create access or refresh tokens in CheckLogin"),
			)
		}

		return c.Status(200).JSON(tokens)
	}
}

// @Summary		Get all exists users
// @Description	Get all exists users
// @Tags		Users
// @Accept		json
// @Produce		json
// @Success		200		{array}		entities.User
// @Failure		500		{object}	entities.ServerError
// @Router		/users	[get]
func GetUsers(pg *databasesqlc.Queries) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {

		users_db, err := pg.GetUsersDb(context.Background())
		if err != nil {
			logFi.Error("GetUsers")
			return c.Status(500).JSON(
				ReturnError(err.Error(), "error: can't get users"),
			)
		}

		return c.Status(200).JSON(users_db)
	}
}

// @Summary		Change user properties
// @Description	Change user properties
// @Tags		Users
// @Accept		json
// @Produce		json
// @Security	JWT
// @Param 		Authorization 	header 		string						true 	"Insert your access token" default(Bearer <Add access token here>)
// @Param		user			body		entities.UserUpdateSwagger	true	"{first_name: Сергей,last_name: Николаевич, full_name:Попов,email: name@gmail.com, job_title: manager }"
// @Success		200				{object}	entities.User
// @Failure		400				{object}	entities.ServerError
// @Failure		500				{object}	entities.ServerError
// @Router		/user			[patch]
func UpdateUserById(pg *databasesqlc.Queries) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {

		var u entities.User
		var err error
		if err = c.BodyParser(&u); err != nil {
			logFi.Error("UpdateUserById")
			return c.Status(500).JSON(
				ReturnError(err.Error(), "error: can't parse request body"),
			)
		}
		if err = u.ValidateUpdate(); err != nil {
			logFi.Error("UpdateUserById")
			return c.Status(400).JSON(
				ReturnError(err.Error(), err.Error()),
			)
		}

		userID, err := GetUserIdFromHeader(c)
		if err != nil {
			logFi.Error("UpdateUserById")
			return c.Status(500).JSON(
				ReturnError(err.Error(), "error: can't find userId"),
			)
		}

		u.Id, err = uuid.Parse(userID)
		if err != nil {
			logFi.Error("UpdateUserById")
			return c.Status(500).JSON(
				ReturnError(err.Error(), "error: can't parse userId"),
			)
		}

		u_db, err := pg.UpdateUser(context.Background(), dto.UpdateUserToUserBd(&u))
		if err != nil {
			logFi.Error("UpdateUserById")
			return c.Status(400).JSON(
				ReturnError(err.Error(), "error: can't update user"),
			)
		}

		return c.Status(200).JSON(u_db)
	}
}

// @Summary		Logical delete user
// @Description	Logical delete user
// @Tags		Users
// @Accept		json
// @Produce		json
// @Security	JWT
// @Param 		Authorization 		header 		string			true 	"Insert your access token" default(Bearer <Add access token here>)
// @Param		user				body		entities.UserDeleteSwagger	true	"{id: uuid.UUID}"
// @Success		200					{object}	entities.User
// @Failure		500					{object}	entities.ServerError
// @Failure		500					{object}	entities.ServerError
// @Router		/user				[delete]
func DeleteUserById(pg *databasesqlc.Queries) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {

		var u entities.User
		if err := c.BodyParser(&u); err != nil {
			logFi.Error("DeleteUserById")
			return c.Status(500).JSON(
				ReturnError(err.Error(), "error: can't parse request body"),
			)
		}

		user_db, err := pg.DeleteUserDb(context.Background(), u.Id)
		if err != nil {
			logFi.Error("DeleteUserById")
			return c.Status(500).JSON(
				ReturnError(err.Error(), "error: can't delete user"),
			)
		}

		return c.Status(200).JSON(user_db)
	}
}

// @Summary		Change password user
// @Description	Change password user
// @Tags		Users
// @Accept		json
// @Produce		json
// @Security	JWT
// @Param 		Authorization 	header 		string							true 	"Insert your access token" default(Bearer <Add access token here>)
// @Param		user			body		entities.UserPasswordSwagger	true	"{password: a-zA-Z0-9_ ,'~!@#$%^&*()?-+={}<>|:;[]} (doble quotes also included)"
// @Success		200				{object}	entities.ServerOk
// @Failure		400				{object}	entities.ServerError
// @Failure		500				{object}	entities.ServerError
// @Router		/change_password	[patch]
func EditPasswordUser(pg *databasesqlc.Queries) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {

		var err error
		var u entities.UserPasswordSwagger
		if err = c.BodyParser(&u); err != nil {
			logFi.Error("EditPasswordUser")
			return c.Status(500).JSON(
				ReturnError(err.Error(), "error: can't parse request body"),
			)
		}

		if err = u.ValidatePassword(); err != nil {
			logFi.Error("EditPasswordUser")
			return c.Status(400).JSON(
				ReturnError(err.Error(), err.Error()),
			)
		}

		userID, err := GetUserIdFromHeader(c)
		if err != nil {
			logFi.Error("EditPasswordUser")
			return c.Status(500).JSON(
				ReturnError(err.Error(), "error: can't find userId"),
			)
		}

		var user entities.UserPasswordID
		user.Id, err = uuid.Parse(userID)
		if err != nil {
			logFi.Error("EditPasswordUser")
			return c.Status(500).JSON(
				ReturnError(err.Error(), "error: can't parse userID"),
			)
		}

		user.Password = GetHashPassword(u.Password)
		if err := pg.EditPassword(context.Background(), dto.EditPasswordToUserBd(&user)); err != nil {
			logFi.Error("EditPasswordUser")
			return c.Status(500).JSON(
				ReturnError(err.Error(), "error: can't edit password"),
			)
		}

		return c.Status(200).JSON(entities.ServerOk{
			Message: "password changed"},
		)
	}
}

// @Summary		Change password user by old password
// @Description	Change password user by old password
// @Tags		Users
// @Accept		json
// @Produce		json
// @Security	JWT
// @Param 		Authorization 	header 		string							true 	"Insert your access token" default(Bearer <Add access token here>)
// @Param		user			body		entities.UserPasswordOldNewSwagger	true	"{password: a-zA-Z0-9_ ,'~!@#$%^&*()?-+={}<>|:;[]} (doble quotes also included)"
// @Success		200				{object}	entities.ServerOk
// @Failure		400				{object}	entities.ServerError
// @Failure		500				{object}	entities.ServerError
// @Router		/change_password_old	[patch]
func EditPasswordUserByOld(pg *databasesqlc.Queries) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {

		var err error
		var u entities.UserPasswordOldNewSwagger
		if err = c.BodyParser(&u); err != nil {
			logFi.Error("EditPasswordUserByOld")
			return c.Status(500).JSON(
				ReturnError(err.Error(), "error: can't parse request body in EditPasswordUserByOld"),
			)
		}

		if err = u.ValidateOldNewPassword(); err != nil {
			logFi.Error("EditPasswordUser")
			return c.Status(400).JSON(
				ReturnError(err.Error(), err.Error()),
			)
		}

		userID, err := GetUserIdFromHeader(c)
		if err != nil {
			logFi.Error("EditPasswordUserByOld")
			return c.Status(500).JSON(
				ReturnError(err.Error(), "error: can't find userId in EditPasswordUserByOld"),
			)
		}

		var user entities.UserPasswordID
		user.Id, err = uuid.Parse(userID)
		if err != nil {
			logFi.Error("EditPasswordUserByOld")
			return c.Status(500).JSON(
				ReturnError(err.Error(), "error: can't parse userID in EditPasswordUserByOld"),
			)
		}

		db_user, err := pg.GetUserByIdDb(context.Background(), user.Id)
		if err != nil {
			logFi.Error("EditPasswordUserByOld")
			return c.Status(500).JSON(
				ReturnError(err.Error(), "error: can't get user by id in EditPasswordUserByOld"),
			)
		}

		if db_user.IsDeleted.Bool {
			logFi.Error("EditPasswordUserByOld")
			return c.Status(500).JSON(
				ReturnError("error: user deleted in EditPasswordUserByOld", "error: user deleted in EditPasswordUserByOld"),
			)
		}

		if !ComparePassword(u.PasswordOld, db_user.Password) {
			logFi.Error("EditPasswordUserByOld")
			return c.Status(500).JSON(
				ReturnError("error: there is no such password in EditPasswordUserByOld", "error: there is no such password in EditPasswordUserByOld"),
			)
		}

		user.Password = GetHashPassword(u.PasswordNew)
		if err := pg.EditPassword(context.Background(), dto.EditPasswordToUserBd(&user)); err != nil {
			logFi.Error("EditPasswordUserByOld")
			return c.Status(500).JSON(
				ReturnError(err.Error(), "error: can't edit password in EditPasswordUserByOld"),
			)
		}

		return c.Status(200).JSON(entities.ServerOk{
			Message: "old password changed to new password"},
		)

	}
}

// @Summary		Get exist user
// @Description	Get exist user
// @Tags		Users
// @Accept		json
// @Produce		json
// @Security	JWT
// @Param 		Authorization 	header 		string			true 	"Insert your access token" default(Bearer <Add access token here>)
// @Param		id				path		string			true	"user_id: uuid"
// @Success		200				{object}	entities.User
// @Failure		500				{object}	entities.ServerError
// @Router		/user/{id}		[get]
func GetUserById(pg *databasesqlc.Queries) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		if c.Params("id") == "" {
			logFi.Error("GetUserById")
			return c.Status(500).JSON(
				ReturnError("error: id can't be empty in GetUserById", "error: id can't be empty"),
			)
		}

		user_id, err := uuid.Parse(c.Params("id"))
		if err != nil {
			logFi.Error("GetUserById")
			return c.Status(500).JSON(
				ReturnError(err.Error(), "error: can't parse uuid"),
			)
		}

		db_user, err := pg.UserByIdDb(context.Background(), user_id)
		if err != nil {
			logFi.Error("GetUserById")
			return c.Status(500).JSON(
				ReturnError(err.Error(), "error: can't get UserById"),
			)
		}

		return c.Status(200).JSON(&db_user)
	}
}

// @Summary		Check email user
// @Description	Check email user, get id, send link to email user, set id to redis
// @Tags		Users
// @Accept		json
// @Produce		json
// @Param		email			body		entities.UserEmail	true	"email: name@gmail.com"
// @Success		200				{object}	entities.ServerOk
// @Failure		400				{object}	entities.ServerError
// @Failure		500				{object}	entities.ServerError
// @Router		/user/new_pass	[post]
func CreateRestorePasswordLink(pg *databasesqlc.Queries, opt *redis.Client) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {

		var u *entities.UserEmail

		if err := c.BodyParser(&u); err != nil {
			logFi.Error("CreateRestorePasswordLink")
			return c.Status(500).JSON(
				ReturnError(err.Error(), "error: can't parse body"),
			)
		}

		email := strings.ToLower(u.Email)

		err := entities.ValidateEmail(email)
		if err != nil {
			logFi.Error("VerificationEmail")
			return c.Status(500).JSON(
				ReturnError(err.Error(), err.Error()),
			)
		}

		db_user, err := pg.GetUserByEmail(context.Background(), email)
		if err != nil {
			logFi.Error("CreateRestorePasswordLink")
			return c.Status(500).JSON(
				ReturnError(err.Error(), "error: can't get UserByEmail"),
			)
		}

		app_email := os.Getenv("APP_EMAIL")
		app_password := os.Getenv("APP_PASSWORD")
		smtpServer := os.Getenv("SMTP_SERVER")
		smtpPort := os.Getenv("SMTP_PORT")

		to := db_user.Email
		auth := smtp.PlainAuth("", app_email, app_password, smtpServer)

		subject := "Change password Link ForSales"

		t, err := template.ParseFiles("changepass.html")
		if err != nil {
			logFi.Error("CreateRestorePasswordLink")
			return c.Status(500).JSON(
				ReturnError(err.Error(), "error: can't parse file changepass.html"),
			)
		}

		token := ulid.Make().String()

		buff := new(bytes.Buffer)
		t.Execute(buff, struct {
			Base_url  string
			Token     string
			User_name string
		}{
			Base_url:  os.Getenv("SERVER_HOST"),
			Token:     token,
			User_name: db_user.FirstName,
		})

		mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"

		msg := []byte("From: " + app_email + "\n" +
			"To: " + to + "\n" + "Subject: " + subject +
			"\n" + mime + buff.String())

		err = smtp.SendMail(smtpServer+":"+smtpPort, auth, app_email, []string{to}, []byte(msg))
		if err != nil {
			logFi.Error("CreateRestorePasswordLink")
			return c.Status(500).JSON(
				ReturnError(err.Error()+"error: send message failed", "error: send message failed"),
			)
		}

		value := db_user.ID.String()

		exp_time, _ := strconv.Atoi(os.Getenv("REDIS_RESTORE_PASSWORD_EXP_TIME"))
		err = opt.Set(context.Background(), token, value, time.Duration(exp_time)*time.Second).Err()
		if err != nil {
			logFi.Error("CreateRestorePasswordLink")
			ReturnError(err.Error(), "error: can't set value into redis")
		}

		return c.Status(200).JSON(entities.ServerOk{
			Message: "ok, verification token sended"},
		)
	}
}

// @Summary		Update password
// @Description	Get token, new password. Get user_id from redis. Update hash password.
// @Tags		Users
// @Accept		json
// @Produce		json
// @Param		user			body		entities.UserPasswordToken	true	"{password: a-zA-Z0-9_ ,'~!@#$%^&*()?-+={}<>|:;[] (doble quotes also included), token:...}"
// @Success		200				{object}	entities.ServerOk
// @Failure		400				{object}	entities.ServerError
// @Failure		500				{object}	entities.ServerError
// @Router		/user/change_pass	[patch]
func ChangePasswordByLink(pg *databasesqlc.Queries, opt *redis.Client) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {

		var p *entities.UserPasswordToken

		if err := c.BodyParser(&p); err != nil {
			logFi.Error("ChangePasswordByLink")
			return c.Status(500).JSON(
				ReturnError(err.Error(), "error: can't parse body in ChangePasswordByLink"),
			)
		}

		user_id, err := opt.Get(context.Background(), p.Token).Result()
		if err != nil {
			logFi.Errorf("ChangePasswordByLink: %s", err)
			return c.Status(500).JSON(
				ReturnError(err.Error(), "error: token expired or can't find in ChangePasswordByLink"),
			)
		}

		id, err := uuid.Parse(user_id)
		if err != nil {
			logFi.Error("ChangePasswordByLink")
			return c.Status(500).JSON(
				ReturnError(err.Error(), "error: can't parse user_id to uuid in ChangePasswordByLink"),
			)
		}

		var u entities.UserPasswordID
		u.Password = GetHashPassword(p.Password)
		u.Id = id
		if err := pg.EditPassword(context.Background(), dto.EditPasswordToUserBd(&u)); err != nil {
			logFi.Error("ChangePasswordByLink")
			return c.Status(500).JSON(
				ReturnError(err.Error(), "error: can't restore Password in ChangePasswordByLink"),
			)
		}

		return c.Status(200).JSON(entities.ServerOk{
			Message: "ok, password restored"},
		)
	}
}
