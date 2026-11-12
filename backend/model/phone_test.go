package model

import "testing"

func TestNormalizePhoneAcceptsNigerianNumbersInEveryWrittenForm(t *testing.T) {
	for _, in := range []string{
		"08031234567", "0803 123 4567", "0803-123-4567", " (0803) 123 4567 ",
		"8031234567", "803 123 4567",
		"2348031234567", "234 803 123 4567",
		"+2348031234567", "+234 803 123 4567", "+234 0803 123 4567", "+234-0803-123-4567",
		"002348031234567",
	} {
		got, err := NormalizePhone(in)
		if err != nil || got != "+2348031234567" {
			t.Errorf("NormalizePhone(%q) = %q, %v; want +2348031234567", in, got, err)
		}
	}
}

func TestNormalizePhoneTreatsNothingAsNoPhone(t *testing.T) {
	for _, in := range []string{"", "   "} {
		got, err := NormalizePhone(in)
		if err != nil || got != "" {
			t.Errorf("NormalizePhone(%q) = %q, %v; want empty and no error", in, got, err)
		}
	}
}

func TestNormalizePhoneKeepsForeignNumbersInInternationalForm(t *testing.T) {
	got, err := NormalizePhone("+44 20 7946 0958")
	if err != nil || got != "+442079460958" {
		t.Errorf("got %q, %v; want +442079460958", got, err)
	}
}

func TestNormalizePhoneRefusesNumbersOfTheWrongShape(t *testing.T) {
	for _, in := range []string{
		"0803123456",        // one digit short (10 with the 0)
		"080312345678",      // one digit too many
		"803123456",         // 9 digits
		"28031234567",       // 11 digits not starting with 0
		"23480312345",       // 234 plus too few digits
		"+234803123456",     // 234 plus 9 digits
		"+234080312345678",  // 234 plus 0 plus 11 digits
		"+2340803123456",    // 234 plus 0 plus 9 digits
		"+1234567",          // too short to be international
		"+1234567890123456", // longer than E.164 allows
		"+0442079460958",    // country codes do not start with 0
		"0803 123 456x",     // letters
		"call me",           // not a number
		"++2348031234567",   // stray plus
		"0803+1234567",      // plus in the middle
	} {
		if got, err := NormalizePhone(in); err == nil {
			t.Errorf("NormalizePhone(%q) = %q; want an error", in, got)
		}
	}
}
