package types

import (
	"context"
	"time"

	"github.com/severgroup-tt/gopkg-errors"
)

const DateFormatYMD = "2006-01-02"
const week = time.Hour * 24 * 7

var DateParseError = errors.Internal.Err(context.Background(), `TimeParseError: should be a string formatted as "2006-01-02"`)

func NewDate(val string) (*Date, error) {
	ret := &Date{}
	err := ret.UnmarshalJSON([]byte(`"` + val + `"`))
	return ret, err
}

type Date struct {
	time.Time
}

func (d Date) ToString() string {
	return d.Format(DateFormatYMD)
}

func (d Date) MarshalJSON() ([]byte, error) {
	return []byte(`"` + d.Format(DateFormatYMD) + `"`), nil
}

func (d *Date) UnmarshalJSON(b []byte) error {
	s := string(b)
	if len(s) != 12 {
		return DateParseError.WithPayloadKV("actual", s)
	}
	ret, err := time.Parse(DateFormatYMD, s[1:11])
	if err != nil {
		return err
	}
	d.Time = ret
	return nil
}

func StringToDate(str string) (*time.Time, error) {
	if str == "" {
		return nil, nil
	}

	t, err := time.Parse(DateFormatYMD, str)

	return &t, err
}

func DateToString(date *time.Time) string {
	if date == nil {
		return ""
	}
	return date.Format(DateFormatYMD)
}
