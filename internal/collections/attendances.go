package collections

import (
	"os"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/cron"

	"github.com/arghyadipchak/insti-attend-gsheets/internal/attendances"
	e "github.com/arghyadipchak/insti-attend-gsheets/internal/errors"
)

var _ core.RecordProxy = (*Attendance)(nil)

type Attendance struct{ core.BaseRecordProxy }

func AttendanceFromRecord(r *core.Record) *Attendance {
	return &Attendance{core.BaseRecordProxy{Record: r}}
}

func (a *Attendance) Record() *core.Record { return a.BaseRecordProxy.Record }

func (a *Attendance) Synced() bool     { return a.GetBool("synced") }
func (a *Attendance) SetSynced(v bool) { a.Set("synced", v) }

func (a *Attendance) SpreadsheetIDs() []string { return a.GetStringSlice(Spreadsheets) }

func (a *Attendance) Records() (attendances.Attendance, *e.Error) {
	var records attendances.Attendance
	if err := a.UnmarshalJSONField("records", &records); err != nil {
		return nil, e.NewError("failed to unmarshal attendance records", err)
	}
	return records, nil
}

func (a *Attendance) SetRecords(records attendances.Attendance) { a.Set("records", records) }

func (a *Attendance) UpdateCounts() error {
	records, err := a.Records()
	if err != nil {
		return err
	}

	a.Set("record_count", len(records))
	a.Set("unsynced_count", records.CountUnsynced())
	return nil
}

func (a *Attendance) Spreadsheets(app core.App) ([]*Spreadsheet, *e.Error) {
	if err := expandOne(app, a, Spreadsheets); err != nil {
		return nil, err
	}

	records := a.ExpandedAll(Spreadsheets)
	spreadsheets := make([]*Spreadsheet, len(records))
	for i, r := range records {
		spreadsheets[i] = SpreadsheetFromRecord(r)
	}

	return spreadsheets, nil
}

func (a *Attendance) Sync(app core.App) (errs []*e.Error) {
	records, err := a.Records()
	if err != nil {
		errs = append(errs, err)
		return
	}
	if len(records) == 0 {
		return
	}

	spreadsheets, err := a.Spreadsheets(app)
	if err != nil {
		errs = append(errs, err)
		return
	}

	for _, spreadsheet := range spreadsheets {
		synced, serrs := spreadsheet.Sync(app, records)

		for rollNo := range synced {
			records[rollNo].Synced = true
		}

		for _, err := range serrs {
			_ = err.AddContext("spreadsheet_id", spreadsheet.Id)
			errs = append(errs, err)
		}
	}

	a.SetRecords(records)
	return
}

const (
	syncCronID      = "sync_attendance"
	defaultCronExpr = "* * * * *"
)

func registerAttendanceHooks(app core.App) error {
	app.OnRecordValidate(Attendances).BindFunc(validateAttendance)

	cronExpr, exists := os.LookupEnv("CRON_SYNC")
	if !exists {
		cronExpr = defaultCronExpr
	} else if _, err := cron.NewSchedule(cronExpr); err != nil {
		app.Logger().Warn("invalid CRON_SYNC, using default", "value", cronExpr, "error", err)
		cronExpr = defaultCronExpr
	}

	syncFn := func() { syncUnsyncedAttendances(app) }
	if err := app.Cron().Add(syncCronID, cronExpr, syncFn); err != nil {
		return e.NewError("failed to register sync cron job", err)
	}

	return nil
}

func validateAttendance(re *core.RecordEvent) error {
	att := AttendanceFromRecord(re.Record)
	if err := att.UpdateCounts(); err != nil {
		return err
	}
	return re.Next()
}

func findUnsyncedAttendances(app core.App) ([]*Attendance, error) {
	records, err := app.FindAllRecords(Attendances, dbx.HashExp{"synced": false})
	if err != nil {
		return nil, e.NewError("failed to find unsynced attendances", err)
	}

	attendances := make([]*Attendance, len(records))
	for i, record := range records {
		attendances[i] = AttendanceFromRecord(record)
	}
	return attendances, nil
}

func syncUnsyncedAttendances(app core.App) {
	logger := app.Logger()

	unsyncedAttendances, err := findUnsyncedAttendances(app)
	if err != nil {
		logger.Error("sync attendance", "error", err)
		return
	}
	if len(unsyncedAttendances) == 0 {
		return
	}

	for _, attendance := range unsyncedAttendances {
		alogger := logger.With("attendance", attendance.Id)
		for _, err := range attendance.Sync(app) {
			alogger.Error("sync attendance", err.ToArgs()...)
		}

		attendance.SetSynced(true)
		if err := app.Save(attendance); err != nil {
			alogger.Error("sync attendance",
				"error", "failed to save attendance record",
				"details", err)
		}
	}
}

func CreateAttendanceRecord(
	app core.App,
	records attendances.Attendance,
	spreadsheetIds []string,
) (string, *e.Error) {

	collection, err := app.FindCollectionByNameOrId(Attendances)
	if err != nil {
		return "", e.NewError("failed to find attendances collection", err)
	}

	attendance := AttendanceFromRecord(core.NewRecord(collection))

	attendance.Set(Spreadsheets, spreadsheetIds)
	attendance.SetRecords(records)

	if err := app.Save(attendance); err != nil {
		return "", e.NewError("failed to save attendance record", err)
	}

	return attendance.Id, nil
}
