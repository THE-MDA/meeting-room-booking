package domain

import (
	"errors"
	"time"
)

type TimeRange struct {
	Start time.Time
	End   time.Time
}

func NewTimeRange(start, end time.Time) (*TimeRange, error) {
	if start.IsZero() || end.IsZero() {
		return nil, errors.New("start and end times are required")
	}

	if !end.After(start) {
		return nil, errors.New("end time must be after start time")
	}

	return &TimeRange{
		Start: start.UTC(),
		End:   end.UTC(),
	}, nil
}

func (tr *TimeRange) Contains(t time.Time) bool {
	return (t.Equal(tr.Start) || t.After(tr.Start)) && t.Before(tr.End)
}

func (tr *TimeRange) Overlaps(other *TimeRange) bool {
	return tr.Start.Before(other.End) && tr.End.After(other.Start)
}
