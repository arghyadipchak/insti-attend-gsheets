package collections

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/pocketbase/pocketbase/core"

	"github.com/arghyadipchak/insti-attend-gsheets/internal/sheets"
	u "github.com/arghyadipchak/insti-attend-gsheets/internal/utils"
)

var _ core.RecordProxy = (*Worksheet)(nil)

type Worksheet struct{ core.BaseRecordProxy }

func WorksheetFromRecord(r *core.Record) *Worksheet {
	return &Worksheet{core.BaseRecordProxy{Record: r}}
}

func (w *Worksheet) Record() *core.Record { return w.BaseRecordProxy.Record }

func (w *Worksheet) TabName() string { return w.GetString("tab_name") }

func (w *Worksheet) Weekdays() []string    { return w.GetStringSlice("weekdays") }
func (w *Worksheet) ActiveFrom() time.Time { return w.GetDateTime("active_from").Time().Local() }
func (w *Worksheet) ActiveTo() time.Time   { return w.GetDateTime("active_to").Time().Local() }
func (w *Worksheet) TimeOfDayOnly() bool   { return w.GetBool("time_of_day_only") }

func (w *Worksheet) ColRoll() int {
	return u.ColumnLetterToIndex(w.GetString("col_roll"))
}
func (w *Worksheet) ColHeaderFormat() string { return w.GetString("col_header_format") }
func (w *Worksheet) ColHeaderAutoAdd() bool  { return w.GetBool("col_header_auto_add") }

func (w *Worksheet) ColRange() (int, int) {
	colFrom, colTo, _ := strings.Cut(w.GetString("col_range"), ":")

	from := 0
	if colFrom != "" {
		from = u.ColumnLetterToIndex(colFrom)
	}

	to := -1
	if colTo != "" {
		to = u.ColumnLetterToIndex(colTo)
	}

	return from, to
}

func (w *Worksheet) RowHeader() int          { return w.GetInt("row_header") }
func (w *Worksheet) RowValueFormat() string  { return w.GetString("row_value_format") }
func (w *Worksheet) RowValueOverwrite() bool { return w.GetBool("row_value_overwrite") }

func (w *Worksheet) RowRange() (int, int) {
	rowFrom, rowTo, _ := strings.Cut(w.GetString("row_range"), ":")

	from := w.RowHeader() + 1
	if rowFrom != "" {
		if idx, err := strconv.Atoi(rowFrom); err == nil {
			from = idx
		}
	}

	to := -1
	if rowTo != "" {
		if idx, err := strconv.Atoi(rowTo); err == nil {
			to = idx
		}
	}

	return from, to
}

func (w *Worksheet) GWorksheet() *sheets.Worksheet {
	colFrom, colTo := w.ColRange()
	rowFrom, rowTo := w.RowRange()

	return sheets.NewWorksheet(
		w.TabName(),
		w.ColRoll(), colFrom, colTo,
		w.ColHeaderFormat(), w.ColHeaderAutoAdd(),
		w.RowHeader(), rowFrom, rowTo,
		w.RowValueFormat(), w.RowValueOverwrite(),
		w.Weekdays(), w.ActiveFrom(), w.ActiveTo(), w.TimeOfDayOnly(),
	)
}

func registerWorksheetHooks(app core.App) error {
	app.OnRecordValidate(Worksheets).BindFunc(validateWorksheet)
	return nil
}

func validateWorksheet(re *core.RecordEvent) error {
	w := WorksheetFromRecord(re.Record)

	if colFrom, colTo := w.ColRange(); colTo != -1 && colFrom > colTo {
		return errors.New("invalid col_range")
	}
	if rowFrom, rowTo := w.RowRange(); rowTo != -1 && rowFrom > rowTo {
		return errors.New("invalid row_range")
	}

	return re.Next()
}
