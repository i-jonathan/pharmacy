package model

import (
	"crypto/rand"
	"crypto/subtle"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/crypto/argon2"
)

type User struct {
	baseModel
	UserName    string      `json:"username"`
	Password    string      `json:"password"`
	RoleID      int         `json:"role_id" db:"role_id"`
	RoleName    string      `json:"role_name" db:"role_name"`
	Permissions Permissions `json:"permissions" db:"permissions" `
}

type Permission struct {
	baseModel
	Resource string `json:"resource" db:"resource"`
	Action   string `json:"action" db:"action"`
}

type Permissions []Permission

func (p *Permissions) Scan(value any) error {
	if value == nil {
		*p = Permissions{}
		return nil
	}

	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into permissions", value)
	}

	return json.Unmarshal(b, p)
}

func (u *User) HashPassword() error {
	p := passwordConfig{
		time:       3,
		memory:     64 * 1024,
		threads:    2,
		keyLength:  32,
		saltLength: 16,
	}

	salt := make([]byte, p.saltLength)
	if _, err := rand.Read(salt); err != nil {
		return err
	}

	hash := argon2.IDKey([]byte(u.Password), salt, p.time, p.memory, p.threads, p.keyLength)
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	u.Password = fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, p.memory, p.time, p.threads, b64Salt, b64Hash)
	return nil
}

func (u *User) VerifyPassword(hash string) (bool, error) {
	parts := strings.Split(hash, "$")

	if len(parts) != 6 {
		return false, fmt.Errorf("invalid hash")
	}

	var version int
	_, err := fmt.Sscanf(parts[2], "v=%d", &version)
	if err != nil {
		return false, err
	}

	if version != argon2.Version {
		return false, fmt.Errorf("invalid hash version: %d", version)
	}

	passConfig := &passwordConfig{}
	_, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &passConfig.memory, &passConfig.time, &passConfig.threads)
	if err != nil {
		return false, err
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, err
	}

	decodedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, err
	}

	passConfig.keyLength = uint32(len(decodedHash))

	comparisonHash := argon2.IDKey([]byte(u.Password), salt, passConfig.time, passConfig.memory, passConfig.threads, passConfig.keyLength)

	return subtle.ConstantTimeCompare(decodedHash, comparisonHash) == 1, nil
}

// MinPasswordLength is the shortest password accepted when one is set, counted in characters.
const MinPasswordLength = 10

var usernamePattern = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

func (u *User) ValidatePassword() error {
	// Surrounding spaces are kept in the stored password but don't count towards its strength.
	core := strings.TrimSpace(u.Password)
	if utf8.RuneCountInString(core) < MinPasswordLength {
		return fmt.Errorf("password must be at least %d characters", MinPasswordLength)
	}
	lowered := strings.ToLower(core)
	if _, common := commonPasswords[lowered]; common {
		return errors.New("this password is too common; choose another")
	}
	if u.UserName != "" && containsAsWord(lowered, strings.ToLower(u.UserName)) {
		return errors.New("the password must not contain the username")
	}
	return nil
}

func (u *User) ValidateUsername() error {
	if len(u.UserName) < 3 || len(u.UserName) > 15 {
		return errors.New("username must be 3 to 15 characters")
	}

	if !usernamePattern.MatchString(u.UserName) {
		return errors.New("username can only use letters, numbers and _")
	}
	return nil
}

// containsAsWord reports whether word appears in s with no letter right before or after it, so
// "mary" is found in "mary123456" but "ada" is not found in "nevada".
func containsAsWord(s, word string) bool {
	for start := 0; ; {
		i := strings.Index(s[start:], word)
		if i < 0 {
			return false
		}
		i += start
		before, _ := utf8.DecodeLastRuneInString(s[:i])
		after, _ := utf8.DecodeRuneInString(s[i+len(word):])
		if !unicode.IsLetter(before) && !unicode.IsLetter(after) {
			return true
		}
		start = i + 1
	}
}

//go:embed common_passwords.txt
var commonPasswordsFile string

// commonPasswords holds the lowercased entries of common_passwords.txt.
var commonPasswords = loadCommonPasswords(commonPasswordsFile)

func loadCommonPasswords(file string) map[string]struct{} {
	set := make(map[string]struct{})
	for _, line := range strings.Split(file, "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			set[line] = struct{}{}
		}
	}
	return set
}
