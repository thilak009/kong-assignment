package forms

import (
	"encoding/json"

	"github.com/go-playground/validator/v10"
)

type OrganizationForm struct{}

type CreateOrganizationForm struct {
	Name        string `json:"name" binding:"required,min=3,max=100"`
	Description string `json:"description" binding:"required,min=10,max=1000"`
}

func (f OrganizationForm) Name(tag string, errMsg ...string) (message string) {
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

func (f OrganizationForm) Description(tag string, errMsg ...string) (message string) {
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

func (f OrganizationForm) Create(err error) string {
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
