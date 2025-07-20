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

	colRollIndex  = columnLetterToIndex("A")
	colDateIndex  = columnLetterToIndex("B")
	colDateFormat = "2 Jan"

	credentialsFile = "credentials.json"
	webhookAddr     = ":8080"
	authToken       string

	attendanceChan = make(chan AttendanceMessage, 10)
	sheetStopped   = make(chan struct{})
	webhookStopped = make(chan struct{})
)

func read_env() {
	if value, exists := os.LookupEnv("SPREADSHEET_ID"); exists {
		spreadsheetId = value
	} else {
		log.Fatal("env variable not set: SPREADSHEET_ID")
	}

	if value, exists := os.LookupEnv("CREDENTIALS_FILE"); exists {
		credentialsFile = value
	}

	if value, exists := os.LookupEnv("COL_ROLL"); exists {
		colRollIndex = columnLetterToIndex(value)
	}

	if value, exists := os.LookupEnv("COL_DATE_START"); exists {
		colDateIndex = columnLetterToIndex(value)
	}

	if value, exists := os.LookupEnv("COL_DATE_FORMAT"); exists {
		colDateFormat = value
	}

	if value, exists := os.LookupEnv("WEBHOOK_ADDR"); exists {
		webhookAddr = value
	}

	if value, exists := os.LookupEnv("AUTH_TOKEN"); exists {
		authToken = value
	}
}

func main() {
	read_env()

	go sheet()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	server := &http.Server{Addr: webhookAddr, Handler: http.HandlerFunc(webhookHandler)}

	go func() {
		defer close(webhookStopped)

		log.Println("[webhook] starting on", webhookAddr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Println("[webhook] failed to start server:", err)
		}
	}()

	select {
	case <-sheetStopped:
		if err := server.Shutdown(context.Background()); err != nil {
			log.Println("[webhook] failed to shutdown server:", err)
		} else {
			<-webhookStopped
			log.Println("[webhook] stopped")
		}

	case <-webhookStopped:
		close(attendanceChan)
		<-sheetStopped

	case <-sigChan:
		println()
		close(attendanceChan)
		if err := server.Shutdown(context.Background()); err != nil {
			log.Println("[webhook] failed to shutdown server", err)
		} else {
			<-webhookStopped
			log.Println("[webhook] stopped")
		}
		<-sheetStopped
	}
}
