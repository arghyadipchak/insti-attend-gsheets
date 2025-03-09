package main

import (
	"encoding/json"
	"errors"
	"time"

	"google.golang.org/api/sheets/v4"
)

type AttendanceMessage struct {
	UUID       string
	Attendance map[string]AttendanceRecord
}

type AttendanceRecord struct {
	Timestamp time.Time `json:"timestamp"`
}

func (a *AttendanceRecord) UnmarshalJSON(data []byte) error {
	var aux struct {
		Timestamp *time.Time `json:"timestamp"`
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if aux.Timestamp == nil {
		return errors.New("missing required field: timestamp")
	}

	a.Timestamp = *aux.Timestamp
	return nil
}

func readAttendance(data []byte) (map[string]AttendanceRecord, error) {
	var attendance map[string]AttendanceRecord
	if err := json.Unmarshal(data, &attendance); err != nil {
		return nil, err
	}

	return attendance, nil
}

func getSheetNames(sheets []*sheets.Sheet) string {
	sheetsName := ""
	for _, sheet := range sheets {
		sheetsName += sheet.Properties.Title + ", "
	}

	return sheetsName[:len(sheetsName)-2]
}

func columnIndexToLetter(index int) string {
	column := ""
	for index >= 0 {
		column = string(rune('A'+(index%26))) + column
		index = index/26 - 1
	}

	return column
}
