package model

import (
	"strings"
	"testing"
)

func TestValidatePasswordRejectsCommonPasswords(t *testing.T) {
	for _, password := range []string{
		"basketball", "1q2w3e4r5t", "Password12", "QWERTYUIOP", "basketball ", " qwertyuiop",
	} {
		err := (&User{Password: password}).ValidatePassword()
		if err == nil || !strings.Contains(err.Error(), "too common") {
			t.Errorf("%q is a common password and should be refused, got %v", password, err)
		}
	}
}

func TestValidatePasswordRejectsThePasswordContainingTheUsername(t *testing.T) {
	err := (&User{UserName: "ada_cashier", Password: "Ada_Cashier2025"}).ValidatePassword()
	if err == nil || !strings.Contains(err.Error(), "username") {
		t.Fatalf("a password containing the username should be refused, got %v", err)
	}
}

func TestValidatePasswordAcceptsAnUncommonPassphrase(t *testing.T) {
	err := (&User{UserName: "ada_cashier", Password: "calm river 47"}).ValidatePassword()
	if err != nil {
		t.Fatalf("an uncommon passphrase should be accepted: %v", err)
	}
}

func TestCommonPasswordListIsLoaded(t *testing.T) {
	if len(commonPasswords) < 5000 {
		t.Fatalf("expected the embedded list to hold thousands of entries, got %d",
			len(commonPasswords))
	}
}

func TestValidateUsernameExplainsTheRuleInPlainWords(t *testing.T) {
	cases := map[string]string{
		"ab":           "username must be 3 to 15 characters",
		"ada lovelace": "username can only use letters, numbers and _",
	}
	for name, want := range cases {
		err := (&User{UserName: name}).ValidateUsername()
		if err == nil || err.Error() != want {
			t.Errorf("%q: want %q, got %v", name, want, err)
		}
	}
}

func TestValidatePasswordChecksTheUsernameAsAWord(t *testing.T) {
	refused := map[string]string{
		"mary":  "mary123456",
		"jay":   "jay2026!!xx",
		"amaka": "amaka is my name",
		"ada":   "my name is ADA ok",
	}
	for name, password := range refused {
		err := (&User{UserName: name, Password: password}).ValidatePassword()
		if err == nil || !strings.Contains(err.Error(), "username") {
			t.Errorf("%q with username %q should be refused, got %v", password, name, err)
		}
	}
	allowed := map[string]string{
		"ada": "nevada sunset 2026",
		"jay": "majestic river 9",
	}
	for name, password := range allowed {
		if err := (&User{UserName: name, Password: password}).ValidatePassword(); err != nil {
			t.Errorf("%q only contains %q inside a word and should be allowed: %v",
				password, name, err)
		}
	}
}
