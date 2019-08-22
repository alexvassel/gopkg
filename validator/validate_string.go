package validator

import (
	"gopkg.in/go-playground/validator.v9"
	"strconv"
	"strings"
)

// ValidateMinLetter ...
func ValidateMinLetter(fl validator.FieldLevel) bool {
	val := strings.TrimSpace(fl.Field().String())

	min, err := strconv.Atoi(fl.Param())
	if err != nil {
		return false
	}

	return len(val) >= min
}
