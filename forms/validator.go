package forms

import (
	"reflect"
	"sync"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// DefaultValidator ...
type DefaultValidator struct {
	once     sync.Once
	validate *validator.Validate
}

var _ binding.StructValidator = &DefaultValidator{}

// ValidateStruct ...
func (v *DefaultValidator) ValidateStruct(obj interface{}) error {
	if kindOfData(obj) == reflect.Struct {
		v.lazyinit()
		if err := v.validate.Struct(obj); err != nil {
			return err
		}
	}
	return nil
}

// Engine ...
func (v *DefaultValidator) Engine() interface{} {
	v.lazyinit()
	return v.validate
}

func (v *DefaultValidator) lazyinit() {
	v.once.Do(func() {
		v.validate = validator.New()
		v.validate.SetTagName("binding")

		// Register custom validators
		v.validate.RegisterValidation("semver", semverValidator)
	})
}

func kindOfData(data interface{}) reflect.Kind {
	value := reflect.ValueOf(data)
	valueType := value.Kind()
	if valueType == reflect.Ptr {
		valueType = value.Elem().Kind()
	}
	return valueType
}
