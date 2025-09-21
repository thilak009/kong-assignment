package forms

import (
	"encoding/json"
	"regexp"
	"time"

	"github.com/go-playground/validator/v10"
)

type ServiceVersionForm struct{}

type CreateServiceVersionForm struct {
	Version          string    `form:"version" json:"version" binding:"required,semver"`
	Description      string    `form:"description" json:"description" binding:"required,min=10,max=1000"`
	ReleaseTimestamp time.Time `form:"releaseTimestamp" json:"releaseTimestamp"`
}

type UpdateServiceVersionForm struct {
	Description      string     `form:"description" json:"description" binding:"omitempty,min=10,max=1000"`
	ReleaseTimestamp *time.Time `form:"releaseTimestamp" json:"releaseTimestamp,omitempty"`
}

// semverValidator validates semantic version format (e.g., 1.0.0, 2.1.3-beta)
func semverValidator(fl validator.FieldLevel) bool {
	semverRegex := `^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`
	re := regexp.MustCompile(semverRegex)
	return re.MatchString(fl.Field().String())
}

func (f ServiceVersionForm) Version(tag string, errMsg ...string) (message string) {
	switch tag {
	case "required":
		if len(errMsg) == 0 {
			return "Please enter the version"
		}
		return errMsg[0]
	case "semver":
		return "Version must be a valid semantic version (e.g., 1.0.0, 2.1.3-beta)"
	default:
		return "Something went wrong, please try again later"
	}
}

func (f ServiceVersionForm) Description(tag string, errMsg ...string) (message string) {
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

// TODO: validator for releasetimestamp

func (f ServiceVersionForm) Create(err error) string {
	switch err.(type) {
	case validator.ValidationErrors:

		if _, ok := err.(*json.UnmarshalTypeError); ok {
			return "Something went wrong, please try again later"
		}

		for _, err := range err.(validator.ValidationErrors) {
			if err.Field() == "Version" {
				return f.Version(err.Tag())
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

func (f ServiceVersionForm) Update(err error) string {
	switch err.(type) {
	case validator.ValidationErrors:

		if _, ok := err.(*json.UnmarshalTypeError); ok {
			return "Something went wrong, please try again later"
		}

		for _, err := range err.(validator.ValidationErrors) {
			if err.Field() == "Description" {
				return f.Description(err.Tag())
			}
		}

	default:
		return "Invalid request"
	}

	return "Something went wrong, please try again later"
}
