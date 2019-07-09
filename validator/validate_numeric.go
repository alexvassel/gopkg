package validator

import (
	"strconv"

	"gopkg.in/go-playground/validator.v9"
)

// ValidateStrGTE ...
func ValidateStrGTE(fl validator.FieldLevel) bool {
	val := fl.Field().String()
	if len(val) == 0 {
		return true
	}

	lval, err := strconv.ParseFloat(val, 32)
	if err != nil {
		return false
	}

	rval, err := strconv.ParseFloat(fl.Param(), 32)
	if err != nil {
		return false
	}

	return lval >= rval
}

// ValidateStrLTE ...
func ValidateStrLTE(fl validator.FieldLevel) bool {
	val := fl.Field().String()
	if len(val) == 0 {
		return true
	}

	lval, err := strconv.ParseFloat(val, 32)
	if err != nil {
		return false
	}

	rval, err := strconv.ParseFloat(fl.Param(), 32)
	if err != nil {
		return false
	}

	return lval <= rval
}