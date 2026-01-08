package util

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"

	"google.golang.org/api/sheets/v4"
)

func JoinSheetNames(filtered []*SheetTimeFilter) string {
	var names []string
	for _, filter := range filtered {
		sheetName := filter.Name
		if filter.TimeSpan != nil {
			sheetName = fmt.Sprintf("%s(%s-%s)",
				filter.Name,
				filter.TimeSpan.StartTime.Format("15:04"),
				filter.TimeSpan.EndTime.Format("15:04"))
		}

		names = append(names, sheetName)
	}

	return strings.Join(names, ", ")
}

func FilterSheets(sheets []*sheets.Sheet, filters []*SheetTimeFilter) (filtered []*SheetTimeFilter) {
	if len(filters) == 0 {
		for _, sheet := range sheets {
			filtered = append(filtered, &SheetTimeFilter{Name: sheet.Properties.Title})
		}
		return
	}

	for _, sheet := range sheets {
		for _, filter := range filters {
			if sheet.Properties.Title == filter.Name {
				filtered = append(filtered, filter)
				break
			}
		}
	}

	return
}

func ColumnIndexToLetter(index uint32) string {
	var bytes []byte
	for {
		bytes = append(bytes, byte('A'+(index%26)))
		if index < 26 {
			break
		}
		index = index/26 - 1
	}

	slices.Reverse(bytes)
	return string(bytes)
}

func ColumnLetterToIndex(letter string) (index uint32, err error) {
	if len(letter) == 0 {
		err = errors.New("column letter empty")
		return
	}

	for _, char := range letter {
		if char < 'A' || char > 'Z' {
			err = errors.New("invalid column letter: " + letter)
			return
		}
		index = index*26 + uint32(char-'A'+1)
	}

	index--
	return
}

type TimeSpan struct {
	StartTime time.Time
	EndTime   time.Time
}

func (ts *TimeSpan) Matches(timestamp time.Time) bool {
	hour, min, _ := timestamp.Clock()
	timeOnly := time.Date(0, 1, 1, hour, min, 0, 0, time.UTC)

	startHour, startMin, _ := ts.StartTime.Clock()
	startTimeOnly := time.Date(0, 1, 1, startHour, startMin, 0, 0, time.UTC)

	endHour, endMin, _ := ts.EndTime.Clock()
	endTimeOnly := time.Date(0, 1, 1, endHour, endMin, 0, 0, time.UTC)

	return (timeOnly.Equal(startTimeOnly) || timeOnly.After(startTimeOnly)) &&
		(timeOnly.Equal(endTimeOnly) || timeOnly.Before(endTimeOnly))
}

type SheetTimeFilter struct {
	Name     string
	TimeSpan *TimeSpan
}

func ParseSheetName(sheetSpec string) (*SheetTimeFilter, error) {
	re := regexp.MustCompile(`^(.+?)\((\d{1,2}:\d{2})-(\d{1,2}:\d{2})\)$`)
	matches := re.FindStringSubmatch(sheetSpec)

	if len(matches) == 0 {
		return &SheetTimeFilter{
			Name:     sheetSpec,
			TimeSpan: nil,
		}, nil
	}

	sheetName := matches[1]
	startTimeStr := matches[2]
	endTimeStr := matches[3]

	startTime, err := time.Parse("15:04", startTimeStr)
	if err != nil {
		return nil, fmt.Errorf("invalid start time '%s': %w", startTimeStr, err)
	}

	endTime, err := time.Parse("15:04", endTimeStr)
	if err != nil {
		return nil, fmt.Errorf("invalid end time '%s': %w", endTimeStr, err)
	}

	return &SheetTimeFilter{
		Name: sheetName,
		TimeSpan: &TimeSpan{
			StartTime: startTime,
			EndTime:   endTime,
		},
	}, nil
}

func (stf *SheetTimeFilter) Matches(timestamp time.Time) bool {
	return stf.TimeSpan == nil || stf.TimeSpan.Matches(timestamp)
}
