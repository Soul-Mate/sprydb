package mapper

import "time"

type Time struct {
	time.Time
	layout string
}

func NewTime(t time.Time, layout string) *Time {
	return &Time{
		t,
		layout,
	}
}