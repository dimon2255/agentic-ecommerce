package validate

import (
	"fmt"
	"net/mail"
	"regexp"

	"github.com/dimon2255/agentic-ecommerce/api/internal/apperror"
)

var uuidRegex = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// Validator accumulates field validation errors.
type Validator struct {
	errors map[string]string
}

// New creates a Validator.
func New() *Validator {
	return &Validator{errors: make(map[string]string)}
}

// Required checks that value is non-empty.
func (v *Validator) Required(field, value string) {
	if value == "" {
		v.addError(field, "required")
	}
}

// MinLength checks that value has at least min characters.
func (v *Validator) MinLength(field, value string, min int) {
	if len(value) < min {
		v.addErrorf(field, "must be at least %d characters", min)
	}
}

// MaxLength checks that value has at most max characters.
func (v *Validator) MaxLength(field, value string, max int) {
	if len(value) > max {
		v.addErrorf(field, "must be at most %d characters", max)
	}
}

// Email checks that value is a valid email address.
func (v *Validator) Email(field, value string) {
	if value == "" {
		return // use Required for empty check
	}
	if _, err := mail.ParseAddress(value); err != nil {
		v.addError(field, "invalid email address")
	}
}

// UUID checks that value is a valid UUID v4 format.
func (v *Validator) UUID(field, value string) {
	if value == "" {
		return // use Required for empty check
	}
	if !uuidRegex.MatchString(value) {
		v.addError(field, "invalid UUID format")
	}
}

// IntRange checks that value is between min and max (inclusive).
func (v *Validator) IntRange(field string, value, min, max int) {
	if value < min || value > max {
		v.addErrorf(field, "must be between %d and %d", min, max)
	}
}

// FloatMin checks that value is at least min.
func (v *Validator) FloatMin(field string, value, min float64) {
	if value < min {
		v.addErrorf(field, "must be at least %.2f", min)
	}
}

// OneOf checks that value is one of the allowed values.
func (v *Validator) OneOf(field, value string, allowed []string) {
	if value == "" {
		return
	}
	for _, a := range allowed {
		if value == a {
			return
		}
	}
	v.addErrorf(field, "must be one of: %v", allowed)
}

// Validate returns nil if no errors, or an InvalidInputError with field details.
func (v *Validator) Validate() error {
	if len(v.errors) == 0 {
		return nil
	}
	return apperror.NewInvalidInput("validation failed", v.errors)
}

// AddError records a custom validation error for a field.
func (v *Validator) AddError(field, msg string) {
	if _, exists := v.errors[field]; !exists {
		v.errors[field] = msg
	}
}

func (v *Validator) addError(field, msg string) {
	v.AddError(field, msg)
}

func (v *Validator) addErrorf(field, format string, args ...any) {
	if _, exists := v.errors[field]; !exists {
		v.errors[field] = fmt.Sprintf(format, args...)
	}
}
