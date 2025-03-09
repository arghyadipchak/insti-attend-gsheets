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
		log.Println("[attender] failed to retrieve sheets client:", err)
		return
	}

	if spreadsheet, err := srv.Spreadsheets.Get(spreadsheetId).Do(); err != nil {
		log.Println("[attender] failed to retrieve spreadsheet:", err)
		return
	} else {
		log.Println("[attender] spreadsheet retrieved ::")
		log.Println("[attender]   id:", spreadsheetId)
		log.Println("[attender]   name:", spreadsheet.Properties.Title)
		log.Printf("[attender]   sheets (%d): %s", len(spreadsheet.Sheets), getSheetNames(spreadsheet.Sheets))
	}

	log.Println("[attender] serving")

	for msg := range attendanceChan {
		log.Println("[attender] processing attendance:", msg.UUID)

		spreadsheet, err := srv.Spreadsheets.Get(spreadsheetId).Do()
		if err != nil {
			log.Println("[attender] failed to retrieve spreadsheet:", err)
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
				log.Println("[attender] empty sheet:", sheetName)
				continue
			}

			dateIndexMap := make(map[string]int)
			for i, colName := range resp.Values[0] {
				dateIndexMap[colName.(string)] = i
			}

			var updates []*sheets.ValueRange
			for i, row := range resp.Values[skipRows:] {
				if len(row) == 0 {
					continue
				}

				rollNo := row[rollColIndex].(string)
				if record, exists := msg.Attendance[rollNo]; exists {
					if dateIndex, found := dateIndexMap[record.Timestamp.Format(colDateLayout)]; found {
						rb := &sheets.ValueRange{
							Range:  fmt.Sprintf("%s!%s%d", sheetName, columnIndexToLetter(dateIndex), i+skipRows+1),
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
					log.Printf("[attender] failed to update sheet %s: %v", sheetName, err)
				} else {
					log.Printf("[attender] updated sheet: %s (%d attendes)", sheetName, len(updates))
				}
			}
		}
	}

	log.Println("[attender] stopped")
}
