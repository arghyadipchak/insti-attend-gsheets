package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/arghyadipchak/insti-attend-gsheets/internal/config"
	"github.com/arghyadipchak/insti-attend-gsheets/internal/msg"
	"github.com/arghyadipchak/insti-attend-gsheets/internal/sheets"
	"github.com/arghyadipchak/insti-attend-gsheets/internal/webhook"
)

func main() {
	config.Init()

	go sheets.Runner()
	go webhook.Runner()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	select {
	case <-msg.SheetStopped:
		close(msg.WebhookStop)
		<-msg.WebhookStopped

	case <-msg.WebhookStopped:
		close(msg.AttendanceQueue)
		<-msg.SheetStopped

	case <-sigChan:
		log.Println("[attender] received shutdown signal, stopping services...")
		close(msg.WebhookStop)
		close(msg.AttendanceQueue)
		<-msg.SheetStopped
		<-msg.WebhookStopped
	}
}
