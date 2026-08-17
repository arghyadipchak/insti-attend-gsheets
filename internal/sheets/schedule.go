package sheets

import (
	"slices"
	"time"
)

var weekdayByName = map[string]time.Weekday{
	"Sunday":    time.Sunday,
	"Monday":    time.Monday,
	"Tuesday":   time.Tuesday,
	"Wednesday": time.Wednesday,
	"Thursday":  time.Thursday,
	"Friday":    time.Friday,
	"Saturday":  time.Saturday,
}

func weekdaysFromNames(names []string) []time.Weekday {
	weekdays := make([]time.Weekday, 0, len(names))
	for _, name := range names {
		if weekday, ok := weekdayByName[name]; ok {
			weekdays = append(weekdays, weekday)
		}
	}
	return weekdays
}

func timeOfDay(t time.Time) time.Time {
	return time.Date(0, time.January, 1, t.Hour(), t.Minute(), t.Second(), 0, t.Location())
}

type schedule struct {
	allowedWeekdays []time.Weekday
	activeFrom      time.Time
	activeTo        time.Time
	timeOfDayOnly   bool
}

func newSchedule(
	allowedWeekdays []string,
	activeFrom, activeTo time.Time,
	timeOfDayOnly bool,
) schedule {
	if timeOfDayOnly {
		if !activeFrom.IsZero() {
			activeFrom = timeOfDay(activeFrom)
		}
		if !activeTo.IsZero() {
			activeTo = timeOfDay(activeTo)
		}
	}

	return schedule{
		allowedWeekdays: weekdaysFromNames(allowedWeekdays),
		activeFrom:      activeFrom,
		activeTo:        activeTo,
		timeOfDayOnly:   timeOfDayOnly,
	}
}

func (s *schedule) matches(timestamp time.Time) bool {
	if len(s.allowedWeekdays) > 0 && !slices.Contains(s.allowedWeekdays, timestamp.Weekday()) {
		return false
	}

	if s.activeFrom.IsZero() && s.activeTo.IsZero() {
		return true
	}

	if s.timeOfDayOnly {
		timestamp = timeOfDay(timestamp)
	}

	if !s.activeFrom.IsZero() && timestamp.Before(s.activeFrom) {
		return false
	}
	if !s.activeTo.IsZero() && timestamp.After(s.activeTo) {
		return false
	}

	return true
}
