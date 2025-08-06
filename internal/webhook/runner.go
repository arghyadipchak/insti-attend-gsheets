package webhook

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/arghyadipchak/insti-attend-gsheets/internal/config"
	"github.com/arghyadipchak/insti-attend-gsheets/internal/msg"
)

var (
	logO = log.New(os.Stdout, "[webhook] ", log.LstdFlags)
	logE = log.New(os.Stderr, "[webhook] ", log.LstdFlags)

	stopped = make(chan struct{})
)

func Runner() {
	defer close(msg.WebhookStopped)

	server := &http.Server{Addr: config.WebhookAddr, Handler: http.HandlerFunc(handler)}

	go func() {
		defer close(stopped)

		logO.Println("starting on", config.WebhookAddr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logE.Println("failed to start server:", err)
		}
	}()

	select {
	case <-msg.WebhookStop:
		if err := server.Shutdown(context.Background()); err != nil {
			logE.Println("failed to shutdown server:", err)
		} else {
			<-stopped
			logO.Println("stopped")
		}

	case <-stopped:
	}
}
