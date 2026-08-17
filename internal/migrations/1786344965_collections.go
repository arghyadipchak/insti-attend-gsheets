package migrations

import (
	"fmt"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

const (
	Attendances     = "attendances"
	ServiceAccounts = "service_accounts"
	Spreadsheets    = "spreadsheets"
	Tokens          = "tokens"
	Webhooks        = "webhooks"
	Worksheets      = "worksheets"

	manyRelations = 65535

	gotimeDocsURL = "https://github.com/maniartech/gotime/blob/master/docs/" +
		"core-concepts/nites.md#complete-format-reference"
)

var weekdays = []string{
	"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday",
}

func init() {
	m.Register(func(app core.App) error {
		for _, f := range []func(app core.App) error{
			createServiceAccountsCollection,
			createWorksheetsCollection,
			createSpreadsheetsCollection,
			createTokensCollection,
			createWebhooksCollection,
			createAttendancesCollection,
		} {
			if err := f(app); err != nil {
				return err
			}
		}
		return nil
	}, func(app core.App) error {
		for _, name := range []string{
			Attendances,
			Webhooks,
			Tokens,
			Spreadsheets,
			Worksheets,
			ServiceAccounts,
		} {
			if collection, err := app.FindCollectionByNameOrId(name); err == nil {
				if err := app.Delete(collection); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func createServiceAccountsCollection(app core.App) error {
	collection := core.NewBaseCollection(ServiceAccounts)

	collection.Fields.Add(&core.TextField{Name: "name", Presentable: true, Required: true})

	collection.Fields.Add(&core.JSONField{
		Name: "credentials",
		Help: "Google service account key file (JSON); " +
			"share each spreadsheet with client_email as Editor",
		Required: true,
	})

	return app.Save(collection)
}

func createWorksheetsCollection(app core.App) error {
	collection := core.NewBaseCollection(Worksheets)

	collection.Fields.Add(&core.TextField{Name: "name", Presentable: true, Required: true})

	collection.Fields.Add(&core.TextField{
		Name:     "tab_name",
		Help:     "Exact tab name in Google Sheets (case-sensitive)",
		Required: true,
	})

	collection.Fields.Add(&core.SelectField{
		Name:      "weekdays",
		Help:      "Restrict to these weekdays. Empty = any day",
		Values:    weekdays,
		MaxSelect: len(weekdays),
	})

	collection.Fields.Add(&core.DateField{
		Name: "active_from",
		Help: "Start of the active window. Empty = no lower bound",
	})

	collection.Fields.Add(&core.DateField{
		Name: "active_to",
		Help: "End of the active window. Empty = no upper bound",
	})

	collection.Fields.Add(&core.BoolField{
		Name: "time_of_day_only",
		Help: "Treat active_from/active_to as a daily time-of-day window rather than a date range",
	})

	collection.Fields.Add(&core.TextField{
		Name:                "col_roll",
		Help:                "Column with roll numbers. Default: A",
		Pattern:             "^[A-Z]+$",
		AutogeneratePattern: "A",
		Required:            true,
	})

	collection.Fields.Add(&core.TextField{
		Name:                "col_range",
		Help:                "Column range to search for headers, e.g. \"B:D\". Default: B:",
		Pattern:             "^[A-Z]*:[A-Z]*$",
		AutogeneratePattern: "B:",
		Required:            true,
	})

	collection.Fields.Add(&core.TextField{
		Name: "col_header_format",
		Help: "Format of the column header " +
			"(e.g. \"d mmm\" for \"10 Aug\", \"yyyy-mm-dd\", \"dd/mm/yyyy\"); " +
			"see " + gotimeDocsURL,
		Required: true,
	})

	collection.Fields.Add(&core.BoolField{
		Name: "col_header_auto_add",
		Help: "Auto-insert a new column header when no matching header is found; " +
			"skips the record otherwise",
	})

	collection.Fields.Add(&core.TextField{
		Name:                "row_header",
		Help:                "Row containing the headers. Default: 1",
		Pattern:             "^[0-9]+$",
		AutogeneratePattern: "1",
		Required:            true,
	})

	collection.Fields.Add(&core.TextField{
		Name:                "row_range",
		Help:                "Row range to search for data, e.g. \"2:10\". Default: 2:",
		Pattern:             "^[0-9]*:[0-9]*$",
		AutogeneratePattern: "2:",
		Required:            true,
	})

	collection.Fields.Add(&core.TextField{
		Name: "row_value_format",
		Help: "Value to write into the cell: fixed text (e.g. \"P\" for present) " +
			"or datetime format (e.g. \"hhhh:ii:ss\"). Default: P; " +
			"see " + gotimeDocsURL,
		AutogeneratePattern: "P",
		Required:            true,
	})

	collection.Fields.Add(&core.BoolField{
		Name: "row_value_overwrite",
		Help: "Overwrite existing cell values; when disabled only empty cells are filled",
	})

	return app.Save(collection)
}

func createSpreadsheetsCollection(app core.App) error {
	related, err := findCollections(app, ServiceAccounts, Worksheets)
	if err != nil {
		return err
	}

	collection := core.NewBaseCollection(Spreadsheets)

	collection.Fields.Add(&core.TextField{Name: "name", Presentable: true, Required: true})

	collection.Fields.Add(&core.TextField{
		Name:     "spreadsheet_id",
		Help:     "From the sheet URL: .../d/{spreadsheet_id}/edit",
		Required: true,
	})

	collection.Fields.Add(&core.TextField{
		Name: "title",
		Help: "Auto-fetched from Google Sheets",
	})

	collection.Fields.Add(&core.RelationField{
		Name:         "service_account",
		Help:         "Must have Editor access to this spreadsheet",
		CollectionId: related[ServiceAccounts].Id,
		Required:     true,
	})

	collection.Fields.Add(&core.RelationField{
		Name:         Worksheets,
		CollectionId: related[Worksheets].Id,
		MaxSelect:    manyRelations,
		Required:     false,
	})

	return app.Save(collection)

}

func createTokensCollection(app core.App) error {
	collection := core.NewBaseCollection(Tokens)

	collection.Fields.Add(&core.TextField{Name: "name", Presentable: true, Required: true})

	collection.Fields.Add(&core.TextField{
		Name:                "secret",
		Help:                "Sent by callers as \"Authorization: Bearer {secret}\"",
		Min:                 32,
		Pattern:             "[A-Za-z0-9]+",
		AutogeneratePattern: "[A-Za-z0-9]{64}",
		Required:            true,
	})

	return app.Save(collection)
}

func createWebhooksCollection(app core.App) error {
	related, err := findCollections(app, Tokens, Spreadsheets)
	if err != nil {
		return err
	}

	collection := core.NewBaseCollection(Webhooks)

	collection.Fields.Add(&core.TextField{Name: "name", Presentable: true, Required: true})

	collection.Fields.Add(&core.TextField{
		Name:                "slug",
		Help:                "Used in the URL path (POST /wbh/{slug})",
		Pattern:             "^[a-z0-9]+(-[a-z0-9]+)*$",
		AutogeneratePattern: "[a-z0-9]{15}",
		Required:            true,
	})

	collection.Fields.Add(&core.BoolField{
		Name: "enabled",
		Help: "Disabled webhooks reject all requests without deleting the record",
	})

	collection.Fields.Add(&core.RelationField{
		Name:         Tokens,
		CollectionId: related[Tokens].Id,
		MaxSelect:    manyRelations,
	})

	collection.Fields.Add(&core.RelationField{
		Name:         Spreadsheets,
		CollectionId: related[Spreadsheets].Id,
		MaxSelect:    manyRelations,
	})

	collection.AddIndex(Webhooks+"_slug", true, "`slug`", "")

	return app.Save(collection)
}

func createAttendancesCollection(app core.App) error {
	related, err := findCollections(app, Spreadsheets)
	if err != nil {
		return err
	}

	collection := core.NewBaseCollection(Attendances)

	collection.Fields.Add(&core.BoolField{
		Name: "synced",
		Help: "Managed by the cron job; set to false to trigger a re-sync",
	})

	collection.Fields.Add(&core.RelationField{
		Name:         Spreadsheets,
		CollectionId: related[Spreadsheets].Id,
		MaxSelect:    manyRelations,
		Required:     true,
	})

	collection.Fields.Add(&core.JSONField{
		Name:     "records",
		Help:     "roll number → attendance record map; populated from the webhook payload",
		Required: true,
	})

	collection.Fields.Add(&core.NumberField{
		Name:    "record_count",
		Help:    "Total records in this batch (read-only)",
		OnlyInt: true,
	})

	collection.Fields.Add(&core.NumberField{
		Name:    "unsynced_count",
		Help:    "Records not yet synced to Sheets (read-only)",
		OnlyInt: true,
	})

	collection.Fields.Add(&core.AutodateField{Name: "created", OnCreate: true})
	collection.Fields.Add(&core.AutodateField{Name: "updated", OnUpdate: true})

	collection.AddIndex(Attendances+"_synced", false, "`synced`", "")

	return app.Save(collection)
}

func findCollections(app core.App, names ...string) (map[string]*core.Collection, error) {
	out := make(map[string]*core.Collection, len(names))
	for _, name := range names {
		if c, err := app.FindCollectionByNameOrId(name); err == nil {
			out[name] = c
		} else {
			return nil, fmt.Errorf("find collection %q: %w", name, err)
		}
	}
	return out, nil
}
