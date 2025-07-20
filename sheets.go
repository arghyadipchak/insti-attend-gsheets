package main

import (
	"context"
	"fmt"
	"log"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

func sheet() {
	defer close(sheetStopped)

	srv, err := sheets.NewService(context.Background(), option.WithCredentialsFile(credentialsFile))
	if err != nil {
		log.Println("[sheet] failed to retrieve sheets client:", err)
		return
	}

	if spreadsheet, err := srv.Spreadsheets.Get(spreadsheetId).Do(); err != nil {
		log.Println("[sheet] failed to retrieve spreadsheet:", err)
		return
	} else {
		log.Println("[sheet] spreadsheet retrieved ::")
		log.Println("[sheet]   id:", spreadsheetId)
		log.Println("[sheet]   name:", spreadsheet.Properties.Title)
		log.Printf("[sheet]   sheets (%d): %s", len(spreadsheet.Sheets), getSheetNames(spreadsheet.Sheets))
	}

	log.Println("[sheet] serving")

	for msg := range attendanceChan {
		log.Println("[sheet] processing attendance:", msg.UUID)

		spreadsheet, err := srv.Spreadsheets.Get(spreadsheetId).Do()
		if err != nil {
			log.Println("[sheet] failed to retrieve spreadsheet:", err)
			return
		}

		for _, sheet := range spreadsheet.Sheets {
			sheetName := sheet.Properties.Title

			resp, err := srv.Spreadsheets.Values.Get(spreadsheetId, sheetName).Do()
			if err != nil {
				log.Printf("[sheet] failed to retrieve data from sheet %s: %v", sheetName, err)
				continue
			}

			if len(resp.Values) == 0 {
				log.Println("[sheet] empty sheet:", sheetName)
				continue
			}

			dateIndexMap := make(map[string]int)
			for i, colName := range resp.Values[0] {
				dateIndexMap[colName.(string)] = i
			}

			var updates []*sheets.ValueRange
			for i, row := range resp.Values[colDateIndex:] {
				if len(row) == 0 {
					continue
				}

				rollNo := row[colRollIndex].(string)
				if record, exists := msg.Attendance[rollNo]; exists {
					if dateIndex, found := dateIndexMap[record.Timestamp.Format(colDateFormat)]; found {
						rb := &sheets.ValueRange{
							Range:  fmt.Sprintf("%s!%s%d", sheetName, columnIndexToLetter(dateIndex), i+colDateIndex+1),
							Values: [][]interface{}{{"P"}},
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
				_, err = srv.Spreadsheets.Values.BatchUpdate(spreadsheetId, rb).Do()
				if err != nil {
					log.Printf("[sheet] failed to update sheet %s: %v", sheetName, err)
				} else {
					log.Printf("[sheet] updated sheet: %s (%d attendes)", sheetName, len(updates))
				}
			}
		}
	}

	log.Println("[sheet] stopped")
}
