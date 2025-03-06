package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

var (
	spreadsheetId string

	credentialsFile = "credentials.json"
	webhookAddr     = ":8080"
	authToken       string

	attendanceChan  = make(chan map[string]AttendanceRecord, 10)
	attenderStopped = make(chan struct{})
	webhookStopped  = make(chan struct{})
)

func init() {
	if value, exists := os.LookupEnv("SPREADSHEET_ID"); exists {
		spreadsheetId = value
	} else {
		log.Fatal("env variable not set: SPREADSHEET_ID")
	}

	if value, exists := os.LookupEnv("CREDENTIALS_FILE"); exists {
		credentialsFile = value
	}

	if value, exists := os.LookupEnv("WEBHOOK_ADDR"); exists {
		webhookAddr = value
	}

	if value, exists := os.LookupEnv("AUTH_TOKEN"); exists {
		authToken = value
	}
}

func main() {
	go attender()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	server := &http.Server{Addr: webhookAddr, Handler: http.HandlerFunc(webhookHandler)}

	go func() {
		defer close(webhookStopped)

		log.Printf("[webhook] starting on %s", webhookAddr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("[webhook] failed to start server: %v", err)
		}
	}()

	select {
	case <-attenderStopped:
		if err := server.Shutdown(context.Background()); err != nil {
			log.Printf("[webhook] failed to shutdown server: %v", err)
		} else {
			<-webhookStopped
			log.Printf("[webhook] stopped")
		}

	case <-webhookStopped:
		close(attendanceChan)
		<-attenderStopped

	case <-sigChan:
		println()
		close(attendanceChan)
		if err := server.Shutdown(context.Background()); err != nil {
			log.Printf("[webhook] failed to shutdown server: %v", err)
		} else {
			<-webhookStopped
			log.Printf("[webhook] stopped")
		}
		<-attenderStopped
	}
}
