package entities

import (
	"regexp"
	"time"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"github.com/google/uuid"
)

const pass = `^[a-zA-Z0-9_,"'~!@#$%^&*()\?\-\+={}<>|:;\[\]\.\s]{8,255}$`
const pass_err = `Password cannot be empty, and must be a string the length between 8 and 255 with a-zA-Z0-9_ ,"'~!@#$%^&*()?-+={}<>|:;[].`
const name = `^[А-ЯЁA-Zа-яёa-z_0-9\-]{3,255}$`
const name_err = "cannot be empty, and must be a string the length between 3 and 255 with А-ЯЁа-яёA-Za-z_0-9-"
const email = "email cannot be empty, and must be a string type email"
const job = `^[А-ЯЁа-яёa-zA-Z0-9_,"'~!@#$%^&*()\?\-\+={}<>|:;\[\]\.\s]{3,255}$`
const job_err = `Job_title must be a string the length between 3 and 255 with А-ЯЁа-яёa-zA-Z0-9_ ,"'~!@#$%^&*()?-+={}<>|:;[].`

type User struct {
	First_name  string    `json:"first_name"`
	Last_name   string    `json:"last_name"`
	Full_name   string    `json:"full_name"`
	Email       string    `json:"email"`
	Password    string    `json:"password"`
	Job_title   string    `json:"job_title"`
	Is_deleted  bool      `json:"is_deleted"`
	Create_date time.Time `json:"create_date"`
	Update_date time.Time `json:"update_date"`
	Id          uuid.UUID `json:"id"`
}

type Users []User

type UserLoginSwagger struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserRegistrationSwagger struct {
	First_name string `json:"first_name"`
	Last_name  string `json:"last_name"`
	Full_name  string `json:"full_name"`
	Email      string `json:"email"`
	Password   string `json:"password"`
	Job_title  string `json:"job_title"`
}

type UserPasswordID struct {
	Password string    `json:"password"`
	Id       uuid.UUID `json:"id"`
}

type UserPasswordSwagger struct {
	Password string `json:"password"`
}

type UserPasswordOldNewSwagger struct {
	PasswordOld string `json:"password_old"`
	PasswordNew string `json:"password_new"`
}

type UserEmail struct {
	Email string `json:"email"`
}

type UserPasswordToken struct {
	Password string `json:"password"`
	Token    string `json:"token"`
}

type UserUpdateSwagger struct {
	First_name string `json:"first_name"`
	Last_name  string `json:"last_name"`
	Full_name  string `json:"full_name"`
	Email      string `json:"email"`
	Job_title  string `json:"job_title"`
}

type UserDeleteSwagger struct {
	Id uuid.UUID `json:"id"`
}

func (u *User) ValidateCreateUser() error {
	return validation.ValidateStruct(u,
		validation.Field(&u.First_name,
			validation.Required,
			validation.Match(regexp.MustCompile(name)).
				Error("First_name"+name_err)),
		validation.Field(&u.Last_name,
			validation.Match(regexp.MustCompile(name)).
				Error("Last_name"+name_err)),
		validation.Field(&u.Full_name,
			validation.Match(regexp.MustCompile(name)).
				Error("Full_name"+name_err)),
		validation.Field(&u.Email,
			validation.Required,
			is.Email.Error(email)),
		validation.Field(&u.Password,
			validation.Required,
			validation.Match(regexp.MustCompile(pass)).
				Error(pass_err)),
		validation.Field(&u.Job_title,
			validation.Match(regexp.MustCompile(job)).
				Error(job_err)),
	)
}

func (u *User) ValidateLogin() error {
	return validation.ValidateStruct(u,
		validation.Field(&u.Password,
			validation.Required,
			validation.Match(regexp.MustCompile(pass)).Error(pass_err)),
		validation.Field(&u.Email,
			validation.Required,
			is.Email.Error(email)),
	)
}

func (u *User) ValidateUpdate() error {
	return validation.ValidateStruct(u,
		validation.Field(&u.Id,
			validation.Required,
			is.UUID),
		validation.Field(&u.First_name,
			validation.Required,
			validation.Match(regexp.MustCompile(name)).
				Error("First_name"+name_err)),
		validation.Field(&u.Last_name,
			validation.Match(regexp.MustCompile(name)).
				Error("Last_name"+name_err)),
		validation.Field(&u.Full_name,
			validation.Match(regexp.MustCompile(name)).
				Error("Full_name"+name_err)),
		validation.Field(&u.Email,
			validation.Required,
			is.Email.Error(email)),
		validation.Field(&u.Job_title,
			validation.Match(regexp.MustCompile(job)).
				Error(job_err)),
	)
}

func (u *UserPasswordSwagger) ValidatePassword() error {
	return validation.ValidateStruct(u,
		validation.Field(&u.Password,
			validation.Required,
			validation.Match(regexp.MustCompile(pass)).
				Error(pass_err)),
	)
}

func (u *UserPasswordOldNewSwagger) ValidateOldNewPassword() error {
	return validation.ValidateStruct(u,
		validation.Field(&u.PasswordOld,
			validation.Required,
			validation.Match(regexp.MustCompile(pass)).
				Error(pass_err)),
		validation.Field(&u.PasswordNew,
			validation.Required,
			validation.Match(regexp.MustCompile(pass)).
				Error(pass_err)),
	)

}

func ValidateEmail(email string) error {
	return validation.Validate(email,
		validation.Required,
		is.Email.Error(email))
}
