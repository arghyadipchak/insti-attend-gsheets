package collections

import (
	"maps"

	"github.com/pocketbase/pocketbase/core"

	"github.com/arghyadipchak/insti-attend-gsheets/internal/attendances"
	e "github.com/arghyadipchak/insti-attend-gsheets/internal/errors"
)

const ServiceAccountField = "service_account"

var _ core.RecordProxy = (*Spreadsheet)(nil)

type Spreadsheet struct{ core.BaseRecordProxy }

func SpreadsheetFromRecord(r *core.Record) *Spreadsheet {
	return &Spreadsheet{core.BaseRecordProxy{Record: r}}
}

func (s *Spreadsheet) Record() *core.Record { return s.BaseRecordProxy.Record }

func (s *Spreadsheet) SpreadsheetID() string { return s.GetString("spreadsheet_id") }

func (s *Spreadsheet) Title() string         { return s.GetString("title") }
func (s *Spreadsheet) SetTitle(title string) { s.Set("title", title) }

func (s *Spreadsheet) ServiceAccount(app core.App) (*ServiceAccount, *e.Error) {
	if err := expandOne(app, s, ServiceAccountField); err != nil {
		return nil, err
	}
	return ServiceAccountFromRecord(s.ExpandedOne(ServiceAccountField)), nil
}

func (s *Spreadsheet) Worksheets(app core.App) ([]*Worksheet, *e.Error) {
	if err := expandOne(app, s, Worksheets); err != nil {
		return nil, err
	}

	records := s.ExpandedAll(Worksheets)
	worksheets := make([]*Worksheet, len(records))
	for i, r := range records {
		worksheets[i] = WorksheetFromRecord(r)
	}

	return worksheets, nil
}

func (s *Spreadsheet) RefreshTitle(app core.App) *e.Error {
	sa, err := s.ServiceAccount(app)
	if err != nil {
		return err
	}

	client, err := sa.NewClient()
	if err != nil {
		return err
	}

	title, err := client.FetchTitle(s.SpreadsheetID())
	if err != nil {
		return err
	}

	s.SetTitle(title)
	return nil
}

func (s *Spreadsheet) Sync(
	app core.App,
	records attendances.Attendance,
) (synced map[string]struct{}, errs []*e.Error) {
	worksheets, err := s.Worksheets(app)
	if err != nil {
		errs = append(errs, err)
		return
	}
	if len(worksheets) == 0 {
		return nil, nil
	}

	sa, err := s.ServiceAccount(app)
	if err != nil {
		return nil, []*e.Error{err}
	}

	client, err := sa.NewClient()
	if err != nil {
		return nil, []*e.Error{err}
	}

	synced = make(map[string]struct{})
	for _, worksheet := range worksheets {
		wSynced, wErrs := client.Sync(app, s.SpreadsheetID(), worksheet.GWorksheet(), records)
		maps.Copy(synced, wSynced)

		for _, err := range wErrs {
			_ = err.AddContext("worksheet_id", worksheet.Id)
			errs = append(errs, err)
		}
	}

	return
}

func registerSpreadsheetHooks(app core.App) error {
	app.OnRecordValidate(Spreadsheets).BindFunc(validateSpreadsheet)
	return nil
}

func validateSpreadsheet(re *core.RecordEvent) error {
	s := SpreadsheetFromRecord(re.Record)
	if err := s.RefreshTitle(re.App); err != nil {
		return err
	}
	return re.Next()
}
