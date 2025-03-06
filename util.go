package main

import (
	"encoding/json"
	"time"
)

type AttendanceRecord struct {
	Timestamp time.Time `json:"timestamp"`
}

func readAttendance(data []byte) (map[string]AttendanceRecord, error) {
	var attendance map[string]AttendanceRecord
	if err := json.Unmarshal(data, &attendance); err != nil {
		return nil, err
	}

	return attendance, nil
}

func columnIndexToLetter(index int) string {
	column := ""
	for index >= 0 {
		column = string(rune('A'+(index%26))) + column
		index = index/26 - 1
	}

	return column
}
