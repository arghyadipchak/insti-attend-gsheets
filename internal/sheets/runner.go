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
		logO.Println("spreadsheet retrieved ::")
		logO.Println("  id:", config.SpreadsheetId)
		logO.Println("  name:", spreadsheet.Properties.Title)
		logO.Printf("  sheets (%d): %s", len(spreadsheet.Sheets), util.JoinSheetNames(spreadsheet.Sheets))
	}

	logO.Println("serving")

	for msg := range msg.AttendanceQueue {
		logO.Println("processing attendance:", msg.UUID)

		spreadsheet, err := srv.Spreadsheets.Get(config.SpreadsheetId).Do()
		if err != nil {
			logE.Println("failed to retrieve spreadsheet:", err)
			return
		}

		for _, sheet := range spreadsheet.Sheets {
			sheetName := sheet.Properties.Title

			resp, err := srv.Spreadsheets.Values.Get(config.SpreadsheetId, sheetName).Do()
			if err != nil {
				logE.Printf("failed to retrieve data from sheet %s: %v", sheetName, err)
				continue
			}

			if len(resp.Values) == 0 {
				logE.Println("empty sheet:", sheetName)
				continue
			}

			dateMap := make(map[string]string)
			for i, colName := range resp.Values[config.RowHeader-1][config.ColStartIndex:] {
				dateMap[colName.(string)] = util.ColumnIndexToLetter(uint32(i) + config.ColStartIndex)
			}

			var updates []*sheets.ValueRange
			for i, row := range resp.Values[config.RowStart-1:] {
				if len(row) == 0 {
					continue
				}

				rollNo := row[config.ColRollIndex].(string)
				if record, exists := msg.Attendance[rollNo]; exists {
					if dateCol, found := dateMap[record.Timestamp.Format(config.ColDateFormat)]; found {
						rowValue := config.RowFormat
						if config.RowIsTime {
							rowValue = record.Timestamp.Format(config.RowFormat)
						}

						rb := &sheets.ValueRange{
							Range:  fmt.Sprintf("%s!%s%d", sheetName, dateCol, uint32(i)+config.RowStart),
							Values: [][]interface{}{{rowValue}},
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
					logO.Printf("updated sheet: %s (%d attendes)", sheetName, len(updates))
				}
			}
		}
	}

	logO.Println("stopped")
}
