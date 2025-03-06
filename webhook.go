package main

import (
	"io"
	"log"
	"net/http"
)

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
		log.Printf("[webhook] invalid request method: %s", r.Method)
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
		log.Printf("[webhook] failed to read request body: %v", err)
		http.Error(w, "failed to read request body", http.StatusInternalServerError)
		return
	}

	attendance, err := readAttendance(body)
	if err != nil {
		log.Printf("[webhook] failed to parse JSON: %v", err)
		http.Error(w, "failed to parse JSON", http.StatusBadRequest)
		return
	}

	if len(attendance) > 0 {
		log.Printf("[webhook] received %d attendance records", len(attendance))
		attendanceChan <- attendance
	}

	w.WriteHeader(http.StatusOK)
}
