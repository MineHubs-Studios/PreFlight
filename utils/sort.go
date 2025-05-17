package utils

import (
	"sort"
)

// SortStrings sorts a string slice in-place.
func SortStrings(items []string) {
	sort.Strings(items)
}
