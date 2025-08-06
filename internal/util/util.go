package util

import (
	"errors"
	"slices"
	"strings"

	"google.golang.org/api/sheets/v4"
)

func JoinSheetNames(sheets []*sheets.Sheet) string {
	var names []string
	for _, sheet := range sheets {
		names = append(names, sheet.Properties.Title)
	}

	return strings.Join(names, ", ")
}

func ColumnIndexToLetter(index uint32) string {
	var bytes []byte
	for {
		bytes = append(bytes, byte('A'+(index%26)))
		if index < 26 {
			break
		}
		index = index/26 - 1
	}

	slices.Reverse(bytes)
	return string(bytes)
}

func ColumnLetterToIndex(letter string) (index uint32, err error) {
	if len(letter) == 0 {
		err = errors.New("column letter empty")
		return
	}

	for _, char := range letter {
		if char < 'A' || char > 'Z' {
			err = errors.New("invalid column letter: " + letter)
			return
		}
		index = index*26 + uint32(char-'A'+1)
	}

	index--
	return
}
