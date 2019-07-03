package errors

import (
	"context"
	"reflect"
	"regexp"
	"strings"

	"github.com/severgroup-tt/gopkg-errors"

	"gopkg.in/go-playground/validator.v9"
)

const ValidationMsg = "Invalid request payload"

// Global validation codes
const (
	CommonTagCode    = "INVALID"
	RequiredTagCode  = "EMPTY"
	LenTagCode       = "INVALID_LENGTH"
	SizeTagCode      = "OUT_OF_RANGE"
	TooFewItemsCode  = "TOO_FEW_ELEMENTS"
	TooManyItemsCode = "TOO_MANY_ELEMENTS"
)

// Converter converts error from validation errors
func Converter() errors.ErrorConverter {
	return func(ctx context.Context, err error) (*errors.Error, bool) {
		for {
			if errValidation, ok := err.(validator.ValidationErrors); ok {
				return convertValidationError(ctx, errValidation), true
			}
			errC, ok := err.(errors.Causer)
			if !ok {
				return nil, false
			}
			err = errC.Cause()
		}
	}
}

func convertValidationError(ctx context.Context, err validator.ValidationErrors) *errors.Error {
	result := errors.BadRequest.ErrWrap(ctx, ValidationMsg, err)
	for _, fieldErr := range err {
		status := CommonTagCode
		switch fieldErr.Tag() {
		case "required":
			status = RequiredTagCode
		case "len":
			status = LenTagCode
		case "min":
			status = SizeTagCode
			if fieldErr.Kind() == reflect.Slice {
				status = TooFewItemsCode
			}
		case "max":
			status = SizeTagCode
			if fieldErr.Kind() == reflect.Slice {
				status = TooManyItemsCode
			}
		case "eq", "ne", "lt", "lte", "gt", "gte":
			status = SizeTagCode
		}

		field := toSnakeCase(fieldErr.Field())
		result = result.WithPayloadKV(field, status)
	}

	return result
}

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

func toSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}
