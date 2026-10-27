package service_test

import (
	"context"
	"database/sql"
	"net/http"
	"testing"

	"pharmacy/internal/constant"
	"pharmacy/internal/testdb"
	"pharmacy/internal/types"
	"pharmacy/model"
	"pharmacy/repository"
	"pharmacy/service"
)

func TestCreateUserAgainstTheDatabase(t *testing.T) {
	db := testdb.New(t, 0)
	store := repository.NewStore(db)
	svc := service.NewAdminService(store)
	ctx := context.Background()

	var cashier, admin int
	if err := db.Get(&cashier, `SELECT id FROM roles WHERE name = 'Cashier'`); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&admin, `SELECT id FROM roles WHERE name = 'Admin'`); err != nil {
		t.Fatal(err)
	}

	// The create-admin command adds the first user with no one signed in (actor 0).
	firstAdmin, err := svc.CreateUser(ctx, types.CreateUserRequest{
		Username: "owner", Password: "first admin pass", RoleID: admin,
	}, 0)
	if err != nil {
		t.Fatalf("first admin: %v", err)
	}

	id, err := svc.CreateUser(ctx, types.CreateUserRequest{
		Username: "ada", Password: "ten chars!", RoleID: cashier,
	}, firstAdmin)
	if err != nil {
		t.Fatalf("cashier: %v", err)
	}

	stored, err := store.FetchUserWithPassword(ctx, "ada")
	if err != nil {
		t.Fatal(err)
	}
	if stored.RoleID != cashier {
		t.Fatalf("expected role %d, got %d", cashier, stored.RoleID)
	}
	if ok, err := (&model.User{Password: "ten chars!"}).VerifyPassword(
		stored.Password,
	); err != nil ||
		!ok {
		t.Fatalf("the stored password does not verify: %v", err)
	}

	var actors []sql.NullInt64
	err = db.Select(&actors, `SELECT actor_id FROM audit_log WHERE entity_type = $1 ORDER BY id`,
		constant.AuditEntityUser)
	if err != nil {
		t.Fatal(err)
	}
	if len(actors) != 2 || actors[0].Valid || !actors[1].Valid ||
		int(actors[1].Int64) != firstAdmin {
		t.Fatalf("expected audit actors [none, %d], got %v", firstAdmin, actors)
	}
	var auditedID int
	db.Get(
		&auditedID,
		`SELECT entity_id FROM audit_log WHERE entity_type = $1 ORDER BY id DESC LIMIT 1`,
		constant.AuditEntityUser,
	)
	if auditedID != id {
		t.Fatalf("audit entry points at user %d, want %d", auditedID, id)
	}

	_, err = svc.CreateUser(ctx, types.CreateUserRequest{
		Username: "ada", Password: "another password", RoleID: cashier,
	}, firstAdmin)
	if status(err) != http.StatusConflict {
		t.Fatalf("a taken username should be a 409, got %v", err)
	}

	var users int
	db.Get(&users, `SELECT count(*) FROM users`)
	if users != 2 {
		t.Fatalf("a duplicate must not add a row, have %d users", users)
	}
}
