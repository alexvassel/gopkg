package validator

import (
	"github.com/go-playground/validator/v10"
	"strconv"
	"strings"
	"unicode/utf8"
)

// ValidateMinLetter ...
func ValidateMinLetter(fl validator.FieldLevel) bool {
	val := strings.TrimSpace(fl.Field().String())

	min, err := strconv.Atoi(fl.Param())
	if err != nil {
		return false
	}

	return utf8.RuneCountInString(val) >= min
}
