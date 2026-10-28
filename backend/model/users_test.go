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
