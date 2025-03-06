package main

import (
	"context"
	"fmt"
	"log"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

func attender() {
	defer close(attenderStopped)

	srv, err := sheets.NewService(context.Background(), option.WithCredentialsFile(credentialsFile))
	if err != nil {
		log.Printf("[attender] failed to retrieve sheets client: %v", err)
		return
	}

	if _, err := srv.Spreadsheets.Get(spreadsheetId).Do(); err != nil {
		log.Printf("[attender] failed to retrieve spreadsheet: %v", err)
		return
	}

	log.Println("[attender] started")

	for attendance := range attendanceChan {
		log.Println("[attender] processing attendance:", attendance)

		spreadsheet, err := srv.Spreadsheets.Get(spreadsheetId).Do()
		if err != nil {
			log.Printf("[attender] failed to retrieve spreadsheet: %v", err)
			return
		}

		for _, sheet := range spreadsheet.Sheets {
			sheetName := sheet.Properties.Title

			resp, err := srv.Spreadsheets.Values.Get(spreadsheetId, sheetName).Do()
			if err != nil {
				log.Printf("[attender] failed to retrieve data from sheet %s: %v", sheetName, err)
				continue
			}

			if len(resp.Values) == 0 {
				log.Printf("[attender] no data found in sheet %s", sheetName)
				continue
			}

			dateIndexMap := make(map[string]int)
			for i, colName := range resp.Values[0] {
				dateIndexMap[colName.(string)] = i
			}

			var updates []*sheets.ValueRange
			for i, row := range resp.Values[1:] {
				if len(row) == 0 {
					continue
				}

				rollNo := row[0].(string)
				if record, exists := attendance[rollNo]; exists {
					dateCol := record.Timestamp.Format("2 Jan")
					if dateIndex, found := dateIndexMap[dateCol]; found {
						cellRange := fmt.Sprintf("%s!%s%d", sheetName, columnIndexToLetter(dateIndex), i+2)
						rb := &sheets.ValueRange{
							Range:  cellRange,
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
					log.Printf("[attender] failed to write data to sheet %s: %v", sheetName, err)
				} else {
					log.Printf("[attender] sheet updated: %s", sheetName)
				}
			}
		}
	}

	log.Println("[attender] stopped")
}
