package validator

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

type ValidationError struct {
	s string
}

func (v *ValidationError) Error() string {
	return v.s
}

// Validation contains
type Validation struct {
	validate *validator.Validate
}

func NewValidator() *Validation {
	v := validator.New(validator.WithRequiredStructEnabled())
	v.RegisterValidation("hello", validateHello)

	return &Validation{v}
}

func (v *Validation) Struct(i interface{}) error {
	err := v.validate.Struct(i)
	if err == nil {
		return nil
	}
	errs := err.(validator.ValidationErrors)

	if len(errs) == 0 {
		return nil
	}

	var b []byte
	b = append(b, fmt.Sprintf("%q %s", errs[0].Field(), toMsg(errs[0]))...)
	for _, fe := range errs[1:] {
		b = append(b, fmt.Sprintf(", %q %s", fe.Field(), toMsg(fe))...)
	}

	return &ValidationError{s: string(b)}
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
