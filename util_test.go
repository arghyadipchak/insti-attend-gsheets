package main

import "testing"

func TestColumnIndexToLetter(t *testing.T) {
	tests := []struct {
		index    int
		expected string
	}{
		{0, "A"},
		{1, "B"},
		{25, "Z"},
		{26, "AA"},
		{27, "AB"},
		{51, "AZ"},
		{52, "BA"},
		{701, "ZZ"},
		{702, "AAA"},
	}

	for _, test := range tests {
		result := columnIndexToLetter(test.index)
		if result != test.expected {
			t.Errorf("columnIndexToLetter(%d) = %s; expected %s", test.index, result, test.expected)
		}
	}
}

func TestColumnLetterToIndex(t *testing.T) {
	tests := []struct {
		letter   string
		expected int
	}{
		{"A", 0},
		{"B", 1},
		{"Z", 25},
		{"AA", 26},
		{"AB", 27},
		{"AZ", 51},
		{"BA", 52},
		{"ZZ", 701},
		{"AAA", 702},
	}

	for _, test := range tests {
		result := columnLetterToIndex(test.letter)
		if result != test.expected {
			t.Errorf("letterToColumnIndex(%s) = %d; expected %d", test.letter, result, test.expected)
		}
	}
}
