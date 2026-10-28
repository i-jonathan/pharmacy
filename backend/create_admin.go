package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"pharmacy/httperror"
	"pharmacy/internal/types"
	"pharmacy/repository"
	"pharmacy/service"
	"strings"

	"golang.org/x/term"
)

// adminRoleName is the role create-admin grants. Roles are matched by name, as the permission
// seeding migrations do (see T-099).
const adminRoleName = "Admin"

type adminCreator interface {
	CreateUser(ctx context.Context, req types.CreateUserRequest, actorID int) (int, error)
}

type roleLookup interface {
	FetchRoleIDByName(ctx context.Context, name string) (int, error)
}

// runCreateAdmin is the `create-admin` command: it adds an Admin user, for example the first
// one on a new install, and returns the process exit code. The password comes from
// PHARMACY_ADMIN_PASSWORD when set, otherwise from a hidden prompt.
func runCreateAdmin(args []string) int {
	flags := flag.NewFlagSet("create-admin", flag.ContinueOnError)
	username := flags.String("username", "", "username for the new admin")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}

	name := strings.TrimSpace(*username)
	if name == "" {
		fmt.Print("Username: ")
		// Unbuffered, so anything typed after the username is left for the password prompt.
		line, err := readLine(os.Stdin)
		if err != nil && line == "" {
			fmt.Fprintln(os.Stderr, "create-admin: no username given")
			return 1
		}
		name = strings.TrimSpace(line)
	}

	stdin := int(os.Stdin.Fd())
	password, err := readNewPassword(
		os.Getenv("PHARMACY_ADMIN_PASSWORD"),
		term.IsTerminal(stdin),
		func() ([]byte, error) { return term.ReadPassword(stdin) },
		os.Stdout,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "create-admin: %v\n", err)
		return 1
	}

	store, err := repository.InitStore()
	if err != nil {
		fmt.Fprintf(os.Stderr, "create-admin: could not connect to the database: %v\n", err)
		return 1
	}
	id, err := createAdmin(
		context.Background(), service.NewAdminService(store), store, name, password,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "create-admin: %s\n", userMessage(err))
		return 1
	}
	fmt.Printf("Created admin %q (id %d). Sign in at /user/login.\n", name, id)
	return 0
}

func createAdmin(
	ctx context.Context, svc adminCreator, roles roleLookup, username, password string,
) (int, error) {
	roleID, err := roles.FetchRoleIDByName(ctx, adminRoleName)
	if err != nil {
		return 0, fmt.Errorf("could not find the %s role: %w", adminRoleName, err)
	}
	// No one is signed in, so the audit entry has no actor.
	return svc.CreateUser(ctx, types.CreateUserRequest{
		Username: username, Password: password, RoleID: roleID,
	}, 0)
}

// readNewPassword takes the password from env when set, otherwise asks twice on the terminal
// without echoing. Without a terminal there is no way to confirm it, so env is required.
func readNewPassword(
	env string, terminal bool, readSecret func() ([]byte, error), out io.Writer,
) (string, error) {
	if env != "" {
		return env, nil
	}
	if !terminal {
		return "", errors.New("no terminal to type the password in: set PHARMACY_ADMIN_PASSWORD")
	}

	fmt.Fprint(out, "Password: ")
	first, err := readSecret()
	fmt.Fprintln(out)
	if err != nil {
		return "", err
	}
	fmt.Fprint(out, "Confirm password: ")
	second, err := readSecret()
	fmt.Fprintln(out)
	if err != nil {
		return "", err
	}
	if string(first) != string(second) {
		return "", errors.New("the passwords do not match")
	}
	return string(first), nil
}

// readLine reads up to the next newline one byte at a time, so nothing after it is consumed.
func readLine(r io.Reader) (string, error) {
	var line []byte
	b := make([]byte, 1)
	for {
		n, err := r.Read(b)
		if n == 1 {
			if b[0] == '\n' {
				return strings.TrimRight(string(line), "\r"), nil
			}
			line = append(line, b[0])
		}
		if err != nil {
			return string(line), err
		}
	}
}

// userMessage is the plain message of a service error, without the wrapped cause.
func userMessage(err error) string {
	var httperr *httperror.HTTPError
	if errors.As(err, &httperr) {
		return httperr.Message
	}
	return err.Error()
}
