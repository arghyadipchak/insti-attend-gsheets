package attendances

import (
	"encoding/json"
	"errors"
	"io"
	"time"

	u "github.com/arghyadipchak/insti-attend-gsheets/internal/utils"
)

type AttendanceRecord struct {
	Timestamp time.Time `json:"timestamp"`
	Synced    bool      `json:"synced"`
}

func (a *AttendanceRecord) UnmarshalJSON(data []byte) error {
	var aux struct {
		Timestamp *time.Time `json:"timestamp"`
		Synced    bool       `json:"synced"`
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if aux.Timestamp == nil {
		return errors.New("missing timestamp field")
	}

	a.Timestamp = *aux.Timestamp
	a.Synced = aux.Synced

	return nil
}

type Attendance map[string]*AttendanceRecord

func ReadAttendanceFromIO(body io.ReadCloser) (Attendance, error) {
	var attendance Attendance
	if err := json.NewDecoder(body).Decode(&attendance); err != nil {
		return nil, err
	}
	return attendance, nil
}

func (a *Attendance) UnmarshalJSON(data []byte) error {
	var raw map[string]*AttendanceRecord
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	normalized := make(Attendance, len(raw))
	for rollNumber, record := range raw {
		normalized[u.NormalizeRollNumber(rollNumber)] = record
	}

	*a = normalized
	return nil
}

func (a *Attendance) CountUnsynced() int {
	count := 0
	for _, record := range *a {
		if !record.Synced {
			count++
		}
	}
	return count
}
