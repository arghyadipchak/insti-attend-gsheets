package msg

import (
	"encoding/json"
	"errors"
	"time"
)

var (
	AttendanceQueue = make(chan AttendanceMessage, 10)
	SheetStopped    = make(chan struct{})
	WebhookStop     = make(chan struct{})
	WebhookStopped  = make(chan struct{})
)

type AttendanceMessage struct {
	UUID       string
	Attendance map[string]AttendanceRecord
}

func NewAttendanceMessage(uuid string, attendance map[string]AttendanceRecord) AttendanceMessage {
	return AttendanceMessage{
		UUID:       uuid,
		Attendance: attendance,
	}
}

type AttendanceRecord struct {
	Timestamp time.Time `json:"timestamp"`
}

func (a *AttendanceRecord) UnmarshalJSON(data []byte) (err error) {
	var aux struct {
		Timestamp *time.Time `json:"timestamp"`
	}
	if err = json.Unmarshal(data, &aux); err == nil {
		if aux.Timestamp == nil {
			err = errors.New("missing required field: timestamp")
		} else {
			a.Timestamp = *aux.Timestamp
		}
	}

	return
}

func ReadAttendance(data []byte) (attendance map[string]AttendanceRecord, err error) {
	err = json.Unmarshal(data, &attendance)
	return
}
