package sheets

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/arghyadipchak/insti-attend-gsheets/internal/config"
	"github.com/arghyadipchak/insti-attend-gsheets/internal/msg"
	"github.com/arghyadipchak/insti-attend-gsheets/internal/util"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

var (
	logO = log.New(os.Stdout, "[sheet] ", log.LstdFlags)
	logE = log.New(os.Stderr, "[sheet] ", log.LstdFlags)
)

func Runner() {
	defer close(msg.SheetStopped)

	srv, err := sheets.NewService(context.Background(), option.WithCredentialsFile(config.CredentialsFile))
	if err != nil {
		logE.Println("failed to retrieve sheets client:", err)
		return
	}

	if spreadsheet, err := srv.Spreadsheets.Get(config.SpreadsheetId).Do(); err != nil {
		logE.Println("failed to retrieve spreadsheet:", err)
		return
	} else {
		filteredSheets := util.FilterSheets(spreadsheet.Sheets, config.SheetFilters)

		logO.Println("spreadsheet retrieved ::")
		logO.Println("  id:", config.SpreadsheetId)
		logO.Println("  name:", spreadsheet.Properties.Title)
		logO.Printf("  sheets (%d): %s", len(filteredSheets), util.JoinSheetNames(filteredSheets))

	}

	logO.Println("serving")

	for msg := range msg.AttendanceQueue {
		logO.Println("processing attendance:", msg.UUID)

		spreadsheet, err := srv.Spreadsheets.Get(config.SpreadsheetId).Do()
		if err != nil {
			logE.Println("failed to retrieve spreadsheet:", err)
			return
		}

		filteredSheets := util.FilterSheets(spreadsheet.Sheets, config.SheetFilters)
		if len(filteredSheets) == 0 {
			logE.Println("no sheets found")
			continue
		}

		for _, filter := range filteredSheets {
			sheetName := filter.Name

			resp, err := srv.Spreadsheets.Values.Get(config.SpreadsheetId, sheetName).Do()
			if err != nil {
				logE.Printf("failed to retrieve data from sheet %s: %v", sheetName, err)
				continue
			}

			if len(resp.Values) == 0 {
				logE.Println("empty sheet:", sheetName)
				continue
			}

			if len(resp.Values) < int(config.RowHeader) {
				logE.Printf("sheet %s has insufficient rows for header at row %d", sheetName, config.RowHeader)
				continue
			}
			if len(resp.Values[config.RowHeader-1]) <= int(config.ColStartIndex) {
				logE.Printf("sheet %s has insufficient columns starting from index %d", sheetName, config.ColStartIndex)
				continue
			}

			colMap := make(map[string]string)
			if config.ColIsDate {
				for i, colName := range resp.Values[config.RowHeader-1][config.ColStartIndex:] {
					colMap[colName.(string)] = util.ColumnIndexToLetter(uint32(i) + config.ColStartIndex)
				}
			}

			var updates []*sheets.ValueRange
			for i, row := range resp.Values[config.RowStart-1:] {
				if len(row) <= int(config.ColRollIndex) {
					logE.Printf("row %d has insufficient columns for roll number at index %d", i+int(config.RowStart), config.ColRollIndex)
					continue
				}

				rollNo := row[config.ColRollIndex].(string)
				if record, exists := msg.Attendance[rollNo]; exists {
					if !filter.Matches(record.Timestamp) {
						continue
					}

					rowValue := config.RowFormat
					if config.RowIsTime {
						rowValue = record.Timestamp.Format(config.RowFormat)
					}

					if config.ColIsDate {
						if dateCol, found := colMap[record.Timestamp.Format(config.ColFormat)]; found {
							rb := &sheets.ValueRange{
								Range:  fmt.Sprintf("%s!%s%d", sheetName, dateCol, uint32(i)+config.RowStart),
								Values: [][]interface{}{{rowValue}},
							}
							updates = append(updates, rb)
						}
					} else {
						rb := &sheets.ValueRange{
							Range:  fmt.Sprintf("%s!%s%d", sheetName, config.ColFormat, uint32(i)+config.RowStart),
							Values: [][]interface{}{{rowValue}},
						}
						updates = append(updates, rb)
					}

					if config.ColComment != "" && record.Comment != "" {
						rb := &sheets.ValueRange{
							Range:  fmt.Sprintf("%s!%s%d", sheetName, config.ColComment, uint32(i)+config.RowStart),
							Values: [][]interface{}{{record.Comment}},
						}
						updates = append(updates, rb)
					}
				}
			}

			if len(updates) > 0 {
				rb := &sheets.BatchUpdateValuesRequest{
					ValueInputOption: "RAW",
					Data:             updates,
				}
				_, err = srv.Spreadsheets.Values.BatchUpdate(config.SpreadsheetId, rb).Do()
				if err != nil {
					logE.Printf("failed to update sheet %s: %v", sheetName, err)
				} else {
					logO.Printf("updated sheet: %s (%d entries)", sheetName, len(updates))
				}
			}
		}
	}

	logO.Println("stopped")
}
