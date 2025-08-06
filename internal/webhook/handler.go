package webhook

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/arghyadipchak/insti-attend-gsheets/internal/config"
	"github.com/arghyadipchak/insti-attend-gsheets/internal/msg"
	"github.com/google/uuid"
)

type Response struct {
	UUID string `json:"uuid"`
}

func handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")

	if r.Method == http.MethodOptions {
		logO.Println("preflight request")
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		logE.Println("invalid request method:", r.Method)
		http.Error(w, "invalid request method", http.StatusMethodNotAllowed)
		return
	}

	if config.AuthToken != "" && r.Header.Get("Authorization") != "Bearer "+config.AuthToken {
		logE.Print("unauthorized request")
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		logE.Println("failed to read request body:", err)
		http.Error(w, "failed to read request body", http.StatusInternalServerError)
		return
	}

	attendance, err := msg.ReadAttendance(body)
	if err != nil {
		logE.Println("failed to parse JSON:", err)
		http.Error(w, "failed to parse JSON", http.StatusBadRequest)
		return
	}

	if len(attendance) > 0 {
		recordId := uuid.New().String()

		logO.Printf("received %d attendance records: %s", len(attendance), recordId)
		for rollNo, record := range attendance {
			logO.Printf("  %s: %s", rollNo, record.Timestamp)
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(Response{UUID: recordId}); err != nil {
			logE.Println("failed to write response:", err)
			http.Error(w, "failed to write response", http.StatusInternalServerError)
			return
		}

		msg.AttendanceQueue <- msg.NewAttendanceMessage(recordId, attendance)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}
