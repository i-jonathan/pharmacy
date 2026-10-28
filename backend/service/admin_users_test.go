package service

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"pharmacy/internal/constant"
	"pharmacy/internal/types"
	"pharmacy/model"
	"pharmacy/repository"
	"strings"
	"testing"

	"github.com/jmoiron/sqlx"
)

// fakeUserRepo records what CreateUser asks the repository to do.
type fakeUserRepo struct {
	repository.PharmacyRepository

	roles     map[int]bool
	insertErr error
	inserted  []model.User
	audit     []types.AuditEntry
	commits   int
	rollbacks int
	passwords map[int]string
	usernames map[int]string
}

func (r *fakeUserRepo) BeginTx(context.Context) (*sqlx.Tx, error) { return nil, nil }
func (r *fakeUserRepo) RollbackTx(*sqlx.Tx)                       { r.rollbacks++ }
func (r *fakeUserRepo) CommitTx(*sqlx.Tx) error {
	r.commits++
	return nil
}

func (r *fakeUserRepo) CreateUserTx(_ context.Context, _ *sqlx.Tx, user model.User) (int, error) {
	if r.insertErr != nil {
		return 0, r.insertErr
	}
	if !r.roles[user.RoleID] {
		return 0, repository.ErrInvalidReference
	}
	r.inserted = append(r.inserted, user)
	return 41 + len(r.inserted), nil
}

func (r *fakeUserRepo) InsertAuditLogTx(_ context.Context, _ *sqlx.Tx, e types.AuditEntry) error {
	r.audit = append(r.audit, e)
	return nil
}

func (r *fakeUserRepo) FetchUsernameByID(_ context.Context, userID int) (string, error) {
	name, ok := r.usernames[userID]
	if !ok {
		return "", sql.ErrNoRows
	}
	return name, nil
}

func (r *fakeUserRepo) UpdateUserPassword(_ context.Context, userID int, hashed string) error {
	if r.passwords == nil {
		r.passwords = map[int]string{}
	}
	r.passwords[userID] = hashed
	return nil
}

const cashierRoleID = 3

func newUserRepo() *fakeUserRepo {
	return &fakeUserRepo{
		roles:     map[int]bool{1: true, 2: true, cashierRoleID: true},
		usernames: map[int]string{5: "ada_cashier"},
	}
}

func validUser() types.CreateUserRequest {
	return types.CreateUserRequest{
		Username: "  ada_cashier ", Password: "ten chars!", RoleID: cashierRoleID,
	}
}

func TestCreateUserRejectsInvalidInput(t *testing.T) {
	cases := map[string]func(*types.CreateUserRequest){
		"username too short":          func(u *types.CreateUserRequest) { u.Username = "ab" },
		"username with spaces inside": func(u *types.CreateUserRequest) { u.Username = "ada lovelace" },
		"password of nine characters": func(u *types.CreateUserRequest) { u.Password = "nine char" },
		"nine multibyte characters":   func(u *types.CreateUserRequest) { u.Password = "ééééééééé" },
		"only spaces":                 func(u *types.CreateUserRequest) { u.Password = "          " },
		"no role chosen":              func(u *types.CreateUserRequest) { u.RoleID = 0 },
		"role that does not exist":    func(u *types.CreateUserRequest) { u.RoleID = 99 },
	}
	for name, change := range cases {
		t.Run(name, func(t *testing.T) {
			repo := newUserRepo()
			req := validUser()
			change(&req)

			_, err := NewAdminService(repo).CreateUser(context.Background(), req, 7)

			requireHTTPStatus(t, err, http.StatusBadRequest)
			if len(repo.inserted) != 0 || repo.commits != 0 {
				t.Fatalf("nothing should be saved, got %d inserts and %d commits",
					len(repo.inserted), repo.commits)
			}
		})
	}
}

