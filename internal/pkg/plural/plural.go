package utils

import (
	"golang.org/x/exp/constraints"
)

// Declension returns the correct declension of the word depending on the number
func Declension[N constraints.Integer](count N, singular, few, many string) string {
	n := int(count)
	if n%10 == 1 && n%100 != 11 {
		return singular
	} else if (n%10 >= 2 && n%10 <= 4) && !(n%100 >= 12 && n%100 <= 14) {
		return few
	} else {
		return many
	}
}
