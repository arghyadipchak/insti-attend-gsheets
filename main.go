package main

import (
	"log"

	"github.com/arghyadipchak/insti-attend-gsheets/internal/collections"
	"github.com/arghyadipchak/insti-attend-gsheets/internal/healthcheck"
	_ "github.com/arghyadipchak/insti-attend-gsheets/internal/migrations"
	"github.com/arghyadipchak/insti-attend-gsheets/internal/webhooks"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
	"github.com/pocketbase/pocketbase/tools/osutils"
)

func main() {
	app := pocketbase.New()

	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		Automigrate: osutils.IsProbablyGoRun(),
		Dir:         "internal/migrations",
	})

	if err := collections.Register(app); err != nil {
		log.Fatal(err)
	}

	webhooks.Register(app)
	healthcheck.Register(app)

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
