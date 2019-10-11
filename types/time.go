package types

import (
	"context"
	"github.com/severgroup-tt/gopkg-errors"
	"time"
)

const timeLayout = "15:04"

var TimeParseError = errors.Internal.Err(context.Background(), `TimeParseError: should be a string formatted as "15:04"`)

func NewTime(val string) (*Time, error) {
	ret := &Time{}
	err := ret.UnmarshalJSON([]byte(`"` + val + `"`))
	return ret, err
}

type Time struct {
	time.Time
}

func (t Time) MarshalJSON() ([]byte, error) {
	return []byte(`"` + t.Format(timeLayout) + `"`), nil
}

func (t *Time) UnmarshalJSON(b []byte) error {
	s := string(b)
	if len(s) != 7 {
		return TimeParseError.WithPayloadKV("actual", s)
	}
	ret, err := time.Parse(timeLayout, s[1:6])
	if err != nil {
		return err
	}
	t.Time = ret
	return nil
}

func TimeToString(t *Time) string {
	if t == nil {
		return ""
	}
	return t.Format(timeLayout)
}

func (t *Time) After(dt time.Time) bool {
	return dt.Hour() >= t.Hour() && dt.Minute() >= t.Minute()
}
