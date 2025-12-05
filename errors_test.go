package inrequest

import (
	"errors"
	"testing"
)

func TestParseError(t *testing.T) {
	t.Run("ParseError with underlying error", func(t *testing.T) {
		underlying := errors.New("underlying error")
		err := NewParseError("form", "failed to parse", underlying)

		if err.Type != "form" {
			t.Errorf("expected type 'form', got '%s'", err.Type)
		}
		if err.Message != "failed to parse" {
			t.Errorf("expected message 'failed to parse', got '%s'", err.Message)
		}
		if err.Err != underlying {
			t.Error("underlying error not set correctly")
		}

		expectedMsg := "inrequest: form parse error: failed to parse: underlying error"
		if err.Error() != expectedMsg {
			t.Errorf("expected '%s', got '%s'", expectedMsg, err.Error())
		}

		if err.Unwrap() != underlying {
			t.Error("Unwrap did not return underlying error")
		}
	})

	t.Run("ParseError without underlying error", func(t *testing.T) {
		err := NewParseError("json", "invalid JSON", nil)

		expectedMsg := "inrequest: json parse error: invalid JSON"
		if err.Error() != expectedMsg {
			t.Errorf("expected '%s', got '%s'", expectedMsg, err.Error())
		}

		if err.Unwrap() != nil {
			t.Error("Unwrap should return nil")
		}
	})

	t.Run("IsParseError", func(t *testing.T) {
		parseErr := NewParseError("form", "test", nil)
		bindErr := NewBindError("field", "test", nil)
		regularErr := errors.New("regular error")

		if !IsParseError(parseErr) {
			t.Error("IsParseError should return true for ParseError")
		}
		if IsParseError(bindErr) {
			t.Error("IsParseError should return false for BindError")
		}
		if IsParseError(regularErr) {
			t.Error("IsParseError should return false for regular error")
		}
	})
}

func TestBindError(t *testing.T) {
	t.Run("BindError with field and underlying error", func(t *testing.T) {
		underlying := errors.New("underlying error")
		err := NewBindError("username", "invalid type", underlying)

		if err.Field != "username" {
			t.Errorf("expected field 'username', got '%s'", err.Field)
		}
		if err.Message != "invalid type" {
			t.Errorf("expected message 'invalid type', got '%s'", err.Message)
		}
		if err.Err != underlying {
			t.Error("underlying error not set correctly")
		}

		expectedMsg := "inrequest: bind error on field 'username': invalid type: underlying error"
		if err.Error() != expectedMsg {
			t.Errorf("expected '%s', got '%s'", expectedMsg, err.Error())
		}

		if err.Unwrap() != underlying {
			t.Error("Unwrap did not return underlying error")
		}
	})

	t.Run("BindError with field but no underlying error", func(t *testing.T) {
		err := NewBindError("email", "required", nil)

		expectedMsg := "inrequest: bind error on field 'email': required"
		if err.Error() != expectedMsg {
			t.Errorf("expected '%s', got '%s'", expectedMsg, err.Error())
		}
	})

	t.Run("BindError without field but with underlying error", func(t *testing.T) {
		underlying := errors.New("json error")
		err := NewBindError("", "failed to unmarshal", underlying)

		expectedMsg := "inrequest: bind error: failed to unmarshal: json error"
		if err.Error() != expectedMsg {
			t.Errorf("expected '%s', got '%s'", expectedMsg, err.Error())
		}
	})

	t.Run("BindError without field and without underlying error", func(t *testing.T) {
		err := NewBindError("", "model must be a pointer", nil)

		expectedMsg := "inrequest: bind error: model must be a pointer"
		if err.Error() != expectedMsg {
			t.Errorf("expected '%s', got '%s'", expectedMsg, err.Error())
		}
	})

	t.Run("IsBindError", func(t *testing.T) {
		parseErr := NewParseError("form", "test", nil)
		bindErr := NewBindError("field", "test", nil)
		regularErr := errors.New("regular error")

		if !IsBindError(bindErr) {
			t.Error("IsBindError should return true for BindError")
		}
		if IsBindError(parseErr) {
			t.Error("IsBindError should return false for ParseError")
		}
		if IsBindError(regularErr) {
			t.Error("IsBindError should return false for regular error")
		}
	})
}

func TestErrorUnwrap(t *testing.T) {
	t.Run("errors.Is works with ParseError", func(t *testing.T) {
		underlying := errors.New("underlying")
		parseErr := NewParseError("form", "test", underlying)

		if !errors.Is(parseErr, underlying) {
			t.Error("errors.Is should match underlying error")
		}
	})

	t.Run("errors.Is works with BindError", func(t *testing.T) {
		underlying := errors.New("underlying")
		bindErr := NewBindError("field", "test", underlying)

		if !errors.Is(bindErr, underlying) {
			t.Error("errors.Is should match underlying error")
		}
	})
}
