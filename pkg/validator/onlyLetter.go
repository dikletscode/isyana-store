package validator

import (
	"unicode"
)

func IsContainSymbol(input string) bool {
	for _, l := range input {

		if l != '_' && (unicode.IsSymbol(l) || unicode.IsSpace(l) || unicode.IsPunct(l)) {
			return true
		}
	}
	return false

}

func IsNotValidPassword(input string) bool {
	if len(input) <= 8 || len(input) >= 72 {
		return true
	}
	var IsOneUpper, isOneSymbol bool
	for _, l := range input {

		if unicode.IsUpper(l) {
			IsOneUpper = true
		}
		if unicode.IsSymbol(l) || unicode.IsPunct(l) {
			isOneSymbol = true
		}

	}
	if IsOneUpper && isOneSymbol {
		return false
	}
	return true

}
