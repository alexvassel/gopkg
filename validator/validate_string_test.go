package validator

import (
	"github.com/stretchr/testify/assert"
	"gopkg.in/go-playground/validator.v9"
	"testing"
)

func Test_validateMinLetter(t *testing.T) {
	v := validator.New()
	err := v.RegisterValidation("min_letter", ValidateMinLetter)
	if err != nil {
		t.Fatal(err)
	}

	type field struct {
		Val string `validate:"min_letter=2"`
	}

	cs := map[string]bool{
		"":    false,
		"1":   false,
		"s":   false,
		"12":  true,
		"dc":  true,
		"123": true,
		"dfj": true,
	}

	for value, ok := range cs {
		err := v.Struct(field{Val: value})
		if ok {
			assert.Nil(t, err)
		} else {
			assert.NotNil(t, err)
		}
	}
}
