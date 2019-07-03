package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gopkg.in/go-playground/validator.v9"
)

func Test_validateSGTE(t *testing.T) {
	v := validator.New()
	err := v.RegisterValidation("str_gte", ValidateStrGTE)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Greater", func(t *testing.T) {
		s := struct {
			Val string `validate:"str_gte=1.2345"`
		}{
			Val: "1.3",
		}

		err := v.Struct(s)
		assert.Nil(t, err)
	})

	t.Run("Equal", func(t *testing.T) {
		s := struct {
			Val string `validate:"str_gte=1.2345"`
		}{
			Val: "1.2345",
		}

		err := v.Struct(s)
		assert.Nil(t, err)
	})

	t.Run("Less", func(t *testing.T) {
		s := struct {
			Val string `validate:"str_gte=1.2345"`
		}{
			Val: "1.2",
		}

		err := v.Struct(s)
		assert.NotNil(t, err)
	})
}

func Test_validateStrLTE(t *testing.T) {
	v := validator.New()
	err := v.RegisterValidation("str_lte", ValidateStrLTE)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Greater", func(t *testing.T) {
		s := struct {
			Val string `validate:"str_lte=1.2345"`
		}{
			Val: "1.3",
		}

		err := v.Struct(s)
		assert.NotNil(t, err)
	})

	t.Run("Equal", func(t *testing.T) {
		s := struct {
			Val string `validate:"str_lte=1.2345"`
		}{
			Val: "1.2345",
		}

		err := v.Struct(s)
		assert.Nil(t, err)
	})

	t.Run("Less", func(t *testing.T) {
		s := struct {
			Val string `validate:"str_lte=1.2345"`
		}{
			Val: "1.2",
		}

		err := v.Struct(s)
		assert.Nil(t, err)
	})
}
