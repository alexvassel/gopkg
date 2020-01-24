package types

import "time"

const dateTimeLayout = "2006-01-02 15:04:05.999999999 -0700 MST"
const YMDThmsLayout = "20060102T150405"

func DateTimeToDBString(date time.Time) string {
	return date.Format(dateTimeLayout)
}

func DateTimeToYMDTHms(date time.Time) string {
	return date.Format(YMDThmsLayout)
}
