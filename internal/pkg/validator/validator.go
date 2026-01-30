package validator

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/yourusername/gobank/internal/pkg/apperror"
)

type Validator interface {
	Validate(i interface{}) []apperror.ValidationError
}

type customValidator struct {
	validate *validator.Validate
}

func New() Validator {
	v := validator.New()

	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return &customValidator{validate: v}
}

func (cv *customValidator) Validate(i interface{}) []apperror.ValidationError {
	var errors []apperror.ValidationError

	if err := cv.validate.Struct(i); err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			var message string
			switch err.Tag() {
			case "required":
				message = "This field is required"
			case "email":
				message = "Invalid email format"
			case "min":
				message = "Value is too short (minimum: " + err.Param() + ")"
			case "max":
				message = "Value is too long (maximum: " + err.Param() + ")"
			case "oneof":
				message = "Value must be one of: " + err.Param()
			case "nefield":
				message = "Value must be different from " + err.Param()
			case "uuid":
				message = "Invalid UUID format"
			case "gt":
				message = "Value must be greater than " + err.Param()
			case "gte":
				message = "Value must be greater than or equal to " + err.Param()
			case "lt":
				message = "Value must be less than " + err.Param()
			case "lte":
				message = "Value must be less than or equal to " + err.Param()
			default:
				message = "Validation failed for " + err.Tag()
			}

			errors = append(errors, apperror.ValidationError{
				Field:   err.Field(),
				Message: message,
			})
		}
	}

	return errors
}
