package internal

import (
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
)

var Validate *validator.Validate

func InitValidator() {
	Validate = validator.New()
}

func ValidateStruct(s interface{}) error {
	err := Validate.Struct(s)
	if err != nil {
		if _, ok := err.(*validator.InvalidValidationError); ok {
			return err
		}

		var errStr string
		for _, err := range err.(validator.ValidationErrors) {
			errStr += fmt.Sprintf("Field '%s' failed validation on '%s'\n", err.Field(), err.Tag())
		}
		return errors.New(errStr)
	}
	return nil
}
