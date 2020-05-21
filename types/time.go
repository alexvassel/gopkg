package types

import (
	"context"
	"github.com/severgroup-tt/gopkg-errors"
	"time"
)

const TimeLayout = "15:04"

var TimeParseError = errors.Internal.Err(context.Background(), `TimeParseError: should be a string formatted as "15:04"`)

func NewTime(t *time.Time) Time {
	return Time{t: t}
}

func NewNotEmptyTimeFromString(s string) (Time, error) {
	if s == "" {
		return Time{}, errors.Internal.Err(context.Background(), "TimeParseError: should be not empty")
	}
	t, err := NewTimeFromString(s)
	if err != nil {
		return Time{}, err
	}
	return t, nil
}

func NewTimeFromString(s string) (Time, error) {
	t := Time{}
	if s == "" {
		return t, nil
	}
	if err := t.decode(s); err != nil {
		return t, err
	}
	return t, nil
}

type Time struct {
	t *time.Time
}

func (t Time) MarshalJSON() ([]byte, error) {
	return []byte(`"` + t.t.Format(TimeLayout) + `"`), nil
}

func (t *Time) UnmarshalJSON(b []byte) error {
	s := string(b)
	return t.decode(s[1:6])
}

func (t *Time) decode(s string) error {
	if len(s) != 5 {
		return TimeParseError.WithPayloadKV("actual", s)
	}
	ret, err := time.Parse(TimeLayout, s)
	if err != nil {
		return err
	}
	t.t = &ret
	return nil
}

func (t Time) ToHHMM() string {
	if t.t == nil {
		return ""
	}
	return t.t.Format(TimeLayout)
}

func (t Time) ToMin() int16 {
	if t.t == nil {
		return 0
	}
	return int16(t.t.Hour()*60 + t.t.Minute())
}

func (t Time) ToTime() time.Time {
	if t.t == nil {
		return time.Time{}
	}
	return *t.t
}

func (t Time) After(dt time.Time) bool {
	if t.t == nil {
		return true
	}
	return dt.Hour() >= t.t.Hour() && dt.Minute() >= t.t.Minute()
}

func (t Time) Equal(dt *time.Time) bool {
	if t.t == nil {
		return dt == nil
	}
	if dt == nil {
		return t.t == nil
	}
	return t.t.Equal(*dt)
}