func TestCreateUserSavesAHashedPasswordAndAudits(t *testing.T) {
	repo := newUserRepo()

	id, err := NewAdminService(repo).CreateUser(context.Background(), validUser(), 7)
	if err != nil {
		t.Fatal(err)
	}

	if id != 42 || repo.commits != 1 || len(repo.inserted) != 1 {
		t.Fatalf("expected one committed insert returning 42, got id %d, %d commits, %d inserts",
			id, repo.commits, len(repo.inserted))
	}
	saved := repo.inserted[0]
	if saved.UserName != "ada_cashier" || saved.RoleID != cashierRoleID {
		t.Fatalf("saved the wrong user: %+v", saved)
	}
	if saved.Password == "ten chars!" {
		t.Fatal("password was stored in plain text")
	}
	ok, err := (&model.User{Password: "ten chars!"}).VerifyPassword(saved.Password)
	if err != nil || !ok {
		t.Fatalf("stored hash does not verify: %v", err)
	}

	if len(repo.audit) != 1 {
		t.Fatalf("expected one audit entry, got %d", len(repo.audit))
	}
	entry := repo.audit[0]
	if entry.EntityType != constant.AuditEntityUser || entry.EntityID != 42 ||
		entry.Action != constant.AuditCreated || entry.ActorID != 7 {
		t.Fatalf("wrong audit entry: %+v", entry)
	}
	details := string(entry.Details)
	if !strings.Contains(details, "ada_cashier") || strings.Contains(details, "ten chars!") ||
		strings.Contains(details, saved.Password) {
		t.Fatalf("audit details must name the user and never the password: %s", details)
	}
}

func TestCreateUserReportsATakenUsernameAsConflict(t *testing.T) {
	repo := newUserRepo()
	repo.insertErr = repository.ErrUsernameTaken

	_, err := NewAdminService(repo).CreateUser(context.Background(), validUser(), 7)

	requireHTTPStatus(t, err, http.StatusConflict)
	if repo.commits != 0 {
		t.Fatal("a failed insert must not commit")
	}
}

func TestCreateUserDoesNotCallADatabaseFailureADuplicate(t *testing.T) {
	repo := newUserRepo()
	repo.insertErr = errors.New("connection reset")

	_, err := NewAdminService(repo).CreateUser(context.Background(), validUser(), 7)

	requireHTTPStatus(t, err, http.StatusInternalServerError)
	if repo.commits != 0 {
		t.Fatal("a failed insert must not commit")
	}
}

func TestResetUserPasswordNeedsTenCharacters(t *testing.T) {
	repo := newUserRepo()
	svc := NewAdminService(repo)

	err := svc.ResetUserPassword(context.Background(), 5, "nine char")
	requireHTTPStatus(t, err, http.StatusBadRequest)

	if err := svc.ResetUserPassword(context.Background(), 5, "ten chars!"); err != nil {
		t.Fatalf("a ten-character password should be accepted: %v", err)
	}
	if repo.passwords[5] == "" || repo.passwords[5] == "ten chars!" {
		t.Fatal("expected a hashed password to be saved")
	}
}

func TestCreateUserFromTheCommandLineHasNoActor(t *testing.T) {
	repo := newUserRepo()

	if _, err := NewAdminService(
		repo,
	).CreateUser(context.Background(), validUser(), 0); err != nil {
		t.Fatal(err)
	}
	if !repo.audit[0].NoActor {
		t.Fatal("a user created with no one signed in should be audited with no actor")
	}

	repo = newUserRepo()
	if _, err := NewAdminService(
		repo,
	).CreateUser(context.Background(), validUser(), 7); err != nil {
		t.Fatal(err)
	}
	if repo.audit[0].NoActor {
		t.Fatal("a user created by an admin must record that admin as the actor")
	}
}

func TestResetUserPasswordRefusesOneContainingTheUsername(t *testing.T) {
	err := NewAdminService(
		newUserRepo(),
	).ResetUserPassword(context.Background(), 5, "Ada_Cashier2025")
	requireHTTPStatus(t, err, http.StatusBadRequest)
}

func TestResetUserPasswordOfAnUnknownUserIsNotFound(t *testing.T) {
	err := NewAdminService(
		newUserRepo(),
	).ResetUserPassword(context.Background(), 99, "calm river 47")
	requireHTTPStatus(t, err, http.StatusNotFound)
}

func TestResetUserPasswordKeepsSpacesSoItCanBeTypedBack(t *testing.T) {
	repo := newUserRepo()

	err := NewAdminService(repo).ResetUserPassword(context.Background(), 5, "  calm river 47 ")
	if err != nil {
		t.Fatal(err)
	}
	ok, err := (&model.User{Password: "  calm river 47 "}).VerifyPassword(repo.passwords[5])
	if err != nil || !ok {
		t.Fatal("the password must be stored exactly as typed, spaces included")
	}
}

func TestCreateUserRefusesANegativeActor(t *testing.T) {
	repo := newUserRepo()

	_, err := NewAdminService(repo).CreateUser(context.Background(), validUser(), -1)

	requireHTTPStatus(t, err, http.StatusBadRequest)
	if len(repo.inserted) != 0 {
		t.Fatal("nothing should be saved for an invalid actor")
	}
}
