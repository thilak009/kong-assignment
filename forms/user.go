package forms

import (
	"encoding/json"

	"github.com/go-playground/validator/v10"
)

type UserForm struct{}

type CreateUserForm struct {
	Email    string `form:"email" json:"email" binding:"required,email,min=3,max=100"`
	Name     string `form:"name" json:"name" binding:"required,min=2,max=100"`
	Password string `form:"password" json:"password" binding:"required,min=8,max=100,strongpassword"`
}

type UpdateUserForm struct {
	Password string `form:"password" json:"password" binding:"required,min=8,max=100,strongpassword"`
}

type LoginForm struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (f UserForm) Email(tag string, errMsg ...string) (message string) {
	switch tag {
	case "required":
		if len(errMsg) == 0 {
			return "Please enter the email"
		}
		return errMsg[0]
	case "min", "max":
		return "Email should be between 3 to 100 characters"
	default:
		return "Something went wrong, please try again later"
	}
}

func (f UserForm) Password(tag string, errMsg ...string) (message string) {
	switch tag {
	case "required":
		if len(errMsg) == 0 {
			return "Please enter the password"
		}
		return errMsg[0]
	case "min", "max":
		return "Password should be between 8 to 100 characters"
	case "strongpassword":
		return "Password must contain at least one uppercase letter, one lowercase letter, and one special character"
	default:
		return "Something went wrong, please try again later"
	}
}

func (f UserForm) Create(err error) string {
	switch err.(type) {
	case validator.ValidationErrors:

		if _, ok := err.(*json.UnmarshalTypeError); ok {
			return "Something went wrong, please try again later"
		}

		for _, err := range err.(validator.ValidationErrors) {
			if err.Field() == "Email" {
				return f.Email(err.Tag())
			}
			if err.Field() == "Password" {
				return f.Password(err.Tag())
			}
		}

	default:
		return "Invalid request"
	}

	return "Something went wrong, please try again later"
}
