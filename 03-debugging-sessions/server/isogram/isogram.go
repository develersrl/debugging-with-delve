package isogram

import "unicode"

// IsIsogram takes a string and returns true if it is an isogram
// false otherwise.
//
// see: https://en.wikipedia.org/wiki/Heterogram_(literature)
// for more on isograms.
func IsIsogram(s string) bool {
	found := make(map[rune]bool)

	for _, r := range s {
		if !unicode.IsLetter(r) {
			continue
		}

		r = unicode.ToLower(r)
		if found[r] {
			return false
		}

		found[r] = true
	}

	return true
}
