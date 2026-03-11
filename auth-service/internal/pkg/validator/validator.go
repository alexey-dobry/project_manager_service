package validator

import (
	"errors"
	"fmt"
	"strings"

	gv "github.com/go-playground/validator/v10"
)

// Validator — синглтон-обёртка над validator.v10.
type Validator struct {
	v *gv.Validate
}

func New() *Validator {
	return &Validator{v: gv.New()}
}

// Validate возвращает map "field" -> "human-readable reason" для удобного клиента.
func (vd *Validator) Validate(s any) (map[string]string, error) {
	if err := vd.v.Struct(s); err != nil {
		var ve gv.ValidationErrors
		if !errors.As(err, &ve) {
			return nil, err
		}
		out := make(map[string]string, len(ve))
		for _, fe := range ve {
			out[strings.ToLower(fe.Field())] = describe(fe)
		}
		return out, errors.New("validation failed")
	}
	return nil, nil
}

func describe(fe gv.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "is required"
	case "email":
		return "must be a valid email"
	case "min":
		return fmt.Sprintf("must be at least %s characters", fe.Param())
	case "max":
		return fmt.Sprintf("must be at most %s characters", fe.Param())
	case "oneof":
		return fmt.Sprintf("must be one of: %s", fe.Param())
	case "uuid":
		return "must be a valid UUID"
	default:
		return fmt.Sprintf("failed on rule '%s'", fe.Tag())
	}
}
