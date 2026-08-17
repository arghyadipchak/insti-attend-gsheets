package migrations

import (
	"os"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		settings := app.Settings()

		if value, exists := os.LookupEnv("APP_URL"); exists {
			settings.Meta.AppURL = value
		}
		settings.Meta.AppName = "Insti Attend GSheets"
		settings.Meta.HideControls = true

		settings.Logs.MaxDays = 7
		settings.Logs.LogIP = true
		settings.Logs.LogAuthId = true

		if err := app.Save(settings); err != nil {
			return err
		}

		email, exists := os.LookupEnv("USER_EMAIL")
		if !exists {
			return nil
		}

		password, exists := os.LookupEnv("USER_PASSWORD")
		if !exists {
			return nil
		}

		superuserCollection, err := app.FindCollectionByNameOrId(core.CollectionNameSuperusers)
		if err != nil {
			return err
		}

		superUser := core.NewRecord(superuserCollection)
		superUser.SetEmail(email)
		superUser.SetPassword(password)

		return app.Save(superUser)
	}, nil)
}
