package model

import (
	"errors"
	"strings"
)

// ErrInvalidPhone is returned for a phone number that can't be read. Its message is shown to the
// person typing the number.
var ErrInvalidPhone = errors.New(
	"enter a Nigerian number with 11 digits starting 0 (e.g. 0803 123 4567), " +
		"or a number starting with + and the country code",
)

const (
	nigeriaCode = "234"
	// nigerianNationalDigits is the length of a Nigerian number without the trunk 0.
	nigerianNationalDigits = 10
	// An international number is at most 15 digits (E.164) and at least a short local one.
	minInternationalDigits = 8
	maxInternationalDigits = 15
)

// NormalizePhone turns a typed phone number into the one form that is stored and searched:
// +234 and the 10 digits of a Nigerian number, or + and the digits of a foreign one. Nigerian
// numbers may be written with or without the trunk 0 (11 digits with it) and with or without
// 234, +234 or 00234. Spaces, dashes, dots and brackets are ignored. Empty input is no phone and
// gives "".
func NormalizePhone(raw string) (string, error) {
	text := strings.TrimSpace(raw)
	if text == "" {
		return "", nil
	}
	plus := strings.HasPrefix(text, "+")
	text = strings.TrimPrefix(text, "+")

	var digits strings.Builder
	for _, r := range text {
		switch {
		case r >= '0' && r <= '9':
			digits.WriteRune(r)
		case r == ' ' || r == '-' || r == '.' || r == '(' || r == ')':
		default:
			return "", ErrInvalidPhone
		}
	}
	number := digits.String()

	// 00 is the other way of writing + in front of a country code.
	if !plus && strings.HasPrefix(number, "00") {
		plus, number = true, number[2:]
	}

	if plus {
		if national, ok := strings.CutPrefix(number, nigeriaCode); ok {
			return nigerian(strings.TrimPrefix(national, "0"))
		}
		if len(number) < minInternationalDigits || len(number) > maxInternationalDigits ||
			strings.HasPrefix(number, "0") {
			return "", ErrInvalidPhone
		}
		return "+" + number, nil
	}

	// Without a + the length tells the forms apart, because a 10-digit local number could
	// itself begin with 234.
	switch len(number) {
	case nigerianNationalDigits:
		return nigerian(number)
	case nigerianNationalDigits + 1:
		if strings.HasPrefix(number, "0") {
			return nigerian(number[1:])
		}
	case len(nigeriaCode) + nigerianNationalDigits:
		if strings.HasPrefix(number, nigeriaCode) {
			return nigerian(number[len(nigeriaCode):])
		}
	case len(nigeriaCode) + nigerianNationalDigits + 1:
		if strings.HasPrefix(number, nigeriaCode+"0") {
			return nigerian(number[len(nigeriaCode)+1:])
		}
	}
	return "", ErrInvalidPhone
}

// nigerian checks the digits after +234 (and after any trunk 0) and writes them in full.
func nigerian(national string) (string, error) {
	if len(national) != nigerianNationalDigits || strings.HasPrefix(national, "0") {
		return "", ErrInvalidPhone
	}
	return "+" + nigeriaCode + national, nil
}
