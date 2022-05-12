package isogram

import "testing"

var testCases = []struct {
	name     string
	input    string
	expected bool
}{
	{
		name:     "empty word",
		input:    "",
		expected: true,
	},
	{
		name:     "isogram",
		input:    "lumberjack",
		expected: true,
	},
	{
		name:     "non-isogram",
		input:    "advanced",
		expected: false,
	},
}

func TestIsogram(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := IsIsogram(tc.input)
			if got != tc.expected {
				t.Fatalf("IsIsogram(%q) returned %t, expected %t", tc.input, got, tc.expected)
			}
		})
	}
}
