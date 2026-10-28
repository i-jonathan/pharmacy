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

// fakeDirectory knows the roles and which users already hold the Admin role.
type fakeDirectory struct {
	fakeRoles
	admins []string
}

func (f fakeDirectory) FetchUsernamesByRole(context.Context, int) ([]string, error) {
	return f.admins, nil
}

func TestCreateAdminUsesTheAdminRoleWithNoActor(t *testing.T) {
	svc := &fakeAdminCreator{}

	_, err := createAdmin(
		context.Background(),
		svc,
		fakeDirectory{fakeRoles: fakeRoles{"Admin": 1}},
		"owner",
		"a long secret",
		false,
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
	_, err := createAdmin(
		context.Background(),
		&fakeAdminCreator{},
		fakeDirectory{},
		"owner",
		"x",
		false,
	)
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

func TestCreateAdminRefusesOnceAnAdminExists(t *testing.T) {
	svc := &fakeAdminCreator{}
	dir := fakeDirectory{fakeRoles: fakeRoles{"Admin": 1}, admins: []string{"owner"}}

	_, err := createAdmin(context.Background(), svc, dir, "second", "a long secret", false)

	if err == nil || !strings.Contains(err.Error(), "an admin already exists (owner)") {
		t.Fatalf("expected a refusal naming the existing admin, got %v", err)
	}
	if strings.Contains(strings.ToLower(err.Error()), "recover") {
		t.Fatalf("the refusal must not advertise the override: %v", err)
	}
	if svc.got.Username != "" {
		t.Fatal("no user should be created")
	}
}

func TestCreateAdminRecoveryAddsAnotherAdminAndSaysSo(t *testing.T) {
	svc := &fakeAdminCreator{}
	dir := fakeDirectory{fakeRoles: fakeRoles{"Admin": 1}, admins: []string{"owner"}}

	if _, err := createAdmin(
		context.Background(),
		svc,
		dir,
		"rescue",
		"a long secret",
		true,
	); err != nil {
		t.Fatal(err)
	}
	if svc.got.Username != "rescue" || !svc.got.Recovery {
		t.Fatalf("expected a recovery admin, got %+v", svc.got)
	}
}

func TestCreateAdminHelpDoesNotMentionRecovery(t *testing.T) {
	var out bytes.Buffer
	createAdminFlags(&out).Usage()
	if strings.Contains(strings.ToLower(out.String()), "recover") ||
		!strings.Contains(out.String(), "-username") {
		t.Fatalf("help should list -username and hide the override, got:\n%s", out.String())
	}
}

func TestCreateAdminAcceptsBothRecoverySpellings(t *testing.T) {
	for _, args := range [][]string{{"--recover"}, {"-r"}} {
		fs := createAdminFlags(&bytes.Buffer{})
		recovering := fs.Lookup("recover")
		if err := fs.Parse(args); err != nil || recovering.Value.String() != "true" {
			t.Fatalf("%v should turn recovery on, got %v (%v)", args, recovering.Value, err)
		}
	}
}
