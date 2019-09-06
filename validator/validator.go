package validator

import (
	"sync"

	"gopkg.in/go-playground/validator.v9"
)

var (
	initOnce sync.Once
	validate *validator.Validate
)

// New create new validator instance with custom tag
func New() *validator.Validate {
	initOnce.Do(func() {
		validate = validator.New()
		customValidations := map[string]validator.Func{
			"date_ymd":       ValidateDateYMD,
			"date_rfc3339":   ValidateDateRfc3339,
			"str_gte":        ValidateStrGTE,
			"str_lte":        ValidateStrLTE,
			"field_empty":    ValidateFieldEmpty,
			"field_required": ValidateFieldRequired,
			"time_hhmm":      ValidateTimeHHMM,
			"time_hhmmss":    ValidateTimeHHMMSS,
			"min_letter":     ValidateMinLetter,
		}
		for tag, fn := range customValidations {
			_ = validate.RegisterValidation(tag, fn)
		}
	})

	return validate
}
