package sheets

import "strings"

func sheetRange(sheetName string, a1Range string) string {
	quotedSheetName := "'" + strings.ReplaceAll(sheetName, "'", "''") + "'"
	if a1Range == "" {
		return quotedSheetName
	}
	return quotedSheetName + "!" + a1Range
}
