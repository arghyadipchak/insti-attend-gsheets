package utils

import (
	"slices"
	"strings"
)

func ColumnLetterToIndex(letter string) int {
	index := 0
	for _, char := range letter {
		if char >= 'A' && char <= 'Z' {
			index = index*26 + int(char-'A'+1)
		}
	}
	return index - 1
}

func ColumnIndexToLetter(index int) string {
	bytes := make([]byte, 0, 3)
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

func NormalizeRollNumber(rollNumber string) string {
	return strings.ToUpper(strings.TrimSpace(rollNumber))
}
