package collections

import (
	e "github.com/arghyadipchak/insti-attend-gsheets/internal/errors"
	"github.com/pocketbase/pocketbase/core"
)

const (
	Attendances     = "attendances"
	ServiceAccounts = "service_accounts"
	Spreadsheets    = "spreadsheets"
	Tokens          = "tokens"
	Webhooks        = "webhooks"
	Worksheets      = "worksheets"
)

func Register(app core.App) error {
	for _, fn := range []func(core.App) error{
		registerServiceAccountHooks,
		registerSpreadsheetHooks,
		registerWorksheetHooks,
		registerAttendanceHooks,
	} {
		if err := fn(app); err != nil {
			return err
		}
	}
	return nil
}

type collection interface{ Record() *core.Record }

func expandOne(app core.App, record collection, field string) *e.Error {
	errs := app.ExpandRecord(record.Record(), []string{field}, nil)
	if err := errs[field]; err != nil {
		return e.NewError("failed to expand field "+field, err)
	}
	return nil
}
