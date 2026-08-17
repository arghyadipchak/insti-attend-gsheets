package sheets

import (
	"errors"
	"fmt"

	"github.com/pocketbase/pocketbase/core"
	gsheets "google.golang.org/api/sheets/v4"

	"github.com/arghyadipchak/insti-attend-gsheets/internal/attendances"
	e "github.com/arghyadipchak/insti-attend-gsheets/internal/errors"
	u "github.com/arghyadipchak/insti-attend-gsheets/internal/utils"
)

func sheetIDCacheKey(spreadsheetId, name string) string { return spreadsheetId + "\x00" + name }

type candidate struct {
	rollNumber string
	cellRange  string
	update     *gsheets.ValueRange
}

func (c *Client) Sync(
	app core.App,
	spreadsheetId string,
	ws *Worksheet,
	records attendances.Attendance,
) (synced map[string]struct{}, errs []*e.Error) {
	resp, err := c.Spreadsheets.Values.BatchGet(spreadsheetId).
		Ranges(ws.headerRange(), ws.rollRange()).Do()
	if err != nil {
		errs = append(errs, e.NewError("failed to retrieve header and roll ranges", err))
		return
	}
	if len(resp.ValueRanges) != 2 {
		errMsg := fmt.Errorf("expected 2 ranges, got %d", len(resp.ValueRanges))
		errs = append(errs, e.NewError("unexpected value ranges count", errMsg))
		return
	}

	headerRow := valuesRow(resp.ValueRanges[0])
	rollColumn := columnValues(resp.ValueRanges[1])

	updates, newHeaders, headers := ws.planUpdates(headerRow, rollColumn, records)
	if len(updates) == 0 {
		return
	}

	if len(newHeaders) > 0 {
		cols, err := c.addHeaderColumns(spreadsheetId, ws, headerRow, newHeaders)
		if err == nil {
			for h, col := range cols {
				headers[h] = col
			}
		} else {
			errs = append(errs, err)
		}
	}

	worksheetUpdates, targetCells := ws.resolve(updates, headers, nil)

	unconditional := []*gsheets.ValueRange{}
	var checked []candidate

	synced = make(map[string]struct{})

	for rollNo, update := range worksheetUpdates {
		if ws.rowValueOverwrite {
			unconditional = append(unconditional, update)
			synced[rollNo] = struct{}{}
			continue
		}

		checked = append(checked, candidate{
			rollNumber: rollNo,
			cellRange:  sheetRange(ws.name, targetCells[rollNo]),
			update:     update,
		})
	}

	finalUpdates := unconditional

	if len(checked) > 0 {
		ranges := make([]string, len(checked))
		for i, chk := range checked {
			ranges[i] = chk.cellRange
		}

		resp, err := c.Spreadsheets.Values.BatchGet(spreadsheetId).Ranges(ranges...).Do()
		if err == nil {
			for i, chk := range checked {
				if i < len(resp.ValueRanges) && cellHasValue(resp.ValueRanges[i]) {
					continue
				}

				finalUpdates = append(finalUpdates, chk.update)
				synced[chk.rollNumber] = struct{}{}
			}
		} else {
			errs = append(errs, e.NewError("failed to check existing cell values", err))
		}
	}

	if len(finalUpdates) > 0 {
		batchRequest := &gsheets.BatchUpdateValuesRequest{
			ValueInputOption: "USER_ENTERED",
			Data:             finalUpdates,
		}
		req := c.Spreadsheets.Values.BatchUpdate(spreadsheetId, batchRequest)
		if _, err := req.Do(); err != nil {
			errs = append(errs, e.NewError("failed to batch update values", err))
		}
	}

	return
}

func (c *Client) addHeaderColumns(
	spreadsheetId string,
	ws *Worksheet,
	headerRow []any,
	headers []string,
) (map[string]string, *e.Error) {
	sheetId, err := c.sheetIdByName(spreadsheetId, ws.name)
	if err != nil {
		return nil, e.NewError("failed to resolve sheet ID", err)
	}

	insertAt := ws.colRangeFrom + len(headerRow)

	cells := make([]*gsheets.CellData, len(headers))
	cols := make(map[string]string, len(headers))
	for i := range headers {
		cells[i] = &gsheets.CellData{
			UserEnteredValue: &gsheets.ExtendedValue{StringValue: &headers[i]},
		}
		cols[headers[i]] = u.ColumnIndexToLetter(insertAt + i)
	}

	requests := []*gsheets.Request{
		{
			InsertDimension: &gsheets.InsertDimensionRequest{
				Range: &gsheets.DimensionRange{
					SheetId:    sheetId,
					Dimension:  "COLUMNS",
					StartIndex: int64(insertAt),
					EndIndex:   int64(insertAt + len(headers)),
				},
				InheritFromBefore: insertAt > ws.colRangeFrom,
			},
		},
		{
			UpdateCells: &gsheets.UpdateCellsRequest{
				Range: &gsheets.GridRange{
					SheetId:          sheetId,
					StartRowIndex:    int64(ws.rowHeader - 1),
					EndRowIndex:      int64(ws.rowHeader),
					StartColumnIndex: int64(insertAt),
					EndColumnIndex:   int64(insertAt + len(headers)),
				},
				Rows:   []*gsheets.RowData{{Values: cells}},
				Fields: "userEnteredValue",
			},
		},
	}

	if _, err := c.Spreadsheets.BatchUpdate(spreadsheetId, &gsheets.BatchUpdateSpreadsheetRequest{
		Requests: requests,
	}).Do(); err != nil {
		c.sheetIDCache.Delete(sheetIDCacheKey(spreadsheetId, ws.name))
		return nil, e.NewError("failed to insert columns and write headers", err)
	}

	return cols, nil
}

func (c *Client) sheetIdByName(spreadsheetId, name string) (int64, error) {
	cacheKey := sheetIDCacheKey(spreadsheetId, name)
	if v, ok := c.sheetIDCache.Load(cacheKey); ok {
		return v.(int64), nil
	}

	resp, err := c.Spreadsheets.Get(spreadsheetId).Fields("sheets.properties").Do()
	if err != nil {
		return 0, err
	}

	var (
		sheetId int64
		found   bool
	)

	for _, sheet := range resp.Sheets {
		key := sheetIDCacheKey(spreadsheetId, sheet.Properties.Title)
		c.sheetIDCache.Store(key, sheet.Properties.SheetId)
		if sheet.Properties.Title == name {
			sheetId, found = sheet.Properties.SheetId, true
		}
	}

	if !found {
		return 0, errors.New("sheet not found")
	}

	return sheetId, nil
}

func valuesRow(vr *gsheets.ValueRange) []any {
	if vr == nil || len(vr.Values) == 0 {
		return nil
	}
	return vr.Values[0]
}

func columnValues(vr *gsheets.ValueRange) []any {
	if vr == nil {
		return nil
	}
	col := make([]any, len(vr.Values))
	for i, row := range vr.Values {
		if len(row) > 0 {
			col[i] = row[0]
		}
	}
	return col
}

func cellHasValue(vr *gsheets.ValueRange) bool {
	return vr != nil && len(vr.Values) > 0 && len(vr.Values[0]) > 0
}
