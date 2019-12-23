package validator

import (
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
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
		"я":   false,
		"d":   false,
		"12":  true,
		"яв":  true,
		"gh":  true,
		"123": true,
		"ява": true,
		"jfg": true,
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
