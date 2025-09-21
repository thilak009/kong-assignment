package forms

import (
	"encoding/json"

	"github.com/go-playground/validator/v10"
)

type ServiceForm struct{}

type CreateServiceForm struct {
	Name        string `form:"name" json:"name" binding:"required,min=3,max=100"`
	Description string `form:"description" json:"description" binding:"required,min=10,max=1000"`
}

func (f ServiceForm) Name(tag string, errMsg ...string) (message string) {
	switch tag {
	case "required":
		if len(errMsg) == 0 {
			return "Please enter the name"
		}
		return errMsg[0]
	case "min", "max":
		return "Name should be between 3 to 100 characters"
	default:
		return "Something went wrong, please try again later"
	}
}

func (f ServiceForm) Description(tag string, errMsg ...string) (message string) {
	switch tag {
	case "required":
		if len(errMsg) == 0 {
			return "Please enter the description"
		}
		return errMsg[0]
	case "min", "max":
		return "Description should be between 10 to 1000 characters"
	default:
		return "Something went wrong, please try again later"
	}
}

func (f ServiceForm) Create(err error) string {
	switch err.(type) {
	case validator.ValidationErrors:

		if _, ok := err.(*json.UnmarshalTypeError); ok {
			return "Something went wrong, please try again later"
		}

		for _, err := range err.(validator.ValidationErrors) {
			if err.Field() == "Name" {
				return f.Name(err.Tag())
			}
			if err.Field() == "Description" {
				return f.Description(err.Tag())
			}
		}

	default:
		return "Invalid request"
	}

	return "Something went wrong, please try again later"
}
