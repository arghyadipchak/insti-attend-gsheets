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
