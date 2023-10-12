package validator

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

// Validation contains
type Validation struct {
	validate *validator.Validate
}

type ValidationError struct {
	validator.FieldError
}

func (v *ValidationError) Error() string {
	return fmt.Sprintf("%s %s", v.FieldError.Field(), toMsg(v.FieldError))
}

func NewValidator() *Validation {
	v := validator.New(validator.WithRequiredStructEnabled())
	v.RegisterValidation("hello", validateHello)

	return &Validation{v}
}

func (v *Validation) Struct(i interface{}) []error {
	err := v.validate.Struct(i)
	if err == nil {
		return nil
	}
	ferrs := err.(validator.ValidationErrors)
	if len(ferrs) == 0 {
		return nil
	}

	errs := make([]error, 0, len(ferrs))
	for _, ferr := range ferrs {
		// cast the FieldError into our ValidationError and append to the slice
		verr := &ValidationError{ferr.(validator.FieldError)}
		errs = append(errs, verr)
	}

	return errs
}

func validateHello(fl validator.FieldLevel) bool {
	return fl.Field().String() == "hello"
}

// for getting custom error message from go-validator fe.Tag()
func toMsg(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "field required"
	case "uuid":
		return "invalid uuid"
	}

	return fmt.Sprintf("failed on tag %s", fe.Tag())
}
