package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/google/uuid"
)

type Response struct {
	UUID string `json:"uuid"`
}

func webhookHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")

	if r.Method == http.MethodOptions {
		log.Println("[webhook] preflight request")
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		log.Println("[webhook] invalid request method:", r.Method)
		http.Error(w, "invalid request method", http.StatusMethodNotAllowed)
		return
	}

	if authToken != "" && r.Header.Get("Authorization") != "Bearer "+authToken {
		log.Print("[webhook] unauthorized request")
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("[webhook] failed to read request body:", err)
		http.Error(w, "failed to read request body", http.StatusInternalServerError)
		return
	}

	attendance, err := readAttendance(body)
	if err != nil {
		log.Println("[webhook] failed to parse JSON:", err)
		http.Error(w, "failed to parse JSON", http.StatusBadRequest)
		return
	}

	if len(attendance) > 0 {
		recordId := uuid.New().String()

		log.Printf("[webhook] received %d attendance records: %s", len(attendance), recordId)
		for rollNo, record := range attendance {
			log.Printf("[webhook]   %s: %s", rollNo, record.Timestamp)
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(Response{UUID: recordId}); err != nil {
			log.Println("[webhook] failed to write response:", err)
			http.Error(w, "failed to write response", http.StatusInternalServerError)
			return
		}

		attendanceChan <- AttendanceMessage{UUID: recordId, Attendance: attendance}
	} else {
		w.WriteHeader(http.StatusOK)
	}
}
