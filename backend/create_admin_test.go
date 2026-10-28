package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"pharmacy/internal/types"
	"strings"
	"testing"
)

type fakeAdminCreator struct {
	got     types.CreateUserRequest
	actorID int
	err     error
}

func (f *fakeAdminCreator) CreateUser(
	_ context.Context, req types.CreateUserRequest, actorID int,
) (int, error) {
	f.got, f.actorID = req, actorID
	return 1, f.err
}

type fakeRoles map[string]int

func (f fakeRoles) FetchRoleIDByName(_ context.Context, name string) (int, error) {
	id, ok := f[name]
	if !ok {
		return 0, errors.New("no rows")
	}
	return id, nil
}

func TestCreateAdminUsesTheAdminRoleWithNoActor(t *testing.T) {
	svc := &fakeAdminCreator{}

	_, err := createAdmin(
		context.Background(),
		svc,
		fakeRoles{"Admin": 1},
		"owner",
		"a long secret",
	)
	if err != nil {
		t.Fatal(err)
	}
	want := types.CreateUserRequest{Username: "owner", Password: "a long secret", RoleID: 1}
	if svc.got != want || svc.actorID != 0 {
		t.Fatalf("got %+v from actor %d", svc.got, svc.actorID)
	}
}

func TestCreateAdminNeedsAnAdminRole(t *testing.T) {
	_, err := createAdmin(context.Background(), &fakeAdminCreator{}, fakeRoles{}, "owner", "x")
	if err == nil || !strings.Contains(err.Error(), "Admin role") {
		t.Fatalf("expected a missing Admin role error, got %v", err)
	}
}

func TestReadNewPasswordPrefersTheEnvironment(t *testing.T) {
	got, err := readNewPassword("from env", false, nil, &bytes.Buffer{})
	if err != nil || got != "from env" {
		t.Fatalf("got %q, %v", got, err)
	}
}

func TestReadNewPasswordAsksTwiceOnATerminal(t *testing.T) {
	answers := []string{"first try pass", "first try pass"}
	secret := func() ([]byte, error) {
		next := answers[0]
		answers = answers[1:]
		return []byte(next), nil
	}

	got, err := readNewPassword("", true, secret, &bytes.Buffer{})
	if err != nil || got != "first try pass" {
		t.Fatalf("got %q, %v", got, err)
	}
}

func TestReadNewPasswordRefusesAMismatch(t *testing.T) {
	answers := []string{"one password", "another one"}
	secret := func() ([]byte, error) {
		next := answers[0]
		answers = answers[1:]
		return []byte(next), nil
	}

	_, err := readNewPassword("", true, secret, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "do not match") {
		t.Fatalf("expected a mismatch error, got %v", err)
	}
}

func TestReadNewPasswordNeedsTheEnvironmentWithoutATerminal(t *testing.T) {
	_, err := readNewPassword("", false, nil, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "PHARMACY_ADMIN_PASSWORD") {
		t.Fatalf("without a terminal it should ask for PHARMACY_ADMIN_PASSWORD, got %v", err)
	}
}

func TestReadLineStopsAtTheNewline(t *testing.T) {
	in := strings.NewReader("owner\nsecret pass\n")

	line, err := readLine(in)
	if err != nil || line != "owner" {
		t.Fatalf("got %q, %v", line, err)
	}
	rest, _ := io.ReadAll(in)
	if string(rest) != "secret pass\n" {
		t.Fatalf("reading the username must not consume what follows, left %q", rest)
	}
}

func TestCreateAdminHelpExitsCleanly(t *testing.T) {
	if code := runCreateAdmin([]string{"-h"}); code != 0 {
		t.Fatalf("create-admin -h should exit 0, got %d", code)
	}
}
