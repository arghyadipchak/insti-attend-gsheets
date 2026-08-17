package sheets

import (
	"sort"
	"strconv"
	"time"

	"github.com/arghyadipchak/insti-attend-gsheets/internal/attendances"
	u "github.com/arghyadipchak/insti-attend-gsheets/internal/utils"
	"github.com/maniartech/gotime/v2"

	gsheets "google.golang.org/api/sheets/v4"
)

const noBound = -1

type Worksheet struct {
	name string

	colRoll      int
	colRangeFrom int
	colRangeTo   int

	colHeaderFormat  string
	colHeaderAutoAdd bool

	rowHeader    int
	rowRangeFrom int
	rowRangeTo   int

	rowValueFormat    string
	rowValueOverwrite bool

	schedule
}

func NewWorksheet(
	name string,
	colRoll, colRangeFrom, colRangeTo int,
	colHeaderFormat string, colHeaderAutoAdd bool,
	rowHeader, rowRangeFrom, rowRangeTo int,
	rowValueFormat string, rowValueOverwrite bool,
	allowedWeekdays []string, activeFrom, activeTo time.Time, timeOfDayOnly bool,
) *Worksheet {
	return &Worksheet{
		name: name,

		colRoll:      colRoll,
		colRangeFrom: colRangeFrom,
		colRangeTo:   colRangeTo,

		colHeaderFormat:  colHeaderFormat,
		colHeaderAutoAdd: colHeaderAutoAdd,

		rowHeader:    rowHeader,
		rowRangeFrom: rowRangeFrom,
		rowRangeTo:   rowRangeTo,

		rowValueFormat:    rowValueFormat,
		rowValueOverwrite: rowValueOverwrite,

		schedule: newSchedule(allowedWeekdays, activeFrom, activeTo, timeOfDayOnly),
	}
}

func (w *Worksheet) headerRange() string {
	fromCol := u.ColumnIndexToLetter(w.colRangeFrom)
	rowStr := strconv.Itoa(w.rowHeader)
	if w.colRangeTo == noBound {
		return sheetRange(w.name, fromCol+rowStr+":"+rowStr)
	}

	toCol := u.ColumnIndexToLetter(w.colRangeTo)
	return sheetRange(w.name, fromCol+rowStr+":"+toCol+rowStr)
}

func (w *Worksheet) rollRange() string {
	col := u.ColumnIndexToLetter(w.colRoll)
	fromStr := strconv.Itoa(w.rowRangeFrom)
	if w.rowRangeTo == noBound {
		return sheetRange(w.name, col+fromStr+":"+col)
	}

	toStr := strconv.Itoa(w.rowRangeTo)
	return sheetRange(w.name, col+fromStr+":"+col+toStr)
}

func (w *Worksheet) parseHeaders(row []any) map[string]string {
	columns := make(map[string]string, len(row))
	for offset, cell := range row {
		colIndex := w.colRangeFrom + offset
		if colIndex == w.colRoll {
			continue
		}
		if header, ok := cell.(string); ok && header != "" {
			columns[header] = u.ColumnIndexToLetter(colIndex)
		}
	}
	return columns
}

func (w *Worksheet) parseRollRows(col []any) map[string]int {
	rows := make(map[string]int, len(col))
	for offset, cell := range col {
		roll, ok := cell.(string)
		if !ok || roll == "" {
			continue
		}
		rows[u.NormalizeRollNumber(roll)] = w.rowRangeFrom + offset
	}
	return rows
}

type plannedUpdate struct {
	rollNumber string
	row        int
	header     string
	value      string
}

func (w *Worksheet) planUpdates(
	headerRow, rollColumn []any,
	records attendances.Attendance,
) (updates []plannedUpdate, newHeaders []string, headers map[string]string) {
	headers = w.parseHeaders(headerRow)
	rollRows := w.parseRollRows(rollColumn)

	seen := map[string]bool{}

	for rollNumber, record := range records {
		row, ok := rollRows[u.NormalizeRollNumber(rollNumber)]
		if !ok {
			continue
		}

		timestamp := record.Timestamp.Local()
		if !w.matches(timestamp) {
			continue
		}

		header := gotime.Format(timestamp, w.colHeaderFormat)

		if _, exists := headers[header]; !exists {
			if !w.colHeaderAutoAdd {
				continue
			}

			if !seen[header] {
				seen[header] = true
				newHeaders = append(newHeaders, header)
			}
		}

		updates = append(updates, plannedUpdate{
			rollNumber: rollNumber,
			row:        row,
			header:     header,
			value:      gotime.Format(timestamp, w.rowValueFormat),
		})
	}

	sort.Strings(newHeaders)
	return updates, newHeaders, headers
}

func (w *Worksheet) resolve(
	updates []plannedUpdate,
	headers map[string]string,
	newColumns map[string]string,
) (map[string]*gsheets.ValueRange, map[string]string) {
	for header, col := range newColumns {
		headers[header] = col
	}

	out := make(map[string]*gsheets.ValueRange, len(updates))
	targetCells := make(map[string]string, len(updates))

	for _, pu := range updates {
		col, ok := headers[pu.header]
		if !ok {
			continue
		}

		cell := col + strconv.Itoa(pu.row)
		out[pu.rollNumber] = &gsheets.ValueRange{
			Range:  sheetRange(w.name, cell),
			Values: [][]any{{pu.value}},
		}
		targetCells[pu.rollNumber] = cell
	}

	return out, targetCells
}
