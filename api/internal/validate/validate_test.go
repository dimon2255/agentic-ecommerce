package validate

import (
	"testing"

	"github.com/dimon2255/agentic-ecommerce/api/internal/apperror"
)

func TestRequired_Empty(t *testing.T) {
	v := New()
	v.Required("name", "")
	err := v.Validate()
	assertFieldError(t, err, "name", "required")
}

func TestRequired_NonEmpty(t *testing.T) {
	v := New()
	v.Required("name", "test")
	if err := v.Validate(); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestMinLength(t *testing.T) {
	v := New()
	v.MinLength("slug", "ab", 3)
	err := v.Validate()
	assertFieldError(t, err, "slug", "must be at least 3 characters")
}

func TestMinLength_Passes(t *testing.T) {
	v := New()
	v.MinLength("slug", "abc", 3)
	if err := v.Validate(); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestMaxLength(t *testing.T) {
	v := New()
	v.MaxLength("name", "abcdef", 5)
	err := v.Validate()
	assertFieldError(t, err, "name", "must be at most 5 characters")
}

func TestEmail_Valid(t *testing.T) {
	v := New()
	v.Email("email", "user@example.com")
	if err := v.Validate(); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestEmail_Invalid(t *testing.T) {
	v := New()
	v.Email("email", "not-an-email")
	err := v.Validate()
	assertFieldError(t, err, "email", "invalid email address")
}

func TestEmail_Empty_Skipped(t *testing.T) {
	v := New()
	v.Email("email", "")
	if err := v.Validate(); err != nil {
		t.Errorf("expected no error for empty email, got %v", err)
	}
}

func TestUUID_Valid(t *testing.T) {
	v := New()
	v.UUID("id", "550e8400-e29b-41d4-a716-446655440000")
	if err := v.Validate(); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestUUID_Invalid(t *testing.T) {
	v := New()
	v.UUID("id", "not-a-uuid")
	err := v.Validate()
	assertFieldError(t, err, "id", "invalid UUID format")
}

func TestUUID_Empty_Skipped(t *testing.T) {
	v := New()
	v.UUID("id", "")
	if err := v.Validate(); err != nil {
		t.Errorf("expected no error for empty UUID, got %v", err)
	}
}

func TestIntRange_Below(t *testing.T) {
	v := New()
	v.IntRange("quantity", 0, 1, 999)
	err := v.Validate()
	assertFieldError(t, err, "quantity", "must be between 1 and 999")
}

func TestIntRange_Above(t *testing.T) {
	v := New()
	v.IntRange("quantity", 1000, 1, 999)
	err := v.Validate()
	assertFieldError(t, err, "quantity", "must be between 1 and 999")
}

func TestIntRange_Valid(t *testing.T) {
	v := New()
	v.IntRange("quantity", 5, 1, 999)
	if err := v.Validate(); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestFloatMin(t *testing.T) {
	v := New()
	v.FloatMin("price", -1.0, 0)
	err := v.Validate()
	assertFieldError(t, err, "price", "must be at least 0.00")
}

func TestFloatMin_Valid(t *testing.T) {
	v := New()
	v.FloatMin("price", 9.99, 0)
	if err := v.Validate(); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestOneOf_Valid(t *testing.T) {
	v := New()
	v.OneOf("status", "active", []string{"draft", "active", "archived"})
	if err := v.Validate(); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestOneOf_Invalid(t *testing.T) {
	v := New()
	v.OneOf("status", "deleted", []string{"draft", "active", "archived"})
	err := v.Validate()
	if err == nil {
		t.Fatal("expected error")
	}
	ie, ok := err.(*apperror.InvalidInputError)
	if !ok {
		t.Fatalf("expected InvalidInputError, got %T", err)
	}
	if _, exists := ie.Fields["status"]; !exists {
		t.Error("expected status field error")
	}
}

func TestOneOf_Empty_Skipped(t *testing.T) {
	v := New()
	v.OneOf("status", "", []string{"draft", "active"})
	if err := v.Validate(); err != nil {
		t.Errorf("expected no error for empty value, got %v", err)
	}
}

func TestMultipleErrors(t *testing.T) {
	v := New()
	v.Required("name", "")
	v.Required("slug", "")
	v.FloatMin("price", -5, 0)
	err := v.Validate()
	if err == nil {
		t.Fatal("expected error")
	}
	ie, ok := err.(*apperror.InvalidInputError)
	if !ok {
		t.Fatalf("expected InvalidInputError, got %T", err)
	}
	if len(ie.Fields) != 3 {
		t.Errorf("expected 3 field errors, got %d", len(ie.Fields))
	}
}

func TestFirstErrorWins(t *testing.T) {
	v := New()
	v.Required("name", "")
	v.MinLength("name", "", 3)
	err := v.Validate()
	ie := err.(*apperror.InvalidInputError)
	// First error for "name" should be "required", not "must be at least..."
	if ie.Fields["name"] != "required" {
		t.Errorf("expected 'required', got %s", ie.Fields["name"])
	}
}

func TestNoErrors(t *testing.T) {
	v := New()
	v.Required("name", "test")
	v.MinLength("name", "test", 2)
	if err := v.Validate(); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

// helper
func assertFieldError(t *testing.T, err error, field, expectedMsg string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error for field %s", field)
	}
	ie, ok := err.(*apperror.InvalidInputError)
	if !ok {
		t.Fatalf("expected InvalidInputError, got %T", err)
	}
	msg, exists := ie.Fields[field]
	if !exists {
		t.Fatalf("expected field error for %s", field)
	}
	if msg != expectedMsg {
		t.Errorf("field %s: expected %q, got %q", field, expectedMsg, msg)
	}
}
