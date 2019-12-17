package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/go-playground/validator.v10"
)

func Test_validateFieldEmpty(t *testing.T) {
	v := validator.New()
	err := v.RegisterValidation("field_empty", ValidateFieldEmpty)
	if err != nil {
		t.Fatal(err)
	}
	type check struct {
		First  string
		Second string `validate:"field_empty=First|oneof=val1 val2"`
	}

	t.Run("FirstEmpty_SecondEmpty", func(t *testing.T) {
		s := check{
			First:  "",
			Second: "",
		}

		err := v.Struct(s)
		assert.Nil(t, err)
	})

	t.Run("FirstEmpty_SecondAny", func(t *testing.T) {
		s := check{
			First:  "",
			Second: "xyz",
		}

		err := v.Struct(s)
		assert.Nil(t, err)
	})

	t.Run("FirstEmpty_SecondValid", func(t *testing.T) {
		s := check{
			First:  "",
			Second: "val2",
		}

		err := v.Struct(s)
		assert.Nil(t, err)
	})

	t.Run("FirstFilled_SecondEmpty", func(t *testing.T) {
		s := check{
			First:  "val1",
			Second: "",
		}

		err := v.Struct(s)
		assert.NotNil(t, err)
	})

	t.Run("FirstFilled_SecondFilled", func(t *testing.T) {
		s := check{
			First:  "val1",
			Second: "xyz",
		}

		err := v.Struct(s)
		assert.NotNil(t, err)
	})

	t.Run("FirstFilled_SecondValid", func(t *testing.T) {
		s := check{
			First:  "val1",
			Second: "val2",
		}

		err := v.Struct(s)
		assert.Nil(t, err)
	})
}
